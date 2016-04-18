package ircwebchat

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"github.com/oooska/irc"
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
	client := irc.NewConnectionWrapper(ws)

	//Authenticate websocket:
	var user *iwcUser

	for user == nil {
		client.Write(irc.NewMessage("CLIENT-MESSAGE :Enter a username."))
		msg, err := client.Read()
		if err != nil {
			return
		}
		username := strings.TrimSpace(msg.String())

		client.Write(irc.NewMessage("CLIENT-MESSAGE :Enter a password."))
		msg, err = client.Read()
		if err != nil {
			return
		}
		password := strings.TrimSpace(msg.String())
		user = authenticate(username, password)

		if user != nil {
			client.Write(irc.NewMessage("CLIENT-MESSAGE :Successfully logged in."))
		} else {
			client.Write(irc.NewMessage("CLIENT-MESSAGE :Invalid username/password."))
		}
	}

	newclients := getSessionNotifier(user.username)
	if newclients == nil {
		client.Write(irc.NewMessage("Unable to find session. Closing..."))
		log.Printf("Unable to find session for %s", user.username)
		return
	}

	//Notify the client what the user's current nick is
	client.Write(irc.NewMessage("NICK " + user.profile.nick.name))
	newclients <- client

	for {
		if ws.IsServerConn() {
			time.Sleep(100 * time.Millisecond)
		} else {
			log.Println("socketHandler returning after IsServerConn returned false")
			return
		}
	}
}
