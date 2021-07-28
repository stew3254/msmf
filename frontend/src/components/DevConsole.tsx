import * as React from "react";
import {Badge, Button, ButtonGroup} from "react-bootstrap";
import {useParams} from "react-router-dom";
import {APIWebSocket} from "../api/APIWebSocket";
import {useEffect, useRef, useState} from "react";

function DevConsole() {
    const { serverId } = useParams();
    const logId = useRef(0);
    const webSocket = useRef<WebSocket>();
    const [serverState, setServerState] = useState("unknown");
    const [userInput, setUserInput] = useState("");
    const [consoleLog, setConsoleLog] = useState([]);

    useEffect(() => {
        webSocket.current = new APIWebSocket("ws", "server", serverId.toString());

        webSocket.current.onmessage = (message) => {
            const log = {key: logId.current++, message: message.data.toString()};
            setConsoleLog((prevConsoleLog) => [log].concat(prevConsoleLog));
        };

        return () => {
            webSocket.current.close();
        }
    }, [serverState]);

    function handleKeyPress(event: React.KeyboardEvent) {
        if (event.key == "Enter") {
            handleSubmit()
        }
    }

    function handleSubmit() {
        webSocket.current.send(userInput);
        setUserInput("");
    }

    function setState(state: string) {
        fetch(`/api/server/${serverId}/${state}`, {method: 'POST'}).then(() => {
            setServerState(state);
        });
    }

    return (
        <>
            <div className="d-flex justify-content-between">
            <h1>Console for {serverId} <Badge pill bg="info">dev</Badge></h1>
            <ButtonGroup>
                <Button variant="success" onClick={() => setState("start")}>Start</Button>
                <Button variant="primary" onClick={() => setState("restart")}>Restart</Button>
                <Button variant="danger" onClick={() => setState("stop")}>Stop</Button>
            </ButtonGroup>
            </div>
            <pre className="bg-dark text-light p-3 rounded dev-console">
                {consoleLog.map(log => {
                    return <span key={log.key}>{log.message}</span>
                })}
            </pre>
            <div className="input-group">
                <input type="text" className="form-control" placeholder="Type a command"
                       aria-label="Command input"
                       value={userInput}
                       onChange={event => setUserInput(event.target.value)}
                       onKeyPress={handleKeyPress}/>
                <button className="btn btn-primary" type="button" onClick={handleSubmit}>Send</button>
            </div>
        </>
    );
}

export default DevConsole;
