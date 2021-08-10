import * as React from "react";
import {lazy, Suspense, useEffect, useState} from "react";
import {BrowserRouter as Router, Link, Route, Switch} from "react-router-dom";
import {Badge, Container, Nav, Navbar, NavDropdown, NavItem, NavLink, Row} from "react-bootstrap";

const Home = lazy(() => import("./HomeRoute"));
const Login = lazy(() => import("./account/LoginRoute"));
const Invite = lazy(() => import("./account/InviteRoute"));
const Register = lazy(() => import("./account/RegisterRoute"));
const ChangePassword = lazy(() => import("./account/ChangePasswordRoute"));
const NewServerRoute = lazy(() => import("./server/new/NewServerRoute"));
const DevConsole = lazy(() => import("./server/DevConsoleRoute"));

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
                    <Navbar.Collapse role="navigation">
                        <Nav>
                            <NavLink as={Link} to="/server/new">New Server</NavLink>
                        </Nav>
                        <Nav className="ms-auto">
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
                                return <NavLink as={Link} to={`/console/${server.id}`} key={server.id}>
                                    Console {server.name} <Badge pill bg="info">dev</Badge>
                                </NavLink>
                            })}
                        </Nav>
                    </div>
                    <div className="col-md-10">
                        <Suspense fallback={<div>Loading...</div>}>
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
                                <Route path="/server/new">
                                    <NewServerRoute updateServerList={updateServerList}/>
                                </Route>
                                <Route path="/console/:serverId" component={DevConsole}/>
                                <Route path="/">
                                    <Home/>
                                </Route>
                            </Switch>
                        </Suspense>
                    </div>
                </Row>
            </Container>
        </Router>
    );
}
