package ircwebchat

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/oooska/ircwebchat/viewmodels"
)

type chatController struct {
	template *template.Template
}

func (cc chatController) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/chat" {
		cc.get(w, req)
	} else {
		w.WriteHeader(404)
	}
}

func (cc chatController) get(w http.ResponseWriter, req *http.Request) {
	site := viewmodels.GetSite()
	acct, err := validateCookie(w, req)
	site.Title = "IRC Web Chat - Client"
	if err != nil {
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}
	site.Username = acct.Username()

	w.Header().Add("Content-Header", "text/html")
	cc.template.Execute(w, site)
}

/*
The socketHandler for the websocket connection.
Accepts the websocket, hands it off through the socketChan, and
waits until the socket is closed before exiting the function.
//TODO: Actually manage disconnections properly.
*/
func webSocketHandler(ws *websocket.Conn) {
	//Notify the irc manager of a new websocket
	log.Println("socketHandler starting")
	defer log.Println("socketHandler exiting")

	//Client should send its sessionID as first message
	br := bufio.NewReader(ws)
	sessionID, err := br.ReadString('\n')
	if err != nil {
		log.Printf("Error reading from client: %s", err.Error())
		return
	}
	log.Printf("Recieved session ID: '%s' over websocket", sessionID)
	user, err := modelSessions.Lookup(strings.TrimSpace(sessionID))
	if err != nil {
		ws.Write([]byte("Closing connection. Error: " + err.Error()))
		ws.Close()
		return
	}

	err = chatManager.JoinChat(user, sessionID, ws)
	ws.Write([]byte("Error: " + err.Error()))
	ws.Close()
}
