import * as React from "react";
import {Button, Form} from "react-bootstrap";

export default function Login() {
    return (
        <Form>
            <Form.Group className="mb-3">
                <Form.Label>Username</Form.Label>
                <Form.Control type="text"/>
            </Form.Group>

            <Form.Group className="mb-3">
                <Form.Label>Password</Form.Label>
                <Form.Control type="password"/>
            </Form.Group>

            <Button variant="primary" type="submit">Log in</Button>
        </Form>
    )
}
