package msmf.frontend.components

import kotlinx.html.js.onChangeFunction
import kotlinx.html.js.onClickFunction
import kotlinx.html.js.onKeyPressFunction
import msmf.frontend.extensions.connectAPI
import org.w3c.dom.HTMLInputElement
import org.w3c.dom.WebSocket
import react.RBuilder
import react.RComponent
import react.RProps
import react.RState
import react.dom.attrs
import react.dom.button
import react.dom.div
import react.dom.h1
import react.dom.input
import react.dom.pre
import react.setState

data class DevConsoleState(var userInput: String = "", var consoleLog: String = "") : RState

class DevConsole : RComponent<RProps, DevConsoleState>() {
    var webSocket: WebSocket? = null

    init {
        state = DevConsoleState()
    }

    private fun appendLog(message: String) {
        setState {
            consoleLog += message + "\n"
        }
    }

    private fun sendCommand() {
        console.log(state.userInput)
//        webSocket?.send(state.userInput)
        setState {
            consoleLog += userInput + "\n"
            userInput = ""
        }
    }

    override fun RBuilder.render() {
        h1 {
            +"Dev Console"
        }
        pre("bg-dark text-light p-3 rounded h-100") {
            +state.consoleLog
        }
        div("input-group mt-3") {
            input(classes = "form-control") {
                attrs {
                    value = state.userInput
                    onChangeFunction = {
                        val target = it.target as HTMLInputElement
                        setState {
                            userInput = target.value
                        }
                    }
                    onKeyPressFunction = {
                        val event = it.asDynamic()
                        if (event.key == "Enter") {
                            sendCommand()
                        }
                    }
                }
            }
            button(classes = "btn btn-primary") {
                +"Submit"
                attrs {
                    onClickFunction = {
                        sendCommand()
                    }
                }
            }
        }
    }

    override fun componentDidMount() {
        webSocket = WebSocket.connectAPI("server", 1)
        appendLog("Opening ${webSocket?.url}")

        webSocket?.onopen = {
            appendLog("Socket open!")
            console.log(it)
        }

        webSocket?.onmessage = {
            it.data?.toString()?.let { data ->
                appendLog(data)
            }
            console.log(it)
        }

        webSocket?.onerror = {
            console.error(it)
        }

        webSocket?.onclose = {
            console.log(it)
        }
    }

    override fun componentWillUnmount() {
        webSocket?.close()
    }
}
