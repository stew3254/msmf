package routes

import (
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
  ReadBufferSize:  2048,
  WriteBufferSize: 2048,
}

//Websocket handler
func WSHandler(w http.ResponseWriter, r *http.Request) {
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
