package msmf.frontend

import react.dom.render
import kotlinx.browser.document
import kotlinx.browser.window
import msmf.frontend.components.App

fun main() {
    kotlinext.js.require("bootstrap/dist/css/bootstrap.min.css")
    kotlinext.js.require("bootstrap/dist/js/bootstrap.bundle.js")
    window.onload = {
        render(document.getElementById("root")) {
            child(App::class) {}
        }
    }
}
