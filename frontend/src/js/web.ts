function startSocket() {
  const ws = new WebSocket("ws://localhost:8080/websocket");
  ws.onopen = function () {
    ws.send("start");
  };
}

function stopSocket() {
  const ws = new WebSocket("ws://localhost:8080/websocket");
  ws.onopen = function () {
    ws.send("start");
  };
}
