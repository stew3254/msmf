class WebsocketTest {

    ws: WebSocket = new WebSocket("ws://localhost:8080/websocket");

    startSocket(): void {
        let myself: WebsocketTest = this;

        this.ws.onopen = function (): void {
            let text: HTMLInputElement = <HTMLInputElement> document.getElementById("text");
            myself.ws.send(text.value);
        };
    }
}