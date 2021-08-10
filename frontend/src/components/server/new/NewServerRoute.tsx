import * as React from "react";
import {useRef, useState} from "react";
import {useHistory} from "react-router-dom";
import {Alert, Button, Form} from "react-bootstrap";
import {postAPI} from "../../../api/APIFetch";

class CreateServerRequest {
    version?: string;

    constructor(readonly name: string, readonly game: string, readonly port: number, version: string) {
        if (version) {
            this.version = version
        }
    }
}

interface NewServerRouteProps {
    updateServerList: () => void
}

export default function NewServerRoute(props: NewServerRouteProps) {
    const [name, setName] = useState("");
    const [game, setGame] = useState("Minecraft");
    const [port, setPort] = useState(25565);
    const [version, setVersion] = useState("");
    const [error, setError] = useState("");
    const formRef = useRef<HTMLFormElement>();
    const history = useHistory();

    function createServer() {
        if (formRef.current.reportValidity()) {
            const data = new CreateServerRequest(name, game, port, version);
            postAPI(data, "server").then(() => {
                props.updateServerList();
                history.push("/");
            }).catch(setError);
        }
    }

    return (
        <>
            <h1>New Server</h1>
            <Alert variant="danger" show={error.length > 0}>
                {error}
            </Alert>
            <Form ref={formRef}>
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
                    <Form.Control type="text" onChange={event => setVersion(event.target.value)} value={version}
                                  required/>
                </Form.Group>

                <Button onClick={createServer}>Create</Button>
            </Form>
        </>
    )
}
