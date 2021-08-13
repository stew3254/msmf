package routes

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gorm.io/gorm/clause"
	"log"
	"msmf/database"
	"msmf/games"
	"msmf/utils"
	"net/http"
	"strings"
)

// Helper function to check permissions of a user
func checkPerms(w http.ResponseWriter, r *http.Request, perm string, isServerPerm bool) bool {
	// TODO properly check if they are owner of the server

	// Get user token
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}
	token := tokenCookie.Value

	// See if user has permission to create servers
	var hasPerms bool
	if isServerPerm {
		hasPerms = utils.CheckPermissions(&utils.PermCheck{
			FKTable:     "server_perms_per_users",
			Perms:       perm,
			PermTable:   "server_perms",
			Search:      token,
			SearchCol:   "token",
			SearchTable: "users",
		})
	} else {
		hasPerms = utils.CheckPermissions(&utils.PermCheck{
			FKTable:     "perms_per_users",
			Perms:       perm,
			PermTable:   "user_perms",
			Search:      token,
			SearchCol:   "token",
			SearchTable: "users",
		})
	}
	if !hasPerms {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	return true
}

// getServerURL gets the owner and name from the url
func getServerURL(r *http.Request) (ownerName, serverName string) {
	params := mux.Vars(r)
	return params["owner"], params["name"]
}

// getServer gets the server object from the owner and servername supplied in the url
func getServer(r *http.Request, associations bool) (*database.Server, error) {
	var server database.Server
	var err error

	params := mux.Vars(r)
	ownerName, exists := params["owner"]
	if !exists {
		return nil, errors.New("cannot find server")
	}
	serverName, exists := params["name"]
	if !exists {
		return nil, errors.New("cannot find server")
	}

	if associations {
		err = database.DB.Preload(clause.Associations).Joins(
			"INNER JOIN users u ON u.id = servers.owner_id",
		).Where(
			"REPLACE(LOWER(servers.name), ' ', '-') = ? AND u.username = ?", serverName, ownerName,
		).Find(&server).Error
	} else {
		err = database.DB.Joins(
			"INNER JOIN users u ON u.id = servers.owner_id",
		).Where(
			"REPLACE(LOWER(servers.name), ' ', '-') = ? AND u.username = ?", serverName, ownerName,
		).Find(&server).Error
	}

	return &server, err
}

// Helper function to see if a person can view a server before doing other permission checking
// Note, this function does not tell you whether a server exists or not explicitly, it just implies
// whether you could see it or not if it existed. You must still do your own manual checks to see
// if it exists
func canViewServer(ownerName, serverName, token string) (bool, error) {
	// See if they are the server owner
	var count int64
	// See if this user has any user level permissions to be able to view this server
	err := database.DB.Table("users u").Joins(
		"INNER JOIN perms_per_users ppu ON u.id = ppu.user_id",
	).Joins(
		"INNER JOIN user_perms up ON ppu.user_perm_id = up.id",
	).Where("u.token = ? AND ("+
		"up.name = 'administrator' OR "+
		"up.name = 'manage_server_permission' OR "+
		"up.name = 'delete_server'"+
		")", token).Count(&count).Error

	// Complain on db error
	if err != nil {
		return false, err
	}

	// They have perms to view all servers as existing
	if count > 0 {
		return true, nil
	}

	// See if they are an owner
	ownedServersQuery := database.DB.Select("s.*").Table("servers s").Joins(
		"INNER JOIN users u ON s.owner_id = u.id",
	).Where(
		"u.token = ? AND u.username = ? AND REPLACE(LOWER(s.name), ' ', '-') = ?",
		token, ownerName, serverName,
	)

	// See if they have any related user level perms
	serverPermsQuery := database.DB.Table("servers s").Joins(
		"INNER JOIN server_perms_per_users sppu ON s.id = sppu.server_id",
	).Joins(
		"INNER JOIN server_perms sp ON sppu.server_perm_id = sp.id",
	).Joins(
		"INNER JOIN users u ON sppu.user_id = u.id",
	).Where(
		"u.token = ? AND u.username AND REPLACE(LOWER(s.name), ' ', '-') = ?",
		token, ownerName, serverName,
	)

	// Join the two queries
	err = database.DB.Distinct("sp.*").Table("(?) as sp", serverPermsQuery).Joins(
		"FULL OUTER JOIN (?) as o ON sp.id = o.id", ownedServersQuery,
	).Count(&count).Error

	// Some sort of db error
	if err != nil {
		return false, err
	}

	// User is not an owner or has no perms to view this server
	if count == 0 {
		return false, nil
	}

	return true, nil
}

