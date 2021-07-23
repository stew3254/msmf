package routes

import (
	"github.com/gorilla/websocket"
	"log"
	"msmf/database"
	"net/http"
	"strconv"
	"strings"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

// WSHandler accepts incoming connections // Middleware already checks for authentication before this is called
func WSHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	// Can't error due to regex checking on route
	serverID, _ := strconv.Atoi(parts[len(parts)-1])

	// Get user token
	tokenCookie, err := r.Cookie("token")
	// This shouldn't happen
	if err != nil {
		log.Println(err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	token := tokenCookie.Value
	var count int64

	// See if this user has any permissions to be able to view this server console
	err = database.DB.Table("servers").Joins(
		"INNER JOIN server_perms_per_users sp ON servers.id = sp.server_id",
	).Joins(
		"INNER JOIN server_perms p ON sp.server_perm_id = p.id",
	).Joins(
		"INNER JOIN users ON sp.user_id = users.id",
	).Where(
		"users.token = ? AND servers.id = ? AND (p.name = 'administrator' OR p."+
			"name = 'view_logs' OR p.name = 'manage_server_console')", token, serverID,
	).Count(&count).Error

	// User does not have permissions to view this server
	if err != nil || count == 0 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Upgrade the http connection to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Tries to read messages forever
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// Prints and sends a message back
		log.Println(string(p))
		err = conn.WriteMessage(messageType, []byte("Message from the server!"))

		if err != nil {
			log.Println(err)
			return
		}
	}
}
