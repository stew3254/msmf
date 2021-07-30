package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"time"
)

func CreateConnection(url string, auth Auth) (*websocket.Conn, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	cookie, err := Login(client, "http://localhost:8080/login", auth)

	if err != nil {
		return nil, err
	}

	dialer := websocket.Dialer{}
	// Add authentication cookies
	header := http.Header{}
	header.Add("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))

	conn, _, err := dialer.Dial(url, header)
	return conn, err
}

func RunConsole(url string, auth Auth) error {
	stdinChan := make(chan []byte, 1)
	// Read in from stdin
	go func(stdinChan chan<- []byte) {
		// Send messages
		stdin := bufio.NewScanner(os.Stdin)

		for stdin.Scan() {
			// Get the data
			data := stdin.Bytes()

			// Close the channel on error
			if stdin.Err() != nil {
				close(stdinChan)
				return
			}

			// Send over the data
			stdinChan <- data
		}
		// We sent EOF and we are done
		close(stdinChan)
	}(stdinChan)

	for {
		conn, err := CreateConnection(url, auth)

		if err != nil {
			// Sleep for 5 seconds and try again
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Connected to the server")

		// Ping messages for keep alive
		go func(conn *websocket.Conn) {
			// Send message to say don't repeat stdin
			err := conn.WriteControl(websocket.PingMessage, []byte("no-repeat"),
				time.Now().Add(5*time.Minute))
			if err != nil {
				return
			}
			for {
				time.Sleep(5 * time.Second)
				err := conn.WriteMessage(websocket.PingMessage, []byte("Ping!"))
				if err != nil {
					return
				}
			}

		}(conn)

		// Tell the writer to die
		dieChan := make(chan interface{})

		// Read messages from stdin and write them
		go func(conn *websocket.Conn, stdinChan <-chan []byte) {
			for {
				select {
				// Die if you receive a message
				case <-dieChan:
					return
				case data, ok := <-stdinChan:
					if !ok {
						// Channel is closed, so die
						return
					}

					// Write the message
					err := conn.WriteMessage(websocket.TextMessage, data)
					if err != nil {
						// Websocket died so die, a new one will be spun up to replace this
						return
					}
				}
			}
		}(conn, stdinChan)

		// Read messages
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				// Best effort close the connection
				_ = conn.Close()

				// Send a message over the die chan to kill the writer
				dieChan <- true

				fmt.Println("Connection died due to error", err)
				break
			}
			if messageType == websocket.TextMessage {
				fmt.Print(string(data))
			} else if messageType == websocket.BinaryMessage {
				fmt.Printf("ERROR: %s", string(data))
			}
		}
	}
}
