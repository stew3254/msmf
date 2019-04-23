package main

import (
  "database/sql"
  "fmt"
  "strconv"
  _ "github.com/mattn/go-sqlite3"
)

func create_db(name string) {
    database, err := sql.Open("sqlite3", name)
    if err != nil {
      fmt.Println(err)
    }
    statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, password CHARACTER(60))")
    if err != nil {
      fmt.Println(err)
    }
    statement.Exec()
    statement, _ = database.Prepare("INSERT INTO users (id, username, password) VALUES (?, ?, ?)")
    statement.Exec("0", "Administrator", "$2a$10$DWWK/XA/QuEWr2wAKOie1.F0pogHwx2YrQ3dWoVa333xL8IoCpy2e")
    rows, _ := database.Query("SELECT id, username, password FROM users")
    var id int
    var username string
    var password string
    for rows.Next() {
        rows.Scan(&id, &username, &password)
        fmt.Println(strconv.Itoa(id) + ": " + username + " " + password)
    }
}