// Helper function to change running state
func runServer(w http.ResponseWriter, r *http.Request, action string) {
	// Check perms and bail if the perms aren't good
	if !checkPerms(w, r, "restart", true) {
		return
	}
	var err error

	// Get the server
	server, err := getServer(r, false)
	if err != nil || server.ID == nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Not Found")
	}

	if action == "start" {
		err = utils.StartServer(*server.ID, true)
	} else if action == "stop" {
		err = utils.StopServer(*server.ID, true)
	} else {
		// Ignore the first since if there was a problem the second would catch it anyways
		lock := utils.GetLock(*server.ID)
		lock.Lock()
		_ = utils.StopServer(*server.ID, false)
		err = utils.StartServer(*server.ID, false)
		lock.Unlock()
	}

	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update the running status in the db
	if action == "stop" {
		server.Running = false
		database.DB.Save(server)
	} else {
		server.Running = true
		database.DB.Save(server)
	}

	http.Error(w, "", http.StatusNoContent)
}

func CreateServer(w http.ResponseWriter, r *http.Request) {
	// Check perms and bail if the perms aren't good
	if !checkPerms(w, r, "create_server", false) {
		return
	}

	// Get user token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// TODO write better json decoder
	// Get put data
	body := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get port
	var port uint16
	switch p := body["port"].(type) {
	case float64:
		port = uint16(p)
	default:
		utils.ErrorJSON(w, http.StatusBadRequest, "Port must be a valid number")
		return
	}

	// Get name
	var name string
	switch v := body["name"].(type) {
	case string:
		name = v
		// Can't be empty or invalid pattern
		if len(name) == 0 || !utils.ServerPattern.MatchString(name) {
			utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a valid server name")
			return
		}
	default:
		utils.ErrorJSON(w, http.StatusBadRequest, "Name must valid")
		return
	}

	// Get game name
	var gameName string
	var game database.Game
	switch v := body["game"].(type) {
	case string:
		gameName = v
		// See if game exists
		err = database.DB.Where("games.name = ?", gameName).First(&game).Error
		if err != nil {
			utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a supported game")
			return
		}
	default:
		utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a supported game")
		return
	}

	// Get version name
	var versionName string
	switch v := body["version"].(type) {
	case string:
		versionName = v
	}

	// See if port is already taken
	var count int64
	database.DB.Table("servers").Where("servers.port = ?", port).Count(&count)
	if count > 0 {
		utils.ErrorJSON(w, http.StatusBadRequest, "This port has already been allocated")
		return
	}

	// Get user
	var user database.User
	// Shouldn't error since we already confirmed it worked earlier
	database.DB.Where("users.token = ?", token).First(&user)

	// See if name has already been used by this user before
	err = database.DB.Table("servers").Where(
		"servers.owner_id = ? AND servers.name = ?",
		user.ID, name).Count(&count).Error
	if count > 0 {
		utils.ErrorJSON(w, http.StatusBadRequest, "Refuse to add server with same name")
		return
	} else if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// See if version exists
	var version database.Version
	if len(versionName) > 0 {
		if gameName == "Minecraft" && !games.MCIsVersion(versionName) {
			utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a valid Minecraft version")
		}

		err = database.DB.Joins("INNER JOIN games ON games.id = versions.game_id").Where(
			"versions.tag = ? AND games.id = ?",
			versionName,
			game.ID,
		).First(&version).Error
		log.Println(err)
		if err != nil {
			// Add the version to the db
			version = database.Version{
				Tag:  versionName,
				Game: game,
			}
			err = database.DB.Create(&version).Error
			if err != nil {
				// Write the error message out
				utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
	}

	// Create the new server in the db
	server := database.Server{
		Port:    port,
		Name:    name,
		Game:    game,
		Owner:   user,
		Version: version,
	}
	database.DB.Create(&server)

	// Get administrator permission
	var admin database.ServerPerm
	database.DB.Where("server_perms.name = 'administrator'").Find(&admin)

	// Make the user the administrator of the server
	database.DB.Create(&database.ServerPermsPerUser{
		Server:     server,
		ServerPerm: admin,
		User:       user,
	})

	image := game.Image
	// Get parameters
	parameters := games.MakeParameters(body, &image)

	// See if server already exists
	servers := utils.GetGameServers()
	for _, s := range servers {
		if s == utils.ServerName(*server.ID) {
			// Delete the existing server already, something went wrong
			// This might not be the right action to do, but will work for now
			utils.DeleteServer(*server.ID)
		}
	}

	// Actually create Minecraft server
	utils.CreateServer(*server.ID, image, game.IsImage, parameters)

	http.Error(w, "", http.StatusNoContent)
}

func GetServers(w http.ResponseWriter, r *http.Request) {
	// Define the servers
	var servers []database.Server

	// Get user token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// Get query information
	query := r.URL.Query()

	order := "name"

	// Set column ordering
	if len(query.Get("order_by")) > 0 {
		// Only take the first field if they try to put in whitespace
		order = strings.ToLower(strings.Fields(query.Get("order_by"))[0])
	}

	// Set direction
	reverseStr := strings.ToLower(query.Get("reverse"))
	if reverseStr == "true" {
		order += " desc"
	}

	// See if they are the server owner
	var count int64
	// See if this user has any user level permissions to be able to view this server
	err := database.DB.Table("users u").Joins(
		"INNER JOIN perms_per_users ppu ON u.id = ppu.user_id",
	).Joins(
		"INNER JOIN user_perms up ON ppu.user_perm_id = up.id",
	).Where("u.token = ? AND ("+
		"up.name = 'administrator' OR "+
		"up.name = 'manage_server_permission' OR "+
		"up.name = 'delete_server'"+
		")", token).Count(&count).Error

	// Complain on db error
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
	}

	// They have perms to view all servers
	if count > 0 {
		database.DB.Preload(clause.Associations).Order(order).Find(&servers)
	} else {
		// Get owned servers
		ownedServersQuery := database.DB.Select("s.*").Table("servers s").Joins(
			"INNER JOIN users u ON s.owner_id = u.id",
		).Where("u.token = ?", token)

		// Get servers they have perms for
		serverPermsQuery := database.DB.Preload(clause.Associations).Joins(
			"INNER JOIN server_perms_per_users sppu ON servers.id = sppu.server_id",
		).Joins(
			"INNER JOIN server_perms sp ON sppu.server_perm_id = sp.id",
		).Joins(
			"INNER JOIN users u ON sppu.user_id = u.id",
		).Where(
			"u.token = ?", token,
		)

		// Join the two queries
		err = database.DB.Distinct("sp.*").Table("(?) as sp", serverPermsQuery).Joins(
			"FULL OUTER JOIN (?) as o ON sp.id = o.id", ownedServersQuery,
		).Order(order).Find(&servers).Error

		// Some sort of db error
		if err != nil {
			utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Write out the servers
	utils.WriteJSON(w, http.StatusOK, &servers)
}

func GetServer(w http.ResponseWriter, r *http.Request) {
	// Get user token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value
	// Get server ID
	ownerName, serverName := getServerURL(r)

	// If error, there was a database error
	viewable, err := canViewServer(ownerName, serverName, token)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If they can't view it, tell them it's not found
	if !viewable {
		utils.ErrorJSON(w, http.StatusNotFound, "Server does not exist")
		return
	}

	server, err := getServer(r, true)
	if err != nil || server.ID == nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Server does not exist")
		return
	}

	// Write out the server data
	utils.WriteJSON(w, http.StatusOK, &server)
}

func UpdateServer(w http.ResponseWriter, r *http.Request) {
	// Get user token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value
	// Get server params
	ownerName, serverName := getServerURL(r)

	// If error, there was a database error
	viewable, err := canViewServer(ownerName, serverName, token)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If they can't view it, tell them it's not found
	if !viewable {
		utils.ErrorJSON(w, http.StatusNotFound, "Server does not exist")
		return
	}

	// Get the server to see if it actually exists
	server, err := getServer(r, true)
	log.Println(err)
	if err != nil || server.ID == nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Server does not exist")
		return
	}

	// See if they actually have permissions to edit the configuration
	if !checkPerms(w, r, "edit_configuration", true) {
		return
	}

	// Get JSON of body of PATCH
	body := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Delete any nasty keys that shouldn't be in there
	delete(body, "id")
	delete(body, "running")
	delete(body, "owner_id")
	delete(body, "game_id")
	delete(body, "version_id")

	// See if body tries to update version
	value, exists := body["version_tag"]
	var versionTag string
	if exists {
		switch value.(type) {
		case string:
			versionTag = value.(string)
			// Get the version or add it to the db if it doesn't exist
			var version database.Version
			err = database.DB.Where("tag = ?", versionTag).Find(&version).Error
			if err != nil {
				utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
			}

			// Version doesn't exists
			if version.ID == nil {
				version.Tag = versionTag
				version.GameID = server.GameID

				// Add it to the database
				database.DB.Create(&version)

				// Add the version id back into the body map
				body["version_id"] = version.ID

				// TODO add special checks for Minecraft when the version upgrades to something incompatible
				// with the current container
			}
		default:
			delete(body, "version_tag")
		}
	}

	// Update the server with requested fields
	database.DB.Model(&server).Updates(body)

	// User sent something bad
	if database.DB.Error != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, database.DB.Error.Error())
		return
	}

	// Save the updates
	database.DB.Save(&server)

	// Write out the new updated server data
	utils.WriteJSON(w, http.StatusOK, &server)
}

