var WebsocketTest = /** @class */ (function () {
    function WebsocketTest() {
        this.ws = new WebSocket("ws://localhost:8080/websocket");
    }
    WebsocketTest.prototype.startSocket = function () {
        var myself = this;
        this.ws.onopen = function () {
            var text = document.getElementById("text");
            myself.ws.send(text.value);
        };
    };
    return WebsocketTest;
}());
//# sourceMappingURL=web.js.map