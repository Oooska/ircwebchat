package ircwebchat

import (
	"html/template"
	"net/http"

	"github.com/oooska/irc"
	"golang.org/x/net/websocket"
)

/* ircwebchat provides a web-basede IRC client. A user can share the same IRC
session across multiple browsers.

Still in early development stages.

TODO: Currently only sends data to clients. Need to listen to IRCCLients and pass info on to other clients and server
*/

//Socket to listen to new websockets on; should be abstracted away
var newClients chan *ircClient = make(chan *ircClient)

//Register mounts an entry point at /chat/ on the supplied http mux.
//TODO: We currently start the connection to the IRC server here. This
// should be abstracted away.
func Register(mux http.ServeMux) {
	mux.HandleFunc("/chat/", rootHandler)
	mux.Handle("/chat/socket", websocket.Handler(webSocketHandler))

	//Start the IRC connection... //TODO: Move this elsewhere.
	addr := "irc.freenode.net:6667"
	username := "toodles"
	nick := "oooska_test"
	servername := "server_name"
	realname := "test_client"

	conn, err := irc.NewIRCConnection(addr, false)
	if err != nil {
		panic(err)
	}

	conn.Write(irc.UserMessage(username, addr, servername, realname))
	conn.Write(irc.NickMessage(nick))
	conn.Write(irc.NewMessage("join #go_test"))
	go ircManager(conn)
}

/*  Rest of the code below is borrowed/modified from:
https://code.google.com/p/go/source/browse/2012/chat/both/html.go?repo=talks&r=3a315071e5e93d9f0f33e675eae029779b43a3ec
*/
func rootHandler(w http.ResponseWriter, r *http.Request) {
	rootTemplate.Execute(w, r.Host)
}

var rootTemplate = template.Must(template.New("root").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script>

var input, output, websocket;

function showMessage(m) {
        var p = document.createElement("p");
        p.innerHTML = m;
        output.appendChild(p);
}

function onMessage(e) {
		console.log("onMessage(",e,")")
        showMessage(e.data);
}

function onClose() {
        showMessage("Connection closed.");
}

function sendMessage() {
        var m = input.value;
        input.value = "";
        websocket.send(m + "\n");
		console.log("showMessage(",m,")")
        showMessage(m);
}

function onKey(e) {
        if (e.keyCode == 13) {
                sendMessage();
        }
}

function init() {
        input = document.getElementById("input");
        input.addEventListener("keyup", onKey, false);

        output = document.getElementById("output");

        websocket = new WebSocket("ws://{{.}}/chat/socket");
        websocket.onmessage = onMessage;
        websocket.onclose = onClose;
}

window.addEventListener("load", init, false);

</script>
</head>
<body>
<input id="input" type="text">
<div id="output"></div>
</body>
</html>
`))
