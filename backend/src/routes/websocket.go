package routes

import (
	"bufio"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"msmf/database"
	"msmf/utils"
	"net/http"
	"strconv"
	"strings"
)

// Connection is a helper struct to contain both a websocket connection and console pipes
type Connection struct {
	Conn    *websocket.Conn
	Console utils.Console
}

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

	// User does not have permissions to view this server
	if err != nil || count == 0 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Attach the server console
	console, err := utils.AttachServer(utils.GameName(serverID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Upgrade the http connection to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the connection wrapper
	c := Connection{
		Conn:    conn,
		Console: console,
	}

	// Handle the console connection
	go ServerConsole(c)
}

// ServerConsole handles communication between the websocket and the pipes
func ServerConsole(conn Connection) {
	// Create new buffered scanners
	outReader := bufio.NewScanner(conn.Console.Stdout)
	errReader := bufio.NewScanner(conn.Console.Stderr)

	// Allow up to 5 queued messages
	stdoutChan := make(chan []byte, 5)
	stderrChan := make(chan []byte, 5)

	// Handle reading messages from server to send to websocket
	outHandler := func(reader io.ReadCloser, scanner *bufio.Scanner, c chan<- []byte) {
		// Scan the data
		for scanner.Scan() {
			// Send over the data once received
			c <- scanner.Bytes()

			// Close if there is an error
			if scanner.Err() != nil {
				// Best effort
				_ = reader.Close()
				return
			}
		}
	}

	// Make Readers
	go outHandler(conn.Console.Stdout, outReader, stdoutChan)
	go outHandler(conn.Console.Stderr, errReader, stderrChan)

	// Handle reading messages from websocket and send them to the server
	inHandler := func(w io.WriteCloser) {
		for {
			// Read message from websocket and send to the server
			n, data, err := conn.Conn.ReadMessage()
			// If reading fails close the pipe and return
			if err != nil {
				// Best effort
				_ = conn.Conn.Close()
				_ = w.Close()
				return
			}

			// If received a text message pass it on through the pipe
			if n == websocket.TextMessage {
				if len(data) > 0 {
					data = append(data, byte('\n'))
					n, err = w.Write(data)
					// If writing fails close the pipe and return
					if err != nil {
						// Best effort
						_ = conn.Conn.Close()
						_ = w.Close()
						return
					}
				}
			}
		}
	}

	// Read the messages from the websocket and send to the server
	go inHandler(conn.Console.Stdin)

	// Get messages coming from the server and send them to the websocket
	for {
		var err error
		select {
		case data := <-stdoutChan:
			err = conn.Conn.WriteMessage(websocket.TextMessage, data)
		case data := <-stderrChan:
			// Send stderr as a binary message so you can tell them apart
			err = conn.Conn.WriteMessage(websocket.BinaryMessage, data)
		}
		// If there is an issue close the socket and exit
		if err != nil {
			// Best effort
			_ = conn.Conn.Close()
			return
		}
	}
}
