export class APIWebSocket extends WebSocket {
    constructor(...path: string[]) {
        const url = "/api/" + path.join("/");
        const loc = window.location;
        const protocol = loc.protocol == "https:" ? "wss:" : "ws:";
        const host = loc.host;

        super(`${protocol}//${host}${url}`);
    }
}
