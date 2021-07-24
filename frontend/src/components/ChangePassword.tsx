import * as React from "react";
import {Button, Form} from "react-bootstrap";

export default function ChangePassword() {
    return (
        <Form>
            <Form.Group className="mb-3">
                <Form.Label>Current Password</Form.Label>
                <Form.Control type="password"/>
            </Form.Group>

            <Form.Group className="mb-3">
                <Form.Label>New Password</Form.Label>
                <Form.Control type="password"/>
            </Form.Group>

            <Form.Group className="mb-3">
                <Form.Label>Confirm New Password</Form.Label>
                <Form.Control type="password"/>
            </Form.Group>

            <Button variant="primary" type="submit">Change Password</Button>
        </Form>
    )
}
