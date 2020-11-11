package main

import (
	"net/http"
	"strings"
)

// HTMLStrippingFileSystem is an http.FileSystem that "hides" the .html on files
type HTMLStrippingFileSystem struct {
	http.FileSystem
}

// Open is a wrapper around the Open method of the embedded FileSystem
// This allows us to take urls without a file extension and open them as a .html
// UNAVOIDABLE BUG: When going to a directory index.html,
// the url will not have a trailing slash
func (fs HTMLStrippingFileSystem) Open(name string) (http.File, error) {

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
		// if strings.Contains(err.Error(), "no such file or directory") {
		// 	fs.FileSystem.Open()
		// }
		return nil, err
	}
	return file, err
}