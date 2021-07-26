package routes

import (
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
	w.Write(utils.ToJSON(resp))
}
