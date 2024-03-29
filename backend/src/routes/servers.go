package routes

import (
	"encoding/json"
	"gorm.io/gorm/clause"
	"msmf/database"
	"msmf/games"
	"msmf/utils"
	"net/http"
	"strconv"
	"strings"
)

// Helper function to check permissions of a user
func checkPerms(w http.ResponseWriter, r *http.Request, perm string, isServerPerm bool) bool {
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

// Gets a server id from the url
func getServer(url string) (serverID int) {
	parts := strings.SplitN(url, "/", 5)
	// We know the server id will always be this part of the url
	serverID, _ = strconv.Atoi(parts[3])
	return serverID
}

// Helper function to see if a person can view a server before doing other permission checking
// Note, this function does not tell you whether a server exists or not explicitly, it just implies
// whether you could see it or not if it existed. You must still do your own manual checks to see
// if it exists
func canViewServer(serverID int, token string) (bool, error) {
	// See if they are the server owner
	var count int64
	database.DB.Table("servers").Joins(
		"INNER JOIN users ON servers.owner_id = users.id",
	).Where("users.token = ? AND servers.id = ?", token, serverID).Count(&count)
	// They are not the owner or server doesn't exist
	if count == 0 {
		// See if this user has any server level permissions to be able to view this server
		err := database.DB.Table("servers s").Joins(
			"INNER JOIN server_perms_per_users sppu ON s.id = sppu.server_id",
		).Joins(
			"INNER JOIN server_perms sp ON sppu.server_perm_id = sp.id",
		).Joins(
			"INNER JOIN users u ON sppu.user_id = u.id",
		).Where(
			"u.token = ? AND s.id = ?", token, serverID,
		).Count(&count).Error

		// Person has no server level perms to view this server
		if err != nil {
			return false, err
		}

		// They don't have a single relevant permission for the server
		if count == 0 {
			// See if this user has any user level permissions to be able to view this server
			err = database.DB.Table("users u").Joins(
				"INNER JOIN perms_per_users ppu ON u.id = ppu.user_id",
			).Joins(
				"INNER JOIN user_perms up ON ppu.user_perm_id = up.id",
			).Where("u.token = ? AND ("+
				"up.name = 'administrator' OR "+
				"up.name = 'manage_server_permission' OR "+
				"up.name = 'delete_server'"+
				")", token).Count(&count).Error
			if err != nil {
				return false, err
			} else if count == 0 {
				// There are no relevant user level permissions
				return false, nil
			}
		}
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
	if action == "start" {
		err = utils.StartServer(utils.GameName(getServer(r.URL.String())))
	} else if action == "stop" {
		err = utils.StopServer(utils.GameName(getServer(r.URL.String())))
	} else {
		// Ignore the first since if there was a problem the second would catch it anyways
		_ = utils.StopServer(utils.GameName(getServer(r.URL.String())))
		err = utils.StartServer(utils.GameName(getServer(r.URL.String())))
	}

	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get server id
	serverID := getServer(r.URL.String())

	// Update the running status in the db
	if action == "stop" {
		database.DB.Model(&database.Server{}).Where(
			"servers.id = ?", serverID,
		).Update("running", false)
	} else {
		database.DB.Model(&database.Server{}).Where(
			"servers.id = ?", serverID,
		).Update("running", true)
	}

	// Write out response
	resp := make(map[string]string)
	resp["status"] = "Success"
	_, _ = w.Write(utils.ToJSON(&resp))
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
	name := body["name"].(string)
	if len(name) == 0 {
		utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a server name")
		return
	}

	// Get rest of form values
	gameName := body["game"].(string)
	versionName := body["version"].(string)

	// See if game exists
	var game database.Game
	err = database.DB.Where("games.name = ?", gameName).First(&game).Error
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a supported game")
		return
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
		if err != nil {
			// Add the version to the db
			// No error checking on Minecraft versions for now.
			// If you add something stupid that breaks things, it's your own fault
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
		if s == utils.GameName(*server.ID) {
			// Delete the existing server already, something went wrong
			// This might not be the right action to do, but will work for now
			utils.DeleteServer(s)
		}
	}

	// Actually create Minecraft server
	utils.CreateServer(*server.ID, image, game.IsImage, parameters)

	// Write out response
	resp := make(map[string]string)
	resp["status"] = "Success"
	_, _ = w.Write(utils.ToJSON(&resp))
}

func GetServers(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func GetServer(w http.ResponseWriter, r *http.Request) {
	// Get user token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value
	// Get server ID
	serverID := getServer(r.URL.String())

	// If error, there was a database error
	viewable, err := canViewServer(serverID, token)
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
	var server database.Server
	database.DB.Preload(clause.Associations).Where("servers.id = ?", serverID).Find(&server)
	if server.ID == nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Server does not exist")
		return
	}

	// Write out the server data
	_, _ = w.Write(utils.ToJSON(&server))
}

func UpdateServer(w http.ResponseWriter, r *http.Request) {
	// Get user token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value
	// Get server ID
	serverID := getServer(r.URL.String())

	// If error, there was a database error
	viewable, err := canViewServer(serverID, token)
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
	var server database.Server
	database.DB.Preload(clause.Associations).Where("servers.id = ?", serverID).Find(&server)
	if server.ID == nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Server does not exist")
		return
	}

	// Get JSON of body of PATCH
	body := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update the server with requested fields
	database.DB.Model(&server).Updates(body)

	// User sent something bad
	if database.DB.Error != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, database.DB.Error.Error())
		return
	}

	// Write out the new updated server data
	_, _ = w.Write(utils.ToJSON(&server))
}

func DeleteServer(w http.ResponseWriter, r *http.Request) {
	// Get user token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value
	// Get server ID
	serverID := getServer(r.URL.String())

	// See if they are the server owner
	var count int64
	database.DB.Joins(
		"INNER JOIN users ON servers.owner_id = users.id",
	).Where("users.token = ? AND servers.id = ?", token, serverID).Count(&count)
	// They are not the owner
	if count == 0 {
		// Check perms and bail if the perms aren't good
		if !checkPerms(w, r, "delete_server", false) {
			return
		}
	}

	// Delete the server
	utils.DeleteServer(utils.GameName(getServer(r.URL.String())))

	// Delete it from the database
	database.DB.Delete(&database.Server{}, serverID)

	// Write out response
	resp := make(map[string]string)
	resp["status"] = "Success"
	_, _ = w.Write(utils.ToJSON(&resp))
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
