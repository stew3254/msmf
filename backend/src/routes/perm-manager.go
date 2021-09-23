package routes

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"msmf/database"
	"msmf/utils"
	"net/http"
)

// getUserPerms is a simple helper function to get user perms for a user
func getUserPerms(db *gorm.DB, userID int) (perms []database.UserPerm, err error) {
	var myPerms []database.UserPerm
	err = db.Joins(
		"INNER JOIN perms_per_users ppu ON ppu.user_perm_id = user_perms.id",
	).Joins(
		"INNER JOIN users u ON u.id = ppu.user_id",
	).Where(
		"u.id = ?",
		userID,
	).Find(&myPerms).Error
	return myPerms, err
}

// GetUserPerms contains all ways to get user level permissions
func GetUserPerms(w http.ResponseWriter, r *http.Request) {
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// Get url parameters
	params := mux.Vars(r)
	username, exists := params["user"]

	// See if the user has permissions to change other people's permissions
	hasPerms := utils.CheckPermissions(&utils.PermCheck{
		FKTable:     "perms_per_users",
		Perms:       "manage_user_permission",
		PermTable:   "user_perms",
		Search:      token,
		SearchCol:   "token",
		SearchTable: "users",
	})

	type Result struct {
		Username    string `json:"username"`
		Permission  string `json:"permission"`
		Description string `json:"description"`
	}

	var result []Result
	var err error

	// Create a query to get user perms
	query := database.DB.Table("perms_per_users ppu").Select(
		"u.username as username, up.name as permission, up.description as description",
	).Joins(
		"INNER JOIN users u ON ppu.user_id = u.id",
	).Joins(
		"INNER JOIN user_perms up ON ppu.user_perm_id = up.id",
	)

	// See if the user has permissions and didn't just query a specific user
	if hasPerms && !exists {
		err = query.Find(&result).Error
		// See if they have perms and they want to find details on a specific user
	} else if hasPerms && exists {
		// Get user
		var user database.User
		database.DB.Where("username = ?", username).Find(&user)

		// They don't exist
		if user.ID == nil {
			utils.ErrorJSON(w, http.StatusNotFound, "Not found")
			return
		}

		// Look for perms for a specific user
		err = query.Where("username = ?", username).Find(&result).Error
		// They don't have perms and they didn't look for a particular person
	} else if !exists {
		// They are just going to get perms for themselves
		err = query.Where("token = ?", token).Find(&result).Error
		// They don't have perms and are trying to find someone
	} else {
		// They aren't allowed to see other users
		utils.ErrorJSON(w, http.StatusNotFound, "Not found")
		return
	}

	// If we had an error, complain
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Write out the perms
	utils.WriteJSON(w, http.StatusOK, &result)
}

