import * as React from "react";
import {useState} from "react";
import {Button, Form} from "react-bootstrap";
import * as console from "console";

export default function NewServerRoute() {
    const [name, setName] = useState("");
    const [game, setGame] = useState("Minecraft");
    const [port, setPort] = useState(25565);
    const [version, setVersion] = useState("");

    function createServer() {
        const data = JSON.stringify({name: name, game: game, port: port, version: version});
        fetch("/api/server", {method: "POST", body: data}).then(console.log)
    }

    return (
        <>
            <h1>New Server</h1>
            <Form>
                <Form.Group className="mb-3">
                    <Form.Label>Server Name</Form.Label>
                    <Form.Control type="text" onChange={event => setName(event.target.value)} value={name} required/>
                </Form.Group>

                <Form.Group className="mb-3">
                    <Form.Label>Game</Form.Label>
                    <Form.Control type="text" onChange={event => setGame(event.target.value)} value={game} readOnly
                                  required/>
                </Form.Group>

                <Form.Group className="mb-3">
                    <Form.Label>Port</Form.Label>
                    <Form.Control type="number" onChange={event => setPort(parseInt(event.target.value))} value={port}
                                  required/>
                </Form.Group>

                <Form.Group className="mb-3">
                    <Form.Label>Version</Form.Label>
                    <Form.Control type="text" placeholder="latest" onChange={event => setVersion(event.target.value)}
                                  value={version}/>
                </Form.Group>

                <Button onClick={createServer}>Create</Button>
            </Form>
        </>
    )
}
