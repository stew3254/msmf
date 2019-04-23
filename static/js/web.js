function startSocket() {
  var ws = new WebSocket("ws://localhost:8080/websocket");
  ws.onopen = function (event) {
    ws.send("start");
  };
}

function stopSocket() {
  var ws = new WebSocket("ws://localhost:8080/websocket");
  ws.onopen = function (event) {
    ws.send("start");
  };
}
