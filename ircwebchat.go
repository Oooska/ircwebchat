package ircwebchat

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/oooska/ircwebchat/models"

	"golang.org/x/net/websocket"
)

/* ircwebchat provides a web-basede IRC client. A user can share the same IRC
session across multiple browsers.

Still in early development stages.

TODO: Currently only sends data to clients. Need to listen to IRCCLients and pass info on to other clients and server
*/

var templates *template.Template
var modelAccounts = models.NewAccounts()
var modelSessions = models.NewSessions()
var modelSettings = models.NewSettingsManager()
var chatManager = models.NewChatManager()

//Register mounts an entry point at / on the supplied http mux.
//If no mux is supplied, it will be mounted by the default http.Handler
//TODO: We currently start the connection to the IRC server here. This
// should be abstracted away.
func Register(t *template.Template, mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}
	log.Println("Register() called...")
	templates = t

	ic := indexController{template: templates.Lookup("index.html")}
	sc := settingsController{template: templates.Lookup("settings.html")}
	cc := chatController{template: templates.Lookup("chat.html")}
	ac := accountsController{template: templates.Lookup("register.html")}

	mux.Handle("/", ic)
	mux.Handle("/settings", sc)
	mux.Handle("/chat", cc)

	mux.Handle("/register", ac)
	mux.Handle("/login", ac)
	mux.Handle("/logout", ac)

	mux.Handle("/static/", http.HandlerFunc(serveResource))
	mux.Handle("/chat/socket", websocket.Handler(webSocketHandler))

	log.Print("About to start user sessions...")
	chatManager.StartChats(modelAccounts, modelSettings)
	log.Print("User sessions started.")
}

func serveResource(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path[1:]
	var contentType string
	if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(path, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(path, ".html") {
		contentType = "text/html"
	} else if strings.HasSuffix(path, ".js") {
		contentType = "text/javascript"
	} else {
		contentType = "text/plain"
	}

	f, err := os.Open(path)
	if err != nil {
		w.WriteHeader(404)
	}
	defer f.Close()
	br := bufio.NewReader(f)
	w.Header().Add("Content-Type", contentType)

	br.WriteTo(w)
}
