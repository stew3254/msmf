package main

import (
  "database/sql"
  "fmt"
	"github.com/gorilla/websocket"
  //"github.com/mattn/go-sqlite3"
  "golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
  _ "github.com/mattn/go-sqlite3"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

type User struct {
  username string
  password string
  password_hash []byte
}

func create_db() {
    database, err := sql.Open("sqlite3", "msmf.db")
    if err != nil {
      fmt.Println(err)
    }
    statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, password CHARACTER(60))")
    if err != nil {
      fmt.Println(err)
    }
    statement.Exec()

    statement, err = database.Prepare("INSERT INTO users (id, username, password) VALUES (?, ?, ?)")
    if err != nil {
      fmt.Println(err)
    }
    statement.Exec("0", "Administrator", "$2a$10$DWWK/XA/QuEWr2wAKOie1.F0pogHwx2YrQ3dWoVa333xL8IoCpy2e")
}

//Handles login posts
func login(w http.ResponseWriter, r *http.Request) {
  defer r.Body.Close()
  if r.Method == "POST" {
    r.ParseForm()

    //Create the user
    var user User

    //Look through the map
    for key, value := range r.Form {
      if key == "username" {
        user.username = value[0]
      } else if key == "password" {
        user.password = value[0]
      }

      //Stop looking if we don't need to
      if user.username != "" && user.password != "" {
        break
      }
    }

    //Hash the password and then add the username and hashed password to the struct
    if user.username != "" && user.password != "" {
      temp_hash, err := bcrypt.GenerateFromPassword([]byte(user.password), 10)
      if err != nil {
        fmt.Println(err)
      }
      user.password_hash = temp_hash
      database, err := sql.Open("sqlite3", "msmf.db")
      if err != nil {
        fmt.Println(err)
      }
      defer database.Close()

      rows, err := database.Query("SELECT username, password FROM users where username=?", user.username)
      if err != nil {
        fmt.Println(err)
      }

      defer rows.Close()
      in_db := false
      for rows.Next() {
        username := ""
        hash := make([]byte, 60)
        err:=rows.Scan(&username, &hash)
        if err != nil {
          fmt.Println(err)
        }

        if username == user.username {
          in_db = true;
          if bcrypt.CompareHashAndPassword(hash, []byte(user.password)) == nil {
            fmt.Println("Login succesful")
          } else {
            fmt.Println("Login unsuccesful")
          }
        }
      }

      if !in_db {
        fmt.Println("Not in db")
      }

      err = rows.Err()
      if err != nil {
        fmt.Println(err)
      }
    }
  }

  //Make sure the login page gets served correctly
  http.ServeFile(w, r, "./static/login/index.html")
}

func ws_handler(w http.ResponseWriter, r *http.Request) {
  //Upgrade the http connection
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  for {
    messageType, p, err := conn.ReadMessage()
    if err != nil {
      fmt.Println(err)
      return
    }

    fmt.Println(string(p))
    err = conn.WriteMessage(messageType, []byte("Message from the server!"))

    if err != nil {
      fmt.Println(err)
      return
    }
  }
}

func main() {
  create_db()
  http.Handle("/", http.FileServer(http.Dir("./static")))
  http.HandleFunc("/login", login)
  http.HandleFunc("/websocket", ws_handler)
  log.Fatal(http.ListenAndServe(":8080", nil))
}
