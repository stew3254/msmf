package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Stdin Handles stdin from the server
func Stdin(stdin *io.WriteCloser) {
	defer (*stdin).Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		io.Copy(*stdin, reader)
	}
}

// Stderr Handles stderr from the server
func Stderr(stderr *io.ReadCloser) {
	scanner := bufio.NewScanner(*stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		log.Println(m)
	}
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

func CreateServer(name string, image string, isImage bool, parameters []string) {
	var cmdSlice []string
	if isImage {
		cmdSlice = append([]string{"docker", "run"}, parameters...)
		cmdSlice = append(cmdSlice, image)
	} else {
		cmdSlice = append([]string{"docker", "run"}, parameters...)
		cmdSlice = append(cmdSlice, "-F", image)
	}
	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)

	// Get stdin
	// stdin, err := cmd.StdinPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Get stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Get stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	// Handle stderr
	go Stderr(&stderr)

	// Start the docker container
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Print output from the server
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	// cmd.Wait()
}
