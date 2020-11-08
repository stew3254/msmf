package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Checks if a value exists and fail if it doesn't
func checkExists(exists bool, msg string) {
	if !exists {
		log.Fatal(msg)
	}
}

// HTMLHidingFileSystem is an http.FileSystem that hides
// hidden "dot files" from being served.
type htmlStrippingFileSystem struct {
	http.FileSystem
}

// Open is a wrapper around the Open method of the embedded FileSystem
// This allows us to take urls without a file extension and open them as a .html
func (fs htmlStrippingFileSystem) Open(name string) (http.File, error) {
	path := name

	if len(strings.Split(path, ".")) == 1 {
		if path == "/" {
			//Do nothing
		} else if strings.HasSuffix(path, "/"){
			path += "index.html"
		} else {
			path += ".html"
		}
	}

	fmt.Println(path)

	file, err := fs.FileSystem.Open(path)
	if err != nil {
		return nil, err
	}
	return file, err
}