package main

import (
  "bufio"
  "database/sql"
  "fmt"
  "encoding/base64"
  "github.com/gorilla/websocket"
  "golang.org/x/crypto/bcrypt"
  "log"
  "io/ioutil"
  "os"
  "net/http"
)

var upgrader = websocket.Upgrader{
  ReadBufferSize:  2048,
  WriteBufferSize: 2048,
}

type User struct {
  username string
  password string
  password_hash string
}

//Takes a password and returns the base64 encoded hash
func hash(password string) (string, error) {
  //Default cost of hash
  cost := 10
  password_hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
  if err != nil {
    return "", nil
  }
  encoded := base64.StdEncoding.EncodeToString([]byte(password_hash))
  return encoded, nil
}

//Checks to see if a password meets certain requirements
func check_password(password string) (bool) {
  if len(password) < 10 {
    fmt.Println("Password must be at least 10 characters long")
    return false
  }
  return true
}

//Checks to see if the passwords meet certain requirements
func check_passwords(pass1, pass2 string) (bool) {
  if pass1 != pass2 {
    fmt.Println("The two passwords don't match")
    return false
  }
  return check_password(pass1)
}

//Creates the database if it doesn't already exist
func create_db() {
  //See if the database already exists
  exists := false
  files, _ := ioutil.ReadDir(".")
  for _, file := range files {
    if file.Name() == "msmf.db" {
      exists = true
      break
    }
  }

  //Returns if the file exists
  if exists {
    return
  }

  database, err := sql.Open("sqlite3", "msmf.db")
  if err != nil {
    fmt.Println(err)
  }

  //Create the users table
  statement, err := database.Prepare("CREATE TABLE 'users' ('uuid'INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE, 'username'	TEXT NOT NULL UNIQUE, 'password' TEXT NOT NULL);")
  if err != nil {
    fmt.Println(err)
  }
  statement.Exec()

  //Prepare to insert the administrator user into the db
  statement, err = database.Prepare("INSERT INTO users ('uuid', 'username', 'password') VALUES (?, ?, ?)")
  if err != nil {
    fmt.Println(err)
  }

  //Ask user for the password
  good := false
  scanner := bufio.NewScanner(os.Stdin)
  password := ""
  for !good {
    fmt.Print("Administrator password: ")
    scanner.Scan()
    password = scanner.Text()
    fmt.Print("Confirm password: ")
    scanner.Scan()
    confirm_password := scanner.Text()
    if check_passwords(password, confirm_password) {
      good = true
    }
  }

  //Encode and hash the password
  encoded_password, err := hash(password)
  if err != nil {
    fmt.Println(err)
    d

  //Insert the account and encoded password
  statement.Exec("0", "Administrator", encoded_password)
}

//This is used to register a user
func register(w http.ResponseWriter, r *http.Request) {
  defer r.Body.Close()
  if r.Method == "POST" {
    r.ParseForm()

    //Create the user
    var user User
    confirm_password := ""

    //Look through the map
    for key, value := range r.Form {
      if key == "username" {
        user.username = value[0]
      } else if key == "password" {
        user.password = value[0]
      } else if key == "confirm" {
        confirm_password = value[0]
      }

      //Stop looking if we don't need to
      if user.username != "" && user.password != "" && confirm_password != "" {
        break
      }
    }

    //Checks to make sure the passwords match
    if check_passwords(user.password, confirm_password) {
      encoded_password, err := hash(user.password)
      user.password_hash = encoded_password
      if err != nil {
        fmt.Println(err)
      }

      database, err := sql.Open("sqlite3", "msmf.db")
      if err != nil {
        fmt.Println(err)
      }

      //Check if a user is in the db first
      row := database.QueryRow("SELECT username FROM users where username=?", user.username)
      username := ""
      err = row.Scan(&username)

      //If the user isn't in there
      if err == nil {
        fmt.Println("User already exists")
      } else {
        //Insert the user into the db
        statement, err := database.Prepare("INSERT INTO 'users' (username, password) VALUES (?,?)")
        if err != nil {
          fmt.Println(err)
        }

        statement.Exec(user.username, user.password_hash)
        fmt.Println("Added user", user.username)
      }
    }
  }
  //Make sure the login page gets served correctly
  http.ServeFile(w, r, "./static/register.html")
}

//Handles login attempts
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

    //See if the user and password aren't blank
    if user.username != "" && user.password != "" {
      database, err := sql.Open("sqlite3", "msmf.db")
      if err != nil {
        fmt.Println(err)
      }
      defer database.Close()

      //Try to find the user
      row := database.QueryRow("SELECT username, password FROM users where username=?", user.username)

      //Search in the db for the user
      username := ""
      b64_hashed_password := make([]byte, 85)
      //Get the username and password
      err = row.Scan(&username, &b64_hashed_password)
      if err != nil {
        fmt.Println("Not in db")
      } else {
        //Base64 decode the hash and get the raw bcrypt hash
        hashed_password, err := base64.StdEncoding.DecodeString(string(b64_hashed_password))
      if err != nil {
        fmt.Println(err)
      }

        //Make sure the username is correct
        if bcrypt.CompareHashAndPassword(hashed_password, []byte(user.password)) == nil {
          fmt.Println("Login successful")
        } else {
          fmt.Println("Login unsuccessful")
        }
      }
    }
  }

  //Make sure the login page gets served correctly
  http.ServeFile(w, r, "./static/login.html")
}

//Websocket handler
func ws_handler(w http.ResponseWriter, r *http.Request) {
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
  create_db()
  fmt.Println("Starting web server")
  http.Handle("/", http.FileServer(http.Dir("./static")))
  http.HandleFunc("/login", login)
  http.HandleFunc("/register", register)
  http.HandleFunc("/websocket", ws_handler)
  log.Fatal(http.ListenAndServe(":8080", nil))
}
