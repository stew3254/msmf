package main

import (
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"log"
)

func main() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if f.Name() == "server.jar" {
			fmt.Println("Server Jar found!")
			break
		}
	}
	fmt.Println("Server Jar not found!")
}