// UpdateUserPerms allows changes to a user's user level permission status to be changed
func UpdateUserPerms(w http.ResponseWriter, r *http.Request) {
	// Get myUser
	params := mux.Vars(r)
	username := params["user"]

	// Get token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// Get request body
	var body []string
	err := json.NewDecoder(r.Body).Decode(&body)
	// TODO don't use constant to put a hard limit on how many myUser permissions a person can send
	// This right now is just so database queries aren't slow if they are intentionally trying to
	// blast the database with a useless query
	if err != nil || len(body) > 10 {
		utils.WriteJSON(w, http.StatusBadRequest, "Bad request")
	}

	// Get your user
	var myUser database.User
	err = database.DB.Where("token = ?", token).Find(&myUser).Error
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// See if the username is yourself or admin
	if username == myUser.Username || username == "admin" {
		utils.ErrorJSON(w, http.StatusForbidden, "Cannot update user permissions for this user")
		return
	}

	// Start a transaction
	err = database.DB.Transaction(func(db *gorm.DB) error {
		// Check if you have permissions to modify this user
		hasPerms := utils.CheckPermissions(&utils.PermCheck{
			FKTable:     "perms_per_users",
			Perms:       "manage_user_permission",
			PermTable:   "user_perms",
			Search:      token,
			SearchCol:   "token",
			SearchTable: "users",
		})

		if hasPerms {
			// Get all user permissions for yourself
			var myPerms []database.UserPerm
			myPerms, err = getUserPerms(db, *myUser.ID)

			// Complain on error
			if err != nil {
				return err
			}

			// Create a map to check perms later
			myPermsMap := make(map[string]bool)
			for _, perm := range myPerms {
				myPermsMap[perm.Name] = true
			}

			// Get user permissions from supplied list
			var perms []database.UserPerm
			err = db.Where("name in ?", body).Find(&perms).Error
			if err != nil {
				return err
				// Not all permissions were found
			} else if len(perms) != len(body) {
				return errors.New("at least 1 permission is invalid")
			}

			// If you don't have the administrator permission
			if !myPermsMap["administrator"] {
				// Make sure added perms aren't something you don't already have
				for _, perm := range body {
					if !myPermsMap[perm] {
						return errors.New("cannot assign permission you don't already have yourself")
					}
				}
			}

			// Get this user
			var user database.User
			err = db.Where("username = ?", username).Find(&user).Error

			// Complain on error
			if err != nil {
				return err
			} else if user.ID == nil {
				return errors.New("cannot find this user")
			}

			// Get user's exising permissions
			var userPerms []database.UserPerm
			userPerms, err = getUserPerms(db, *user.ID)

			// Complain on error
			if err != nil {
				return err
			}

			// Add any relevant missing perms so the user doesn't lose things
			// that aren't allowed to be touched
			for _, perm := range userPerms {
				// If the perm doesn't exist and you're not admin, add it
				if !myPermsMap["administrator"] && !myPermsMap[perm.Name] {
					perms = append(perms, perm)
				}
			}

			// Delete all user level perms for this user
			err = db.Table("perms_per_users ppu").Where(
				"ppu.user_id = ?",
				*user.ID,
			).Delete(&database.PermsPerUser{}).Error

			// Complain on error
			if err != nil {
				return err
			}

			// If they were being given perms at all, add them
			if len(body) > 0 {
				// Create the perms per users
				ppu := make([]database.PermsPerUser, 0, len(perms))
				for _, perm := range perms {
					ppu = append(ppu, database.PermsPerUser{
						UserID:     *user.ID,
						UserPermID: *perm.ID,
					})
				}

				// Insert the new perms for this user
				err = db.Create(&ppu).Error
			}
		}

		return err
	})

	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
	}

}

