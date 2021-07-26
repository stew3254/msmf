let socket = new WebSocket("ws://localhost:8080/api/ws/server/1");

socket.onopen = function (e) {
    socket.send("/say hello world!");
};

socket.onmessage = function (e) {
    console.log(e.data)
};

socket.send("Boom shakalaka")
alert("Hell yeah!")
