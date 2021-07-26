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

func main() {
	// Make DB connection
	err := database.ConnectDB("postgres")
	if err != nil {
		panic("failed to connect database")
	}

	// Used for debugging purposes
	// database.DropTables()

	// Create all of the tables with constraints and add all necessary starting information
	database.MakeDB()

	// Create new base router for app
	router := mux.NewRouter()

	// Handlers
	// Not working?
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Oof, bad place"))
	})
	// Handle logins
	router.HandleFunc("/login", routes.Login).Methods("POST")
	// Handle changing password
	router.HandleFunc("/change-password", routes.ChangePassword).Methods("POST")

	// Handle API calls
	api := router.PathPrefix("/api").Subrouter()
	// Handle calls to create servers
	api.HandleFunc("/server", routes.CreateServer).Methods("POST")
	// Handle calls to list servers
	api.HandleFunc("/server", routes.GetServers).Methods("GET")
	// Handle calls to view a server
	api.HandleFunc("/server/{id:[0-9]+}", routes.GetServer).Methods("GET")
	// Handle calls to update a server
	api.HandleFunc("/server/{id:[0-9]+}", routes.UpdateServer).Methods("PATCH")
	// Handle calls to delete servers
	api.HandleFunc("/server/{id:[0-9]+}", routes.DeleteServer).Methods("DELETE")
	// Handle calls to start a server
	api.HandleFunc("/server/{id:[0-9]+}/start", routes.StartServer).Methods("GET")
	// Handle calls to stop a server
	api.HandleFunc("/server/{id:[0-9]+}/stop", routes.StopServer).Methods("GET")
	// Handle calls to restart a server
	api.HandleFunc("/server/{id:[0-9]+}/restart", routes.RestartServer).Methods("GET")

	// Handle websocket connections for server consoles
	api.HandleFunc("/ws/server/{id:[0-9]+}", routes.WsServerHandler)

	// Get existing referral codes
	api.HandleFunc("/refer", routes.GetReferrals).Methods("GET")
	// Creating new referral codes
	api.HandleFunc("/refer/new", routes.CreateReferral).Methods("GET")
	// Handle referral code
	api.HandleFunc("/refer/{id:[0-9]+}", routes.Refer).Methods("GET", "POST")

	// Get user permissions
	api.HandleFunc("/perm", routes.GetPerms).Methods("GET")

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

	// Set up ssl
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
