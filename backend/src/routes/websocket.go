package routes

import (
	"github.com/gorilla/websocket"
	"log"
	"msmf/database"
	"msmf/utils"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Specify amount of data that can be read from a websocket at a time
var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

// WsServerHandler accepts incoming connections
func WsServerHandler(w http.ResponseWriter, r *http.Request) {
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

	// Owner does not have permissions to view this server
	if err != nil || count == 0 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// See if server console has already been attached
	var conn *websocket.Conn
	utils.ServerLock.Lock()
	connDetails, exists := utils.AttachedServers[serverID]
	// If we haven't already attached a server
	if !exists {
		// Attach the server console
		console, err := utils.AttachServer(utils.ServerName(serverID))
		if err != nil {
			// Don't forget to unlock
			utils.ServerLock.Unlock()

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Create the ConnChan struct
		connDetails = &utils.ConnDetails{
			MChan:   make(chan []byte, 5), // Take up to 5 messages before blocking
			SPMC:    make(map[*websocket.Conn]utils.PipeChans),
			SLock:   &sync.Mutex{},
			ErrChan: make(chan error, 1),
			Pipes:   console,
		}

		// Add it into the map
		utils.AttachedServers[serverID] = connDetails

		// Remember to unlock
		utils.ServerLock.Unlock()

		// Upgrade the http connection to a websocket
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set up the framework for now handling the attachment of the server pipes to go channels
		utils.ServerConsole(connDetails, console)
	} else {
		// Remember to unlock
		utils.ServerLock.Unlock()

		// Upgrade the http connection to a websocket
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Make the channels for stdin and stdout
	pipes := utils.PipeChans{
		StdoutChan: make(chan []byte, 5),
		StderrChan: make(chan []byte, 5),
	}

	// Register with the SPMC
	connDetails.SLock.Lock()
	connDetails.SPMC[conn] = pipes
	connDetails.SLock.Unlock()

	// Now that the server is attached and handlers are running, link websocket

	writeToSocket := func(conn *websocket.Conn, pipes utils.PipeChans) {
		var err error
		// Constantly read messages
		for {
			select {
			// If stdout send as message type 1
			case data := <-pipes.StdoutChan:
				err = conn.WriteMessage(websocket.TextMessage, data)
			// If stderr send as message type 2
			case data := <-pipes.StderrChan:
				err = conn.WriteMessage(websocket.BinaryMessage, data)
			}
			// Reader already takes care of closing the websocket
			if err != nil {
				log.Println("websocket err:", err)
				// Kill this function
				return
			}
		}
	}

	go writeToSocket(conn, pipes)
	// Read in data from the websocket to send to stdin
	readFromSocket := func(conn *websocket.Conn, connDetails *utils.ConnDetails) {
		// Forever try to read in messages
		for {
			messageType, data, err := conn.ReadMessage()
			// Make sure to add the newline character
			data = append(data, byte('\n'))
			if err != nil {
				log.Println("websocket err:", err)
				// Lock to remove this connection from the SPMC and clean up
				connDetails.SLock.Lock()
				delete(connDetails.SPMC, conn)
				connDetails.SLock.Unlock()
				// Best effort close the connection since something is wrong
				_ = conn.Close()
				// Kill this function
				return
			}

			if messageType == websocket.TextMessage {
				// Before sending to stdin, tell all other open websockets you are sending this message
				// This is important so everyone gets to see the same console state
				connDetails.SLock.Lock()
				for _, pipes := range connDetails.SPMC {
					pipes.StdoutChan <- data
				}
				// Now actually send data over to stdin
				connDetails.MChan <- data
				connDetails.SLock.Unlock()
			}
		}
	}

	readFromSocket(conn, connDetails)
}
