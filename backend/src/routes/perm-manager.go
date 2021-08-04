package routes

import (
	"github.com/gorilla/mux"
	"msmf/database"
	"msmf/utils"
	"net/http"
	"strconv"
)

// GetPerms will see which permissions a user has in the webserver
// If the user supplies a serverID, it will show which users have permissions
// they have for said server. Given both, it will query server permissions only for that user
// TODO finish this function
func GetPerms(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})
	query := r.URL.Query()
	username := query.Get("username")
	serverID, err := strconv.Atoi(query.Get("server_id"))
	// See if a server id was actually supplied, and if it's bad then error
	if len(query.Get("server_id")) > 0 && err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Server id must be an integer value")
		return
	}

	// See if the user exists
	if len(username) > 0 {
		// Base permissions struct
		type Permission struct {
			Name        string
			Description string
		}
		var permissions []Permission

		// We aren't looking for server specific permissions
		if serverID == 0 {
			database.DB.Table("user_perms up").Select("up.name, up.description").Joins(
				"INNER JOIN perms_per_users ppu ON ppu.user_perm_id = up.id",
			).Joins(
				"INNER JOIN users u ON u.id = ppu.user_id",
			).Where("u.username = ?", username).Scan(&permissions)
		} else {
			database.DB.Table("server_perms sp").Select("sp.name, sp.description").Joins(
				"INNER JOIN server_perms_per_users ppu ON ppu.server_perm_id = sp.id",
			).Joins(
				"INNER JOIN users u ON u.id = ppu.user_id",
			).Where("u.username = ? AND ppu.server_id = ?", username, serverID).Scan(&permissions)
		}
		if permissions == nil {
			permissions = []Permission{}
		}
		resp["permissions"] = permissions
	}

	// Write out the response
	utils.WriteJSON(w, http.StatusOK, &resp)
}

// GetUserPerms contains all ways to get user level permissions
func GetUserPerms(w http.ResponseWriter, r *http.Request) {
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// Get url parameters
	params := mux.Vars(r)
	name, exists := params["name"]

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
		database.DB.Where("username = ?", name).Find(&user)

		// They don't exist
		if user.ID == nil {
			utils.ErrorJSON(w, http.StatusNotFound, "Not found")
			return
		}

		// Look for perms for a specific user
		err = query.Where("username = ?", name).Find(&result).Error
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
	http.Error(w, "501 not implemented", http.StatusNotImplemented)
}

// GetServerPerms contains all ways to get server level permissions
func GetServerPerms(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "501 not implemented", http.StatusNotImplemented)
}

// UpdateServerPerms allows changes to a user's server level permission status to be changed
func UpdateServerPerms(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "501 not implemented", http.StatusNotImplemented)
}
