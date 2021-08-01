package main

import (
	"fmt"
	"log"
	"msmf/utils"
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

	// Create all the tables with constraints and add all necessary starting information
	// if it doesn't already exist
	database.MakeDB()

	// If servers were running when msmf stopped, start them up again
	var servers []database.Server
	database.DB.Where("servers.running = ?", true).Find(&servers)
	for _, server := range servers {
		// Run them as goroutines so the serer start up is faster
		go func(server database.Server) {
			log.Printf("Restarting server %d since it is supposed to be started", *server.ID)

			// Lock on restart
			lock := utils.GetLock(*server.ID)
			lock.Lock()
			_ = utils.StopServer(*server.ID, false)
			err := utils.StartServer(*server.ID, false)
			lock.Unlock()

			// TODO come up with a solution to remedy this
			if err != nil {
				log.Printf("Server %d no longer exists in docker\n", *server.ID)
			}
		}(server)
	}

	// Create new base router for app
	router := mux.NewRouter()

	// Handlers

	// Handle logins
	router.HandleFunc("/login", routes.Login).Methods("POST")
	// Handle changing password
	router.HandleFunc("/change-password", routes.ChangePassword).Methods("POST")

	// Handle API calls
	api := router.PathPrefix("/api").Subrouter()
	api.NotFoundHandler = NotFound{}
	api.MethodNotAllowedHandler = MethodNotAllowed{}

	// Handle calls to create servers
	api.HandleFunc("/server", routes.CreateServer).Methods("POST")
	// Handle calls to list servers
	api.HandleFunc("/server", routes.GetServers).Methods("GET")
	// Handle calls to view a server
	api.HandleFunc("/server/{id:[0-9]+}", routes.GetServer).Methods("GET")
	// Handle calls to update a server
	api.HandleFunc("/server/{id:[0-9]+}", routes.UpdateServer).Methods("PATCH")
	// Handle calls to delete servers
	// TODO consider returning No Content instead
	api.HandleFunc("/server/{id:[0-9]+}", routes.DeleteServer).Methods("DELETE")
	// Handle calls to start a server
	// TODO consider returning No Content instead
	api.HandleFunc("/server/{id:[0-9]+}/start", routes.StartServer).Methods("POST")
	// Handle calls to stop a server
	// TODO consider returning No Content instead
	api.HandleFunc("/server/{id:[0-9]+}/stop", routes.StopServer).Methods("POST")
	// Handle calls to restart a server
	// TODO consider returning No Content instead
	api.HandleFunc("/server/{id:[0-9]+}/restart", routes.RestartServer).Methods("POST")

	// Handle websocket connections for server consoles
	api.HandleFunc("/ws/server/{id:[0-9]+}", routes.WsServerHandler)

	// Handle creating and updating integrations with Discord
	// TODO issue with method used, reconsider the method used and possibly add additional ones
	api.HandleFunc("/discord/server/{id:[0-9]+}", routes.MakeIntegration).Methods("PUT")
	// Handle deleting integrations with Discord
	// TODO consider returning No Content instead
	api.HandleFunc("/discord/server/{id:[0-9]+}", routes.DeleteIntegration).Methods("DELETE")
	// Handle getting an integration with Discord
	api.HandleFunc("/discord/server/{id:[0-9]+}", routes.GetIntegration).Methods("GET")

	// Get existing referral codes
	api.HandleFunc("/refer", routes.GetReferrals).Methods("GET")
	// Create new referral codes
	api.HandleFunc("/refer", routes.CreateReferral).Methods("POST")
	// Handle referral code
	api.HandleFunc("/refer/{id:[0-9]+}", routes.Refer).Methods("GET", "POST")

	// Get user permissions assigned to all relevant users
	api.HandleFunc("/perm/user", routes.GetUserPerms).Methods("GET")
	// Get user permissions assigned to a particular user
	api.HandleFunc("/perm/user/{name}", routes.GetUserPerms).Methods("GET")
	// Update permissions assigned to a particular user
	api.HandleFunc("/perm/user/{name}", routes.UpdateUserPerms).Methods("PUT")

	// Get server permissions assigned to all relevant users and servers
	api.HandleFunc("/perm/server", routes.GetServerPerms).Methods("GET")
	// Get server permissions assigned to all relevant users for a particular server
	api.HandleFunc("/perm/server/{id:[0-9]+}", routes.GetServerPerms).Methods("GET")
	// Get server permissions assigned to all relevant servers for a particular user
	api.HandleFunc("/perm/server/user/{name}", routes.GetServerPerms).Methods("GET")
	// Get server permissions assigned to a particular user for a particular server
	api.HandleFunc("/perm/server/{id:[0-9]+}/user/{name}", routes.GetServerPerms).Methods("GET")
	// Update permissions assigned to a particular user for a particular server
	api.HandleFunc("/perm/server/{id:[0-9]+}/user/{name}", routes.UpdateServerPerms).Methods("PUT")

	// Handle static traffic
	router.PathPrefix("/").Handler(http.FileServer(HTMLStrippingFileSystem{http.Dir("static")})).Methods("GET")

	// Add printing path to screen per route
	router.Use(printPath)

	// See if a user is authenticated to a page before displaying
	router.Use(checkAuthenticated)

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
	log.Println("Web server is now listening for connections")
	if ssl {
		log.Fatal(srv.ListenAndServeTLS("certs/cert.crt", "certs/key.pem"))
	} else {
		log.Fatal(srv.ListenAndServe())
	}
}
