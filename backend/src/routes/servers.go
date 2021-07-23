package routes

import (
	"encoding/json"
	"fmt"
	"msmf/database"
	"msmf/games"
	"msmf/utils"
	"net/http"
)

func CreateServer(w http.ResponseWriter, r *http.Request) {
	// Get user token
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	token := tokenCookie.Value

	// See if user has permission to create servers
	hasPerms := utils.CheckPermissions(&utils.PermCheck{
		FKTable:     "perms_per_users",
		Perms:       "create_server",
		PermTable:   "user_perms",
		Search:      token,
		SearchCol:   "token",
		SearchTable: "users",
	})
	if !hasPerms {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// TODO write better json decoder
	// Get put data
	body := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&body)
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
	err = database.DB.Table("servers").Where("servers.user_id = ? AND servers.name = ?", user.ID,
		name).Count(&count).Error
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
		User:    user,
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
		if s == fmt.Sprintf("msmf_server_%d", *server.ID) {
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
	w.Write(utils.ToJSON(&resp))
}

func GetServers(w http.ResponseWriter, r *http.Request) {

}

func ServerHandler(w http.ResponseWriter, r *http.Request) {

}

func UpdateServer(w http.ResponseWriter, r *http.Request) {

}

func DeleteServer(w http.ResponseWriter, r *http.Request) {

}
