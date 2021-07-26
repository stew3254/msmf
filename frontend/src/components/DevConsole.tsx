import * as React from "react";
import {Badge} from "react-bootstrap";

interface DevConsoleState {
    userInput: string,
    consoleLog: string
}

class DevConsole extends React.Component<{}, DevConsoleState> {
    constructor(props) {
        super(props);
        this.state = {consoleLog: "", userInput: ""};

        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleKeyPress = this.handleKeyPress.bind(this);
    }

    handleChange(event) {
        this.setState({
            userInput: event.target.value
        });
    }

    handleKeyPress(event: React.KeyboardEvent) {
        if (event.key == "Enter") {
            this.handleSubmit()
        }
    }

    handleSubmit() {
        console.log(this.state.userInput);
        this.setState((prevState) => ({
            consoleLog: prevState.consoleLog + prevState.userInput + "\n",
            userInput: ""
        }));
    }

    render() {
        return (
            <>
                <h1>Console <Badge pill bg="info">dev</Badge></h1>
                <pre className="bg-dark text-light p-3 rounded h-100">{this.state.consoleLog}</pre>
                <div className="input-group">
                    <input type="text" className="form-control" placeholder="Type a command"
                           aria-label="Command input" value={this.state.userInput} onChange={this.handleChange}
                           onKeyPress={this.handleKeyPress}/>
                    <button className="btn btn-primary" type="button" onClick={this.handleSubmit}>Send</button>
                </div>
            </>
        );
    }
}

export default DevConsole;