func DeleteServer(w http.ResponseWriter, r *http.Request) {
	// Get user token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value
	// Get server ID
	ownerName, serverName := getServerURL(r)

	// See if they are the server owner
	var count int64
	database.DB.Joins(
		"INNER JOIN users ON servers.owner_id = users.id",
	).Where(
		"users.token = ? AND users.username = ? AND REPLACE(LOWER(servers.name, ' ', '-') = ?",
		token,
		ownerName,
		serverName,
	).Count(&count)
	// They are not the owner
	if count == 0 {
		// Check perms and bail if the perms aren't good
		if !checkPerms(w, r, "delete_server", false) {
			return
		}
	}

	// Delete the server
	server, err := getServer(r, false)
	if err == nil && server.ID != nil {
		utils.DeleteServer(*server.ID)

		// Delete it from the database
		database.DB.Delete(server)
	}

	http.Error(w, "", http.StatusNoContent)
}

// StartServer starts the server
func StartServer(w http.ResponseWriter, r *http.Request) {
	runServer(w, r, "start")
}

// StopServer stops the server
func StopServer(w http.ResponseWriter, r *http.Request) {
	runServer(w, r, "stop")
}

// RestartServer restarts the server
func RestartServer(w http.ResponseWriter, r *http.Request) {
	runServer(w, r, "restart")
}
