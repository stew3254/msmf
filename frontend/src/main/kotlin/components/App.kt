package msmf.frontend.components

import kotlinx.html.ButtonType
import kotlinx.html.LI
import kotlinx.html.UL
import react.RBuilder
import react.RComponent
import react.RProps
import react.RState
import react.dom.RDOMBuilder
import react.dom.button
import react.dom.div
import react.dom.h1
import react.dom.li
import react.dom.nav
import react.dom.span
import react.dom.ul
import react.router.dom.browserRouter
import react.router.dom.route
import react.router.dom.routeLink
import react.router.dom.switch

class App : RComponent<RProps, RState>() {
    private fun RDOMBuilder<UL>.navItem(block: RDOMBuilder<LI>.() -> Unit) {
        li("nav-item") {
            block()
        }
    }

    private fun RDOMBuilder<UL>.navItem(title: String) = navItem { +title }

    private fun RDOMBuilder<UL>.navLink(to: String, title: String = to.replaceFirstChar(Char::titlecase)) {
        navItem {
            routeLink(to, className = "nav-link") { +title }
        }
    }

    override fun RBuilder.render() {
        browserRouter {
            nav("navbar navbar-expand-md navbar-dark bg-dark") {
                div("container-fluid") {
                    routeLink("/", className = "navbar-brand") { +"MSMF" }
                    button(classes = "navbar-toggler", type = ButtonType.button) {
                        attrs["data-bs-toggle"] = "collapse"
                        attrs["data-bs-target"] = "#main-nav-dropdown"
                        span(classes = "navbar-toggler-icon") {}
                    }
                    div("collapse navbar-collapse justify-content-end") {
                        attrs["id"] = "main-nav-dropdown"
                        ul("navbar-nav") {
                            navLink("invite")
                            navLink("login")
                            navLink("register")
                            navLink("change-password", "Change Password")
                        }
                    }
                }
            }
            div("container-fluid mt-2") {
                div("row") {
                    div("col-md-2") {
                        ul("nav nav-pills flex-column sticky-top top-0") {
                            navItem("Utilities")
                            navLink("console")
                        }
                    }
                    div("col-md-10") {
                        switch {
                            route("/invite") {
                                child(Invite::class) {}
                            }
                            route("/login") {
                                h1 { +"Login" }
                            }
                            route("/register") {
                                h1 { +"Register" }
                            }
                            route("/change-password") {
                                h1 { +"Change Password" }
                            }
                            route("/console") {
                                h1 { +"Console" }
                            }
                            route("/") {
                                h1 { +"Hello world!" }
                            }
                        }
                    }
                }
            }
        }
    }
}
