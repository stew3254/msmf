package msmf.frontend.components

import kotlinx.browser.window
import kotlinx.coroutines.MainScope
import kotlinx.coroutines.await
import kotlinx.coroutines.launch
import kotlinx.html.ButtonType
import kotlinx.html.js.onClickFunction
import react.RBuilder
import react.RComponent
import react.RProps
import react.RState
import react.dom.attrs
import react.dom.b
import react.dom.button
import react.dom.h1
import react.dom.p
import react.setState

data class InviteState(var generatedCode: String) : RState

class Invite : RComponent<RProps, InviteState>() {
    private val mainScope = MainScope()

    init {
        state = InviteState("No Code Generated!")
    }

    private data class InviteResponse(val code: String)

    private suspend fun fetchInviteCode(): String {
        val response = window.fetch("/api/refer/get")
                .await()
                .json()
                .await()
        return (response as InviteResponse).code
    }

    override fun RBuilder.render() {
        h1 { +"Generate Invite Code" }
        p {
            +"Your invite code is "
            b {
                +state.generatedCode
            }
        }
        button(classes = "btn btn-primary", type = ButtonType.button) {
            +"Generate"
            attrs {
                onClickFunction = {
                    setState(InviteState("loading..."))
                    mainScope.launch {
                        fetchInviteCode().let {
                            setState { generatedCode = it }
                        }
                    }
                }
            }
        }
    }
}
