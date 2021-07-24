import * as React from "react";
import {Badge} from "react-bootstrap";

export default function DevConsole() {
    return (
        <>
            <h1>Console <Badge pill bg="info">dev</Badge></h1>
            <pre className="bg-dark text-light p-3 rounded h-100">
                This will be a console you can type in!
            </pre>
        </>
    )
}
