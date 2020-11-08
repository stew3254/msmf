package main

import (
  "fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Check if a value exists and fail if it doesn't
func checkExists(exists bool, msg string) {
	if !exists {
		log.Fatal(msg)
	}
}

var db *gorm.DB
var err error

func main() {
  // Create new base router for app
  router := mux.NewRouter() 

  // Make DB connection
  db, err = connectDB("postgres")
  if err != nil { 
    panic("failed to connect database") 
  } 
  defer db.Close() 

  // Set up server listen address
  listenAddr, exists := os.LookupEnv("LISTEN")
  if !exists {
    listenAddr = ""
  }

  // Set up server port
  port, exists := os.LookupEnv("PORT")
  if !exists {
    port = "5000"
  }

  // Handle static traffic
  router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))
  
  // Create http server
  srv := &http.Server{
    Handler:      router,
    Addr:         fmt.Sprintf("%s:%s", listenAddr, port),
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
  }
    
  // Start server
  log.Fatal(srv.ListenAndServe())
}