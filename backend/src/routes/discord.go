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
	server, err := getServer(r, false)
	if err != nil || server.ID == nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Server does not exist")
	}

	// Get JSON of body of PUT
	body := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Make sure integration type exists
	var integrationType string
	switch body["type"].(type) {
	case string:
		integrationType = body["type"].(string)
		if integrationType != "webhook" {
			utils.ErrorJSON(w, http.StatusBadRequest, "Can only support type \"webhook\"")
			return
		}
	default:
		utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a valid integration type")
		return
	}

	// Make sure the discord url exists
	// TODO actually check to see if this is a valid webhook
	var discord string
	switch body["discord_url"].(type) {
	case string:
		discord = body["discord_url"].(string)
	default:
		utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a valid url")
		return
	}

	var integration database.DiscordIntegration
	// Try to see if integration exists
	database.DB.Where("server_id = ?", *server.ID).Find(&integration)

	// If the integration doesn't really exist
	if integration.ServerID == nil {
		// Create the integration
		integration = database.DiscordIntegration{
			Type:       integrationType,
			DiscordURL: discord,
			Username:   &server.Name,
			Active:     false,
			Server:     *server,
		}
		// Create an integration to start
		database.DB.Create(&integration)
	} else {
		// Make the changes
		integration.Type = integrationType
		integration.DiscordURL = discord
	}

	// See if the integration should not be active
	active := false
	value, exists := body["active"]
	if exists {
		switch value.(type) {
		case bool:
			active = value.(bool)
		default:
			utils.ErrorJSON(w, http.StatusBadRequest, "Active must be of type bool")
			return
		}
	}

	if integration.Active == false && active == true {
		// Integration wasn't started before, but should be now, so start it
		connDetails, _ := utils.AttachServer(*server.ID, nil)
		database.DB.Table("discord_integrations").Where(
			"server_id = ?", *server.ID,
		).Update("active", true)
		utils.RunDiscordIntegration(connDetails, *server.ID)
	} else if integration.Active == true && active == false {
		// Integration was started before, but shouldn't be now, so stop it
		utils.StopDiscordIntegration(*server.ID)
	}

	// Set the integration to the new active value
	integration.Active = active

	// Get the username if it exists
	var name string
	value, exists = body["username"]
	if exists {
		switch value.(type) {
		// Add the integration username if it exists
		case string:
			name = value.(string)
			integration.Username = &name
		default:
		}
	}

	// Get the avatar url if it exists
	var avatar string
	value, exists = body["avatar_url"]
	if exists {
		switch value.(type) {
		// Add the integration username if it exists
		case string:
			avatar = value.(string)
			integration.AvatarURL = &avatar
		default:
		}
	}

	// Save the integration
	database.DB.Save(&integration)

	http.Error(w, "", http.StatusNoContent)
}

// DeleteIntegration will delete an integration with a server
func DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	// Check integration perms
	if !checkIntPerms(w, r) {
		// Didn't meet the required permissions
		return
	}

	// Get server id
	server, err := getServer(r, false)
	if err != nil || server.ID == nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Server does not exist")
	}

	// Delete the integration
	database.DB.Where("server_id = ?", *server.ID).Delete(&database.DiscordIntegration{})

	http.Error(w, "", http.StatusNoContent)
}

// GetIntegration will show information about the integration with a server
func GetIntegration(w http.ResponseWriter, r *http.Request) {
	// Check integration perms
	if !checkIntPerms(w, r) {
		// Didn't meet the required permissions
		return
	}

	// Get server id
	server, err := getServer(r, false)
	if err != nil || server.ID == nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Server does not exist")
	}

	var integration database.DiscordIntegration
	err = database.DB.Where("server_id = ?", *server.ID).Find(&integration).Error
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
	}

	utils.WriteJSON(w, http.StatusOK, &integration)
}
