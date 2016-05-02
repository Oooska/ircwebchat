package controllers

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/oooska/ircwebchat/chat"
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
	//Client should send its sessionID as first message
	sc := bufio.NewScanner(ws)
	sc.Scan()
	if sc.Err() != nil {
		log.Printf("Error reading from client: %s", sc.Err().Error())
		return
	}
	sessionID := strings.TrimSpace(sc.Text())
	acct, err := chat.LookupSession(sessionID)
	if err != nil {
		ws.Write([]byte("Closing connection. Unable to find user: " + err.Error()))
		ws.Close()
		return
	}

	//Join the irc session that is in progress
	err = chat.JoinChat(acct, sessionID, ws)
	ws.Write([]byte("Error: " + err.Error()))
	ws.Close()
}