// GetServerPerms contains all ways to get server level permissions
func GetServerPerms(w http.ResponseWriter, r *http.Request) {
	// Define the result type
	type Result struct {
		Owner       string `json:"owner"`
		Name        string `json:"name"`
		Username    string `json:"username"`
		Permission  string `json:"permission"`
		Description string `json:"description"`
	}

	var result []Result
	var err error

	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// Get url parameters
	params := mux.Vars(r)
	owner, serverExists := params["owner"]
	name, serverExists := params["name"]
	username, userExists := params["user"]

	// Create a query to get user perms
	query := database.DB.Table("server_perms_per_users sppu").Select(
		"u.username as username, " +
			"sp.name as permission, " +
			"sp.description as description, " +
			"s.name as name, " +
			"o.username as owner",
	).Joins(
		"INNER JOIN users u ON sppu.user_id = u.id",
	).Joins(
		"INNER JOIN server_perms sp ON sppu.server_perm_id = sp.id",
	).Joins(
		"INNER JOIN servers s ON sppu.server_id = s.id",
	).Joins(
		"INNER JOIN users o ON s.owner_id = o.id",
	)

	// They are looking for all servers
	if !serverExists && !userExists {
		// See if the user has user level permissions to change other people's permissions
		hasPerms := utils.CheckPermissions(&utils.PermCheck{
			FKTable:     "perms_per_users",
			Perms:       "manage_server_permission",
			PermTable:   "user_perms",
			Search:      token,
			SearchCol:   "token",
			SearchTable: "users",
		})

		// Check more fine-grained detail since they can't view all servers
		if !hasPerms {
			// Get owned servers
			ownedServersQuery := database.DB.Select("s.*").Table("servers s").Joins(
				"INNER JOIN users u ON s.owner_id = u.id",
			).Where("u.token = ?", token)

			// Get servers they have perms for
			serverPermsQuery := database.DB.Preload(clause.Associations).Select(
				"servers.id as id",
			).Joins(
				"INNER JOIN server_perms_per_users sppu ON servers.id = sppu.server_id",
			).Joins(
				"INNER JOIN server_perms sp ON sppu.server_perm_id = sp.id",
			).Joins(
				"INNER JOIN users u ON sppu.user_id = u.id",
			).Where(
				"u.token = ?", token,
			)

			// TODO need to test this
			// Join the two queries
			err = database.DB.Distinct(
				"u.username as username, "+
					"sp.name as permission, "+
					"sp.description as description, "+
					"s.name as name, "+
					"o.username as owner",
			).Table("(?) as serv", serverPermsQuery).Joins(
				"FULL OUTER JOIN (?) as other ON serv.id = other.id", ownedServersQuery,
			).Joins(
				"INNER JOIN server_perms_per_users sppu ON serv.id = sppu.server_id",
			).Joins(
				"INNER JOIN server_perms sp ON sppu.perm_id = sp.id",
			).Joins(
				"INNER JOIN users u ON sppu.user_id = u.id",
			).Joins(
				"INNER JOIN users o ON serv.owner_id = o.id",
			).Find(&result).Error

			// Tell them there is nothing
			if err != nil {
				utils.WriteJSON(w, http.StatusOK, &[]Result{})
			} else {
				// Show them the servers
				utils.WriteJSON(w, http.StatusOK, &result)
			}
			return
		}

		// Get all server perms per server
		err = query.Find(&result).Error

	} else if serverExists {
		// See if they can view a sever
		canView, err := canViewServer(owner, name, token)

		// Can't view the server
		if !canView {
			utils.ErrorJSON(w, http.StatusForbidden, "Forbidden")
			return
		} else if err != nil {
			utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Get all server perms for a specific server
		err = query.Where(
			"o.username = ? AND REPLACE(LOWER(s.name), ' ', '-') = ?",
			owner,
			name,
		).Find(&result).Error

	} else if userExists {
		// See if the user is yourself
		var user database.User
		database.DB.Where("users.username = ?", username).Find(&user)

		// User doesn't exist
		if user.ID == nil {
			utils.ErrorJSON(w, http.StatusBadRequest, "User does not exist")
			return
		}

		// This user is yourself
		if user.Token == token {
			// Get all servers you can see
			query.Where("users.token = ?", token).Find(&result)
			utils.WriteJSON(w, http.StatusOK, &result)
		}

		// See if the user has user level permissions to change other people's permissions
		hasPerms := utils.CheckPermissions(&utils.PermCheck{
			FKTable:     "perms_per_users",
			Perms:       "manage_server_permission",
			PermTable:   "user_perms",
			Search:      token,
			SearchCol:   "token",
			SearchTable: "users",
		})

		// Can't view all servers to find this user
		if !hasPerms {
			// Create a query to count all instances you and this user share in a server
			findAll := database.DB.Table("server_perms_per_users sppu").Select(
				"sppu.server_id as id, count(u.id)",
			).Joins(
				"INNER JOIN users u on u.id = sppu.user_id",
			).Where(
				"u.token = ? AND u.username = ?", token, username,
			).Group("sppu.server_id")

			// TODO FINISH THIS
			query.Joins("INNER JOIN (?) s ON s.id = sppu.server_id", findAll).Where(
				"s.count = 2 AND u.username = ?", username,
			).Find(&result)

		}

		// Get all server perms for a specific user
		err = query.Where("u.username = ?", username).Find(&result).Error

	} else {
		// See if they can view the sever
		canView, err := canViewServer(owner, name, token)

		// Can't view the server
		if !canView {
			utils.ErrorJSON(w, http.StatusForbidden, "Forbidden")
			return
		}
		if err != nil {
			utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Get all server perms for a specific server and a specific user
		err = query.Where(
			"o.username = ? AND REPLACE(LOWER(s.name), ' ', '-') = ? AND u.username = ?",
			owner,
			name,
			username,
		).Find(&result).Error
	}

	// Complain on error
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Write out the results and exit
	utils.WriteJSON(w, http.StatusOK, &result)
	return
}

// UpdateServerPerms allows changes to a user's server level permission status to be changed
func UpdateServerPerms(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "501 not implemented", http.StatusNotImplemented)
}
