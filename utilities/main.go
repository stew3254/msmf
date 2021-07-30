package main

import "log"

func main() {
	err := RunConsole("ws://localhost:8080/api/ws/server/1", Auth{
		Username: "admin",
		Password: "foobar",
	})

	if err != nil {
		log.Fatalln(err)
	}
}
