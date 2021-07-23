let socket = new WebSocket("ws://localhost:8080/api/ws/server/1");

socket.onopen = function (e) {
    socket.send("hello world!");
};
