import * as React from "react";
import {Badge} from "react-bootstrap";
import {useParams} from "react-router-dom";
import {APIWebSocket} from "../api/APIWebSocket";
import {useEffect, useRef, useState} from "react";

function DevConsole() {
    const { serverId } = useParams();
    const webSocket = useRef<WebSocket>();
    const [userInput, setUserInput] = useState("");
    const [consoleLog, setConsoleLog] = useState("");

    function appendConsoleLog(message: string) {
        setConsoleLog((prevConsoleLog) => prevConsoleLog + message + "\n");
    }

    useEffect(() => {
        webSocket.current = new APIWebSocket("server", serverId.toString());

        webSocket.current.onmessage = (message) => {
            appendConsoleLog(message.data);
        };

        return () => {
            webSocket.current.close();
        }
    });

    function handleKeyPress(event: React.KeyboardEvent) {
        if (event.key == "Enter") {
            handleSubmit()
        }
    }

    function handleSubmit() {
        console.log(userInput);
        webSocket.current.send(userInput);
        appendConsoleLog(userInput);
        setUserInput("");
    }

    return (
        <>
            <h1>Console for {serverId} <Badge pill bg="info">dev</Badge></h1>
            <pre className="bg-dark text-light p-3 rounded h-100">{consoleLog}</pre>
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

// class DevConsole2 extends React.Component<{}, DevConsoleState> {
//     webSocket: WebSocket;
//
//     constructor(props) {
//         super(props);
//         this.state = {consoleLog: "", userInput: ""};
//
//         this.webSocket = new APIWebSocket("server");
//
//         this.handleChange = this.handleChange.bind(this);
//         this.handleSubmit = this.handleSubmit.bind(this);
//         this.handleKeyPress = this.handleKeyPress.bind(this);
//     }
//
//     handleChange(event) {
//         this.setState({
//             userInput: event.target.value
//         });
//     }
//
//     handleKeyPress(event: React.KeyboardEvent) {
//         if (event.key == "Enter") {
//             this.handleSubmit()
//         }
//     }
//
//     handleSubmit() {
//         console.log(this.state.userInput);
//         this.setState((prevState) => ({
//             consoleLog: prevState.consoleLog + prevState.userInput + "\n",
//             userInput: ""
//         }));
//     }
//
//     render() {
//         return (
//             <>
//                 <h1>Console <Badge pill bg="info">dev</Badge></h1>
//                 <pre className="bg-dark text-light p-3 rounded h-100">{this.state.consoleLog}</pre>
//                 <div className="input-group">
//                     <input type="text" className="form-control" placeholder="Type a command"
//                            aria-label="Command input" value={this.state.userInput} onChange={this.handleChange}
//                            onKeyPress={this.handleKeyPress}/>
//                     <button className="btn btn-primary" type="button" onClick={this.handleSubmit}>Send</button>
//                 </div>
//             </>
//         );
//     }
// }

export default DevConsole;
