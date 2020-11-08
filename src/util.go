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

// UNAVOIDABLE BUG: When going to a directory index.html,
// the url will not have a trailing slash
// Open is a wrapper around the Open method of the embedded FileSystem
// This allows us to take urls without a file extension and open them as a .html
func (fs htmlStrippingFileSystem) Open(name string) (http.File, error) {

	// Directory doesn't exist, now checking 
	if len(strings.Split(name, ".")) == 1 {
		if name == "/" {
			//Do nothing
		} else if strings.HasSuffix(name, "/"){
			name += "index.html"
		} else {
			// See if a file exists as a directory already
			file, err := fs.FileSystem.Open(name)
			if err == nil {
				fi, err := file.Stat()
				if err != nil {
					return file, err
				}

				// Prioritize directories over html files with the same name
				if fi.IsDir() {
					name += "/index.html"
				} else {
					return file, err
				}
			} else {
				// Not a directory, so add the html
				name += ".html"
			}
		}
	}

	fmt.Println(name)

	file, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return file, err
}