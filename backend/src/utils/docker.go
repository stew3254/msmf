package utils

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"msmf/database"
	"os/exec"
	"strings"
	"sync"
)

// Console is a container to hold pipes for the attached connection
type Console struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

// ConnContainer is a helper struct to contain both a websocket connection and console pipes
// type ConnContainer struct {
// 	Conn    *websocket.Conn
// 	Console Console
// }

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
	Pipes Console
}

// AttachedServers attaches the MPSC and SPMC per server
var AttachedServers = make(map[int]*ConnDetails)

// ServerLock is a lock for accessing the Attached Servers map
var ServerLock sync.Mutex

// ServerName returns the docker container name
func ServerName(serverID int) string {
	return fmt.Sprintf("msmf_server_%d", serverID)
}

// ServerConsole handles communication between the websocket and the channels
func ServerConsole(connDetails *ConnDetails, console Console) {
	go func(connDetails *ConnDetails, console Console) {
		w := console.Stdin
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
				_ = console.Stdin.Close()
				_ = console.Stdout.Close()
				_ = console.Stderr.Close()

				// Delete this server from the servers attached. Since these pipes died, the server
				// is now down. If anything was previously attached, it's dead now anyways
				ServerLock.Lock()
				// If this key doesn't exist it doesn't matter
				delete(AttachedServers, connDetails.ServerID)
				ServerLock.Unlock()

				// Update the database to say that this server is no longer started
				database.DB.Model(&database.Server{}).Where(
					"servers.id = ?", connDetails.ServerID,
				).Update("running", false)

				// Make sure to kill the docker server in case it didn't already die
				log.Println("Stopping the server due to crash")
				_ = StopServer(connDetails.ServerID)

				// We are done, kill this function
				return
			}
		}
	}(connDetails, console)

	// Create scanners for stdout and stderr
	outReader := bufio.NewScanner(console.Stdout)
	errReader := bufio.NewScanner(console.Stderr)

	// Take in the scanner, connDetails to send stuff to and the corresponding lock
	read := func(scanner *bufio.Scanner, connDetails *ConnDetails, isStdout bool) {
		// Repeatedly scan for more data
		for scanner.Scan() {
			// Read in the data from the scanner
			data := append(scanner.Bytes(), byte('\n'))
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

func GetContainers(running ...bool) (containers []string) {
	cmdSlice := strings.Fields("docker container ls -a")
	if len(running) > 0 {
		if running[0] {
			cmdSlice = strings.Fields("docker container ls")
		}
	}

	out, err := exec.Command(cmdSlice[0], cmdSlice[1:]...).Output()
	containerStr := string(out)
	if err != nil {
		log.Println(err)
	}
	lines := strings.Split(containerStr, "\n")[1:]
	containers = make([]string, 0, len(lines))
	for _, c := range lines {
		containerSlice := strings.Fields(c)
		if len(containerSlice) > 0 {
			container := containerSlice[len(containerSlice)-1]
			containers = append(containers, container)
		}
	}
	return
}

func GetGameServers(running ...bool) (servers []string) {
	for _, c := range GetContainers(running...) {
		if strings.HasPrefix(c, "msmf_server_") {
			servers = append(servers, c)
		}
	}
	return
}

// CreateServer creates the docker container for the server, but does not start it
func CreateServer(serverID int, image string, isImage bool, parameters []string) {
	var cmdSlice []string
	cmdSlice = append([]string{
		"docker",
		"create",
		"-i",
		"--name",
		ServerName(serverID)},
		parameters...,
	)

	// See if we are using a supported image or a dockerfile
	if isImage {
		cmdSlice = append(cmdSlice, image)
	} else {
		cmdSlice = append(cmdSlice, "-F", "game_dockerfiles/"+image)
	}
	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)

	// Create the docker container
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
}

// DeleteServer deletes a server if it exists
func DeleteServer(serverID int) {
	name := ServerName(serverID)
	var cmdSlice []string
	// First stop the container if it is running
	cmdSlice = append([]string{"docker", "stop", name})
	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}

	// Remove the container
	cmdSlice = append([]string{"docker", "rm", name})
	cmd = exec.Command(cmdSlice[0], cmdSlice[1:]...)
	err = cmd.Run()
	if err != nil {
		log.Println(err)
	}
}

// StartServer starts the docker container and attaches to it
func StartServer(serverID int) (err error) {
	name := ServerName(serverID)
	cmd := exec.Command("docker", "start", "-i", name)

	// Create the console
	var console Console

	// Get stdin pipe
	console.Stdin, err = cmd.StdinPipe()
	if err != nil {
		return err
	}

	// Get stdout pipe
	console.Stdout, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Get stderr pipe
	console.Stderr, err = cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Create the ConnChan struct
	connDetails := &ConnDetails{
		MChan:   make(chan []byte, 5), // Take up to 5 messages before blocking
		SPMC:    make(map[*websocket.Conn]PipeChans),
		SLock:   &sync.Mutex{},
		ErrChan: make(chan error, 1),
		Pipes:   console,
	}

	// Add it into the map
	ServerLock.Lock()
	AttachedServers[serverID] = connDetails
	ServerLock.Unlock()

	// Set up the framework for now handling the attachment of the server pipes to go channels
	ServerConsole(connDetails, console)

	// See if there is a discord integration
	var integration database.DiscordIntegration
	err = database.DB.Where("server_id = ?", serverID).Find(&integration).Error
	// If there is no error continue
	if err == nil {
		// See if we are supposed to use it and that it's a webhook
		if integration.Active && integration.Type == "webhook" {
			// Create a fake connection
			discord := &websocket.Conn{}
			discord = nil

			// Make the channels for stdin and stdout
			pipes := PipeChans{
				// A lot of messages might clog this up because of rate limiting
				StdoutChan: make(chan []byte, 20),
				StderrChan: make(chan []byte, 5),
			}

			// Register with the SPMC
			connDetails.SLock.Lock()
			connDetails.SPMC[discord] = pipes
			connDetails.SLock.Unlock()

			// Go handle the integration
			go SendWebhook(integration, pipes)
		}
	}

	// Start the server
	err = cmd.Start()
	return err
}

// StopServer stops the docker container
func StopServer(serverID int) (err error) {
	name := ServerName(serverID)
	cmd := exec.Command("docker", "stop", name)

	// Start the server
	err = cmd.Run()
	return err
}

// AttachServer attaches to the docker container and returns its pipes
func AttachServer(name string) (console Console, err error) {
	// See if the server is running
	running := false
	servers := GetGameServers(true)
	for _, server := range servers {
		if server == name {
			running = true
		}
	}
	if !running {
		return Console{}, errors.New("cannot attach to a server that isn't running")
	}

	// Try to attach to the server
	var cmdSlice []string
	cmdSlice = append([]string{
		"docker",
		"attach",
		name,
	})
	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)

	// Get stdin pipe
	console.Stdin, err = cmd.StdinPipe()
	if err != nil {
		return console, err
	}

	// Get stdout pipe
	console.Stdout, err = cmd.StdoutPipe()
	if err != nil {
		return console, err
	}

	// Get stderr pipe
	console.Stderr, err = cmd.StderrPipe()
	if err != nil {
		return console, err
	}

	// Start the server
	err = cmd.Start()
	return console, err
}
