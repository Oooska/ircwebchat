package controllers

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/oooska/ircwebchat/chat"
	"golang.org/x/net/websocket"
)

type chatController struct {
	template *template.Template
}

func (cc chatController) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	acct, err := validateCookie(w, req)
	if err != nil {
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	sPath := strings.Split(req.URL.Path[1:], "/")
	if len(sPath) == 2 && sPath[0] == "chat" {
		// /chat - last value in sPath is an empty string from strings.Split
		cc.get(acct, w, req)
	} else if len(sPath) == 3 && sPath[0] == "chat" && sPath[2] != "" {
		// /chat/#channel/timestamp
		cc.history(acct, w, req, sPath[1], sPath[2])
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Nothing to see here"))
	}
}

func (cc chatController) get(acct chat.Account, w http.ResponseWriter, req *http.Request) {
	site := sitedata{}
	site.Title = "IRC Web Chat - Client"
	site.Active = "Chat"
	site.Username = acct.Username()

	w.Header().Add("Content-Header", "text/html")
	cc.template.Execute(w, site)
}

//Grabs the last 200 messages from before the specified timestamp for the specified channel
func (cc chatController) history(acct chat.Account, w http.ResponseWriter, req *http.Request, ch string, timestamp string) {
	log.Printf("Looking up history for %s before %s", ch, timestamp)
	tsint, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		w.Write([]byte(err.Error()))
		return
	}

	messages, err := chat.ChatLogs(acct, ch, time.Unix(tsint-1, 0))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Content-Header", "text")
	for _, msg := range messages {
		t := strconv.FormatInt(msg.Timestamp().Unix(), 10)
		w.Write([]byte(t))
		w.Write([]byte(" "))
		w.Write([]byte(msg.String()))
		w.Write([]byte("\r\n"))
	}
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
