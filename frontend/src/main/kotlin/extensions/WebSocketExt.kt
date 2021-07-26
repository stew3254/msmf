package msmf.frontend.extensions

import kotlinx.browser.window
import org.w3c.dom.WebSocket

fun WebSocket.Companion.connectRelative(url: String): WebSocket {
    val loc = window.location
    val protocol = if (loc.protocol == "https:") "wss:" else "ws:"
    val host = loc.host
    val path = (if (url.startsWith("/")) "" else loc.pathname) + url
    return WebSocket("$protocol//$host$path")
}

fun WebSocket.Companion.connectAPI(vararg path: Any) = connectRelative("/" + path.joinToString("/"))
