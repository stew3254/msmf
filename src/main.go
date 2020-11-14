package main

import (
  "fmt"
	"log"
	"net/http"
	"os"
	"time"

  "github.com/gorilla/mux"
  
  "msmf/database"
  "msmf/routes"
)

//This is used to register a user
func register(w http.ResponseWriter, r *http.Request) {
  defer r.Body.Close()

  username := r.FormValue("username")
  passwd := r.FormValue("password")
  confirmPasswd := r.FormValue("confirm_password")

  log.Println(username, passwd, confirmPasswd)

  //Make sure the login page gets served correctly
  http.ServeFile(w, r, "static/register.html")
}

func main() {
  // Make DB connection
  err := database.ConnectDB("postgres")
  if err != nil { 
    panic("failed to connect database") 
  } 

  // Used for debug purposes
  // database.DropTables()
  
  // Create all of the tables with constraints
  database.CreateTables()

  // Create base admin account
  database.MakeAdmin()

  // Create new base router for app
  router := mux.NewRouter() 

  // Handlers
  router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Oof, bad place"))
  })
  router.HandleFunc("/login", routes.Login).Methods("POST")

  // Handle API calls
  api := router.PathPrefix("/api").Subrouter()
  api.HandleFunc("/ws", routes.WSHandler)
  api.HandleFunc("/refer", routes.GetReferrals).Methods("GET", "PUT")
  api.HandleFunc("/refer/{id:[0-9]+}", routes.Refer).Methods("GET", "PUT")

  // Handle static traffic
  router.PathPrefix("/").Handler(http.FileServer(HTMLStrippingFileSystem{http.Dir("static")})).Methods("GET")

  // Add printing path to screen per route
  router.Use(printPath)

  // See if a user is authenticated to a page before displaying
  // router.Use(checkAuthenticated)

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

  // Set up tls
  sslString, exists := os.LookupEnv("USE_SSL")
  var ssl bool
  if sslString == "true" || sslString == "yes" {
    ssl = true
  } else if !exists || sslString == "false" || sslString == "no" {
    ssl = false
  }

  // Create http server
  srv := &http.Server{
    Handler:      router,
    Addr:         fmt.Sprintf("%s:%s", listenAddr, port),
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
  }
    
  // Start server
  log.Println("Starting server")
  if ssl {
    log.Fatal(srv.ListenAndServeTLS("certs/cert.crt", "certs/key.pem"))
  } else {
    log.Fatal(srv.ListenAndServe())
  }
}