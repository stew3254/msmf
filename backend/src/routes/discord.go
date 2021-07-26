package routes

import (
	"encoding/json"
	"msmf/database"
	"msmf/utils"
	"net/http"
)

func checkIntPerms(w http.ResponseWriter, r *http.Request) bool {
	// Get the token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// Has sever level perms to make this permission
	hasPerms := utils.CheckPermissions(&utils.PermCheck{
		FKTable:     "server_perms_per_users",
		Perms:       "edit_configuration",
		PermTable:   "server_perms",
		Search:      token,
		SearchCol:   "token",
		SearchTable: "users",
	})

	if !hasPerms {
		// Check user level perms to make this integration
		hasPerms = utils.CheckPermissions(&utils.PermCheck{
			FKTable:     "perms_per_users",
			Perms:       "manage_server_permissions",
			PermTable:   "user_perms",
			Search:      token,
			SearchCol:   "token",
			SearchTable: "users",
		})
		if !hasPerms {
			utils.ErrorJSON(w, http.StatusForbidden, "Forbidden")
			return false
		}
	}
	return true
}

// MakeIntegration will create or update an integration with a server
func MakeIntegration(w http.ResponseWriter, r *http.Request) {
	// Check integration perms
	if !checkIntPerms(w, r) {
		// Didn't meet the required permissions
		return
	}

	// See if this server exists
	serverID := getServer(r.URL.String())
	var server database.Server
	err := database.DB.First(&server, serverID).Error
	if err != nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Server does not exist")
		return
	}

	// Get JSON of body of PUT
	body := make(map[string]string)
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Make sure the type url exists
	integrationType, exists := body["type"]
	if !exists {
		utils.ErrorJSON(w, http.StatusBadRequest, "Must supply an integration type")
		return
	} else if integrationType != "webhook" {
		utils.ErrorJSON(w, http.StatusBadRequest, "Can only support type \"webhook\"")
		return
	}

	// Make sure the discord url exists
	// TODO actually check to see if this is a valid webhook
	discord, exists := body["discord_url"]
	if !exists {
		utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a discord url")
		return
	}

	// Create the integration
	integration := database.DiscordIntegration{
		Type:       integrationType,
		DiscordURL: discord,
		Active:     false,
		Server:     server,
	}

	// Get the username if it exists
	name, exists := body["username"]
	if exists {
		integration.Username = name
	} else {
		// Add server name otherwise
		integration.Username = server.Name
	}

	// Get the avatar url if it exists
	avatar, exists := body["avatar_url"]
	if exists {
		integration.AvatarURL = avatar
	}

	// Actually create the integration
	database.DB.Create(&integration)

	// Write out success
	resp := make(map[string]string)
	resp["status"] = "Success"
	utils.WriteJSON(w, http.StatusOK, &resp)
}

// DeleteIntegration will delete an integration with a server
func DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	// Check integration perms
	if !checkIntPerms(w, r) {
		// Didn't meet the required permissions
		return
	}

	// Get server id
	serverID := getServer(r.URL.String())

	// Delete the integration
	database.DB.Where("server_id = ?", serverID).Delete(&database.DiscordIntegration{})

	// Write out success
	resp := make(map[string]string)
	resp["status"] = "Success"
	utils.WriteJSON(w, http.StatusOK, &resp)
}

// GetIntegration will show information about the integration with a server
func GetIntegration(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// SendWebhook will send over the data to Discord
func SendWebhook() {

}
