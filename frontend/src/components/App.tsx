import * as React from "react";
import {useEffect, useState} from "react";
import {BrowserRouter as Router, Link, Route, Switch} from "react-router-dom";
import {Badge, Container, Nav, Navbar, NavDropdown, NavItem, NavLink, Row} from "react-bootstrap";
import Home from "./Home";
import Login from "./Login";
import Invite from "./Invite";
import Register from "./Register";
import ChangePassword from "./ChangePassword";
import DevConsole from "./DevConsole";

export default function App() {
    // const location = useLocation();
    const [servers, setServers] = useState([]);

    function updateServerList() {
        fetch("/api/server").then(response => response.json().then(setServers))
    }

    useEffect(updateServerList, []);

    // For now, check every 30 seconds
    // updateServerList();
    // setInterval(updateServerList, 30e3);

    return (
        <Router>
            <Navbar bg="dark" variant="dark">
                <Container fluid>
                    <Navbar.Brand as={Link} to="/">MSMF</Navbar.Brand>
                    <Navbar.Toggle/>
                    <Navbar.Collapse role="navigation" className="justify-content-end">
                        <Nav>
                            <NavDropdown id="account-dropdown" title="Account" align="end">
                                <NavDropdown.Item as={Link} to="/invite">Invite</NavDropdown.Item>
                                <NavDropdown.Item as={Link} to="/login">Login</NavDropdown.Item>
                                <NavDropdown.Item as={Link} to="/register">Register</NavDropdown.Item>
                                <NavDropdown.Item as={Link} to="/change-password">Change Password</NavDropdown.Item>
                            </NavDropdown>
                        </Nav>
                    </Navbar.Collapse>
                </Container>
            </Navbar>

            {/* A <Switch> looks through its children <Route>s and
            renders the first one that matches the current URL. */}
            <Container fluid className="mt-2">
                <Row>
                    <div className="col-md-2">
                        <Nav variant="pills" className="flex-column sticky-top top-0">
                            <NavItem>Utilities</NavItem>
                            {servers.map(server => {
                                return <NavLink as={Link} to={`/console/${server.id}`}>
                                    Console {server.name} <Badge pill bg="info">dev</Badge>
                                </NavLink>
                            })}
                        </Nav>
                    </div>
                    <div className="col-md-10">
                        <Switch>
                            <Route path="/invite">
                                <Invite/>
                            </Route>
                            <Route path="/login">
                                <Login/>
                            </Route>
                            <Route path="/register">
                                <Register/>
                            </Route>
                            <Route path="/change-password">
                                <ChangePassword/>
                            </Route>
                            <Route path="/console/:serverId" component={DevConsole}/>
                            <Route path="/">
                                <Home/>
                            </Route>
                        </Switch>
                    </div>
                </Row>
            </Container>
        </Router>
    );
}
