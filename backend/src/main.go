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

	// Handle API calls
	api := router.PathPrefix("/api").Subrouter()
	api.NotFoundHandler = NotFound{}
	api.MethodNotAllowedHandler = MethodNotAllowed{}

	// Handle API calls for referrals
	// Get existing referral codes
	api.HandleFunc("/refer", routes.GetReferrals).Methods("GET")
	// Create new referral codes
	api.HandleFunc("/refer", routes.CreateReferral).Methods("POST")
	// Handle referral code
	api.HandleFunc("/refer/{code:[a-zA-Z0-9]+}", routes.Refer).Methods("GET", "POST")

	// Handle API calls for users
	// Get a list of all users
	api.HandleFunc("/user", routes.GetUsers).Methods("GET")
	// Get your information
	api.HandleFunc("/user/me", routes.GetUser).Methods("GET")
	// Get information about a specific person ('me' and 'perm' are an invalid usernames)
	api.HandleFunc("/user/{user:[a-z0-9]+}", routes.GetUser).Methods("GET")
	// Update your information (such as display name and changing password)
	api.HandleFunc("/user/me", routes.UpdateUser).Methods("PATCH")
	// Get user permissions assigned to all relevant users
	api.HandleFunc("/user/perm", routes.GetUserPerms).Methods("GET")
	// Delete your user account
	api.HandleFunc("/user/me", routes.DeleteUser).Methods("DELETE")
	// Delete a user from the framework
	api.HandleFunc("/user/{user:[a-z0-9]+}", routes.DeleteUser).Methods("DELETE")
	// Get user permissions assigned to yourself
	api.HandleFunc("/user/me/perm", routes.GetUserPerms).Methods("GET")
	// Get user permissions assigned to a particular user
	api.HandleFunc("/user/{user:[a-z0-9]+}/perm", routes.GetUserPerms).Methods("GET")
	// Update user permissions assigned to a particular user
	api.HandleFunc("/user/{user:[a-z0-9]+}/perm", routes.UpdateUserPerms).Methods("PUT")
	// Get server permissions assigned to all relevant servers for yourself
	api.HandleFunc("/user/me/perm/server", routes.GetServerPerms).Methods("GET")
	// Get server permissions assigned to all relevant servers for a particular user
	api.HandleFunc(
		"/user/{user:[a-z0-9]}/perm/server",
		routes.GetServerPerms,
	).Methods("GET")
	// Get server permissions assigned to yourself for a particular server
	api.HandleFunc(
		"/user/me/perm/server/{owner:[a-z0-9]+}/{name:[a-z0-9]+}",
		routes.GetServerPerms,
	).Methods("GET")
	// Get server permissions assigned to a user for a particular server
	api.HandleFunc(
		"/user/{user:[a-z0-9]}/perm/server/{owner:[a-z0-9]+}/{name:[a-z0-9]+}",
		routes.GetServerPerms,
	).Methods("GET")
	// Update server permissions for yourself and a particular server
	api.HandleFunc(
		"/user/me/perm/server/{owner:[a-z0-9]+}/{name:[a-z0-9]+}",
		routes.UpdateServerPerms,
	).Methods("PUT")
	// Update server permissions for a particular user and server
	api.HandleFunc(
		"/user/{user:[a-z0-9]}/perm/server/{owner:[a-z0-9]+}/{name:[a-z0-9]+}",
		routes.UpdateServerPerms,
	).Methods("PUT")

	// Handle API calls for servers
	// Create a new server
	api.HandleFunc("/server", routes.CreateServer).Methods("POST")
	// List all visible servers
	api.HandleFunc("/server", routes.GetServers).Methods("GET")
	// View a server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}",
		routes.GetServer,
	).Methods("GET")
	// Update a server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}",
		routes.UpdateServer,
	).Methods("PATCH")
	// Delete a server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}",
		routes.DeleteServer,
	).Methods("DELETE")
	// Start a server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}/start",
		routes.StartServer,
	).Methods("POST")
	// Stop a server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}/stop",
		routes.StopServer,
	).Methods("POST")
	// Restart a server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}/restart",
		routes.RestartServer,
	).Methods("POST")

	// Get server permissions assigned to all relevant users and servers
	api.HandleFunc("/server/perm", routes.GetServerPerms).Methods("GET")
	// Get server permissions assigned to all relevant users for a particular server
	api.HandleFunc("/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}/perm",
		routes.GetServerPerms).Methods("GET")
	// Get server permissions assigned to yourself for a particular server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}/perm/me",
		routes.GetServerPerms,
	).Methods("GET")
	// Get server permissions assigned to a particular user for a particular server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}/perm/{user:[a-z0-9]+}",
		routes.GetServerPerms,
	).Methods("GET")
	// Update permissions assigned to yourself for a particular server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}/perm/me",
		routes.UpdateServerPerms,
	).Methods("PUT")
	// Update permissions assigned to a particular user for a particular server
	api.HandleFunc(
		"/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}/perm/{user:[a-z0-9]+}",
		routes.UpdateServerPerms,
	).Methods("PUT")

	// Handle websocket connections for a server console
	api.HandleFunc("/server/{owner:[a-z0-9]+}/{name:[a-z0-9-]+}/console", routes.WsServerHandler)

	// Create or Update an integration with Discord
	api.HandleFunc(
		"/server/{user:[a-z0-9]+}/{name:[a-z0-9-]+}/discord",
		routes.MakeIntegration,
	).Methods("PUT")
	// Delete an integration with Discord
	api.HandleFunc(
		"/server/{user:[a-z0-9]+}/{name:[a-z0-9-]+}/discord",
		routes.DeleteIntegration,
	).Methods("DELETE")
	// View an integration with Discord
	api.HandleFunc(
		"/server/{user:[a-z0-9]+}/{name:[a-z0-9-]+}/discord",
		routes.GetIntegration,
	).Methods("GET")

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
		log.Fatal(srv.ListenAndServeTLS("certs/cert.pem", "certs/privkey.pem"))
	} else {
		log.Fatal(srv.ListenAndServe())
	}
}
