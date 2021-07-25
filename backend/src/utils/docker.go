package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

type Console struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

// GameName returns the docker container name
func GameName(serverID int) string {
	return fmt.Sprintf("msmf_server_%d", serverID)
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
		GameName(serverID)},
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
func DeleteServer(name string) {
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

// StartServer starts the docker container
func StartServer(name string) (err error) {
	var cmdSlice []string
	cmdSlice = append([]string{
		"docker",
		"start",
		name,
	})
	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)

	// Start the server
	err = cmd.Start()
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
