package main

import (
  "encoding/base64"
  "golang.org/x/crypto/bcrypt"
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

//Takes a password and returns the base64 encoded hash
func hash(password string) (string, error) {
  //Default cost of hash
  cost := 10
  passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
  if err != nil {
    return "", nil
  }
  encoded := base64.StdEncoding.EncodeToString([]byte(passwordHash))
  return encoded, nil
}

// HTMLHidingFileSystem is an http.FileSystem that "hides" the .html on files
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

	file, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return file, err
}