package routes

import (
	"bufio"
	"github.com/gorilla/websocket"
	"log"
	"msmf/database"
	"msmf/utils"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// ConnContainer is a helper struct to contain both a websocket connection and console pipes
type ConnContainer struct {
	Conn    *websocket.Conn
	Console utils.Console
}

// PipeChans is a helper struct for the SPMC
type PipeChans struct {
	StdoutChan chan []byte
	StderrChan chan []byte
}

// ConnDetails is a helper struct to hold necessary communication
// information between servers and websockets
type ConnDetails struct {
	// So this can be used to remove from the map later if things go wrong
	ServerID int

	// MPSC - Multiple Producer Single Consumer
	// The channel that all producers will be writing into for stdin
	MChan chan []byte

	// SPMC - Single Producer Multiple Consumer
	// This will connect a single instance of stdout/stderr on a server to multiple websockets
	SPMC map[*websocket.Conn]PipeChans
	// Lock for access to the SPMC
	SLock *sync.Mutex

	// A channel for errors so we know to shut down all go routines related and clean up
	// The function to handle stdin will handle this channel
	ErrChan chan error

	// A place to store all of the server pipes. Guard them well because if they die the server crashes
	Pipes utils.Console
}

// Specify amount of data that can be read from a websocket at a time
var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

// AttachedServers attaches the MPSC and SPMC per server
var AttachedServers = make(map[int]*ConnDetails)

// WsLock is a lock for accessing the Attached Servers map
var WsLock sync.Mutex

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
	WsLock.Lock()
	connDetails, exists := AttachedServers[serverID]
	// If we haven't already attached a server
	if !exists {
		// Attach the server console
		console, err := utils.AttachServer(utils.GameName(serverID))
		if err != nil {
			// Don't forget to unlock
			WsLock.Unlock()

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Create the ConnChan struct
		connDetails = &ConnDetails{
			MChan:   make(chan []byte, 5), // Take up to 5 messages before blocking
			SPMC:    make(map[*websocket.Conn]PipeChans),
			SLock:   &sync.Mutex{},
			ErrChan: make(chan error, 1),
			Pipes:   console,
		}

		// Add it into the map
		AttachedServers[serverID] = connDetails

		// Remember to unlock
		WsLock.Unlock()

		// Upgrade the http connection to a websocket
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create the connection container
		c := ConnContainer{
			Conn:    conn,
			Console: console,
		}

		// Set up the framework for now handling the attachment of the server pipes to go channels
		ServerConsole(connDetails, c)
	} else {
		// Remember to unlock
		WsLock.Unlock()

		// Upgrade the http connection to a websocket
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Make the channels for stdin and stdout
	pipes := PipeChans{
		StdoutChan: make(chan []byte, 5),
		StderrChan: make(chan []byte, 5),
	}

	// Register into the spmc
	connDetails.SLock.Lock()
	connDetails.SPMC[conn] = pipes
	connDetails.SLock.Unlock()

	// Now that the server is attached and handlers are running, link websocket

	writeToSocket := func(conn *websocket.Conn, pipes PipeChans) {
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
	readFromSocket := func(conn *websocket.Conn, connDetails *ConnDetails) {
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

// ServerConsole handles communication between the websocket and the channels
func ServerConsole(connDetails *ConnDetails, c ConnContainer) {
	go func(connDetails *ConnDetails, c ConnContainer) {
		w := c.Console.Stdin
		// Forever write into stdin
		for {
			select {
			// If data received from any websockets, send it into stdin
			case data := <-connDetails.MChan:
				_, err := w.Write(data)
				if err != nil {
					log.Println("stdin error:", err)
					// Send the error to the error handler
					connDetails.ErrChan <- err
				}
			// Handle errors received from any of the pipes by closing all of them and cleaning up
			// TODO see if we can recover from these
			case _ = <-connDetails.ErrChan:
				// Best effort to try to close the pipes, could already be closed
				_ = c.Console.Stdin.Close()
				_ = c.Console.Stdout.Close()
				_ = c.Console.Stderr.Close()

				// Best effort to try to close the websocket, could already be closed
				_ = c.Conn.Close()

				// Delete this server from the servers attached. Since these pipes died, the server
				// is now down. If anything was previously attached, it's dead now anyways
				WsLock.Lock()
				// If this key doesn't exist it doesn't matter
				delete(AttachedServers, connDetails.ServerID)
				WsLock.Unlock()
				// We are done, kill this function
				return
			}
		}
	}(connDetails, c)

	// Create scanners for stdout and stderr
	outReader := bufio.NewScanner(c.Console.Stdout)
	errReader := bufio.NewScanner(c.Console.Stderr)

	// Take in the scanner, connDetails to send stuff to and the corresponding lock
	read := func(scanner *bufio.Scanner, connDetails *ConnDetails, isStdout bool) {
		// Repeatedly scan for more data
		for scanner.Scan() {
			// Read in the data from the scanner
			data := scanner.Bytes()
			if err := scanner.Err(); err != nil {
				if isStdout {
					log.Println("stdout err:", err)
				} else {
					log.Println("stderr err:", err)
				}
				// Send over the error which will handle closing pipes
				connDetails.ErrChan <- err
				return
			}

			// Only one producer can do this at a time
			// This ensures everyone gets their messages in the same order
			connDetails.SLock.Lock()
			// Look through the whole map and send the data on all of the corresponding channels
			for _, pipes := range connDetails.SPMC {
				if isStdout {
					pipes.StdoutChan <- data
				} else {
					pipes.StderrChan <- data
				}
			}
			connDetails.SLock.Unlock()
		}
	}

	// Run readers for stdout and stderr
	go read(outReader, connDetails, true)
	go read(errReader, connDetails, false)
}
