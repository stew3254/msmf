function startSocket() {
  ws.onopen = function () {
    ws.send(document.getElementById("text").value);
  };
}
var ws = new WebSocket("ws://localhost:8080/websocket");
console.log("REE");
