package main

import (
	// "time"
	// "log"
	// "encoding/json"
	"net/http"
	"strings"
)

// type hookedResponseWriter struct {
//     http.ResponseWriter
// 		ignore bool
// }

// func (hrw *hookedResponseWriter) WriteHeader(status int) {
// 	r := hrw.Request
// 	hrw.ResponseWriter.WriteHeader(status)
// 	if status == 404 {
// 		hrw.ignore = true
// 		jsonOut, err := json.Marshal(r.URL.Query())
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		queryParams := string(jsonOut)

// 		r.ParseForm()
// 		jsonOut, err = json.Marshal(r.PostForm)
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		postData := string(jsonOut)

// 		jsonOut, err = json.Marshal(r.Cookies())
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		cookies := string(jsonOut)

// 		// Not complete yet
// 		db.Create(&WebLog{
// 			Time: time.Now(),
// 			IP: strings.Split(r.RemoteAddr, ":")[0],
// 			Method: r.Method,
// 			QueryParams: queryParams,
// 			PostData: postData,
// 			Cookies: cookies,
// 		})
// 			// Write custom error here to hrw.ResponseWriter
// 	}
// }

// func (hrw *hookedResponseWriter) Write(p []byte) (int, error) {
//     if hrw.ignore {
//         return len(p), nil
//     }
//     return hrw.ResponseWriter.Write(p)
// }

// type NotFoundHook struct {
//     h http.Handler
// }

// func (nfh NotFoundHook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//     nfh.h.ServeHTTP(&hookedResponseWriter{ResponseWriter: w}, r)
// }

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
			// Do nothing
		} else if strings.HasSuffix(name, "/") {
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

	file, err := fs.FileSystem.Open(name[1:])
	if err != nil {
		return nil, err
	}
	return file, err
}
