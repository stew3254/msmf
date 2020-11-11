package main

import (
  "fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
  "github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB
var err error

var upgrader = websocket.Upgrader{
  ReadBufferSize:  2048,
  WriteBufferSize: 2048,
}

//This is used to register a user
func register(w http.ResponseWriter, r *http.Request) {
  defer r.Body.Close()
  r.ParseForm()

  // username := r.Form.Get("username")
  // passwd := r.Form.Get("password")
  // confirmPasswd := r.Form.Get("confirm_password")

  //Make sure the login page gets served correctly
  http.ServeFile(w, r, "static/register.html")
}

//Handles login attempts
func login(w http.ResponseWriter, r *http.Request) {
  defer r.Body.Close()
  r.ParseForm()

  // username := r.Form.Get("username")
  // passwd := r.Form.Get("password")

  //Make sure the login page gets served correctly
  http.ServeFile(w, r, "static/login.html")
}

//Websocket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
  //Upgrade the http connection
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println(err)
    return
  }

  //Tries to read messages forever
  for {
    messageType, p, err := conn.ReadMessage()
    if err != nil {
      fmt.Println(err)
      return
    }

    //Prints and sends one back
    fmt.Println(string(p))
    err = conn.WriteMessage(messageType, []byte("Message from the server!"))

    if err != nil {
      fmt.Println(err)
      return
    }
  }
}

func main() {
  // Create new base router for app
  router := mux.NewRouter() 

  // Make DB connection
  db, err = connectDB("postgres")
  if err != nil { 
    panic("failed to connect database") 
  } 
  defer db.Close() 

  // Used for debug purposes
  dropTables(db)
  
  // Create all of the tables with constraints
  createTables(db)

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

  // Handlers
  router.HandleFunc("/login", login).Methods("POST")
  router.HandleFunc("/register", register).Methods("POST")
  router.HandleFunc("/ws", wsHandler)

  // Handle static traffic
  router.PathPrefix("/").Handler(http.FileServer(htmlStrippingFileSystem{http.Dir("static")})).Methods("GET")

  router.Use(logRequest)

  // Create http server
  srv := &http.Server{
    Handler:      router,
    Addr:         fmt.Sprintf("%s:%s", listenAddr, port),
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
  }
    
  // Start server
  log.Println("Starting server")
  log.Fatal(srv.ListenAndServe())
}