package ircwebchat

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/websocket"
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
	acct, err := validateCookie(w, req)
	if err != nil {
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	site := sitedata{}
	site.Title = "IRC Web Chat - Client"
	site.Active = "Chat"
	site.Username = acct.Username()

	w.Header().Add("Content-Header", "text/html")
	cc.template.Execute(w, site)
}

/*
webSocketHandler reads the sessionID from the websocket,
identifies an account and joins an ongoing chat session if one exists
//TODO: Actually manage disconnections properly.
*/
func webSocketHandler(ws *websocket.Conn) {
	//Notify the irc manager of a new websocket
	log.Println("socketHandler starting")
	defer log.Println("socketHandler exiting")

	//Client should send its sessionID as first message
	br := bufio.NewReader(ws)
	sessionID, err := br.ReadString('\n')
	sessionID = strings.TrimSpace(sessionID)
	if err != nil {
		log.Printf("Error reading from client: %s", err.Error())
		return
	}
	log.Printf("Recieved session ID: '%s' over websocket", sessionID)
	acct, err := modelSessions.Lookup(sessionID)
	if err != nil {
		ws.Write([]byte("Closing connection. Unable to find user: " + err.Error()))
		ws.Close()
		return
	}

	err = chatManager.JoinChat(acct, sessionID, ws)
	ws.Write([]byte("Error: " + err.Error()))
	ws.Close()
}
