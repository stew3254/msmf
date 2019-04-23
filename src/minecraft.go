package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

//Handles stdin from the server
func Stdin(stdin *io.WriteCloser) {
	defer (*stdin).Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		io.Copy(*stdin, reader)
	}
}

//Handles stderr from the server
func Stderr(stderr *io.ReadCloser) {
	scanner := bufio.NewScanner(*stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
}

func StartServer() {
	fmt.Print("Input the directory to your server: ")
	//Read in directory name
	reader := bufio.NewReader(os.Stdin)
	dir, err := reader.ReadString('\n')
	dir = dir[0 : len(dir)-1]
	if err != nil {
		log.Fatal(err)
		return
	}

	//Change to that directory
	err = os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
		return
	}

	//Server start command
	cmd := exec.Command("java", "-Xmx1024M", "-Xms1024M", "-jar", "server.jar", "nogui")
	//Change stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	//Catch Control C presses and properly stop the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		io.WriteString(stdin, "stop")
		stdin.Close()
	}()

	//Changes stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	//Changes stderr pipe
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	//Start the server
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	//Handle input to the server
	go Stdin(&stdin)
	if err != nil {
		log.Fatal(err)
	}

	//Print error from the server
	go Stderr(&stderr)

	//Print output from the server
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	cmd.Wait()
}
