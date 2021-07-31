package routes

import (
	"github.com/gorilla/websocket"
	"log"
	"msmf/database"
	"msmf/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
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

	// Get the connection details
	connDetails, _ := utils.AttachServer(serverID, nil)

	// Upgrade the http connection to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Make the channels for stdin and stdout
	pipes := utils.PipeChans{
		StdoutChan: make(chan []byte, 5),
		StderrChan: make(chan []byte, 5),
	}

	connDetails.SLock.Lock()
	// Register with the SPMC
	connDetails.SPMC[conn] = pipes
	// Backfill history on connect
	history := connDetails.History
	for i := 0; i < history.Len(); i += 1 {
		// Get the item from the buffer
		item := history.Buffer[(i+history.Index)%history.Cap()]

		// This loses whether the message was an error message or not
		err := conn.WriteMessage(websocket.TextMessage, item)
		if err != nil {
			// Close the connection
			_ = conn.Close()
		}
	}
	connDetails.SLock.Unlock()

	// Set the ping handler to handle adding no-repeat as well
	conn.SetPingHandler(func(data string) error {
		if data == "no-repeat" {
			// Add the connection to the no repeat map
			connDetails.SLock.Lock()
			connDetails.NoRepeat[conn] = pipes
			connDetails.SLock.Unlock()
		} else if data == "repeat" {
			// Remove the connection from the no repeat map
			connDetails.SLock.Lock()
			delete(connDetails.NoRepeat, conn)
			connDetails.SLock.Unlock()
		} else {
			// Send back pong message
			err = conn.WriteControl(websocket.PongMessage, []byte("Pong!"),
				time.Now().Add(5*time.Minute))
			if err != nil {
				return err
			}
		}
		return nil
	})

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
				// Lock to remove this connection from the SPMC and No Repeat and clean up
				connDetails.SLock.Lock()
				delete(connDetails.SPMC, conn)
				delete(connDetails.NoRepeat, conn)
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
				// Get History
				history := connDetails.History
				// If history is full pop one
				if history.IsFull() {
					history.Pop()
				}
				// Add to history
				history.Push(data)
				for c, pipes := range connDetails.SPMC {
					// If a connection is not in no repeat, send the message
					_, exists := connDetails.NoRepeat[c]
					if !exists || (c != nil && conn != c) {
						pipes.StdoutChan <- data
					}
				}
				// Now actually send data over to stdin
				connDetails.MChan <- data
				connDetails.SLock.Unlock()
			}
		}
	}

	readFromSocket(conn, connDetails)
}
