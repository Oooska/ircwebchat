package ircwebchat

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/websocket"
)

/* ircwebchat provides a web-basede IRC client. A user can share the same IRC
session across multiple browsers.

Still in early development stages.

TODO: Currently only sends data to clients. Need to listen to IRCCLients and pass info on to other clients and server
*/

//Register mounts an entry point at / on the supplied http mux.
//If no mux is supplied, it will be mounted by the default http.Handler
//TODO: We currently start the connection to the IRC server here. This
// should be abstracted away.
func Register(mux *http.ServeMux) {
	log.Println("Register() called...")
	templates = populateTemplates()

	if mux != nil {
		mux.Handle("/", http.HandlerFunc(templateHandler))
		mux.Handle("/static/", http.HandlerFunc(serveResource))
		mux.Handle("/chat/socket", websocket.Handler(webSocketHandler))
	} else {
		http.Handle("/", http.HandlerFunc(templateHandler))
		http.Handle("/static/", http.HandlerFunc(serveResource))
		http.Handle("/chat/socket", websocket.Handler(webSocketHandler))
	}

	user := iwcUser{
		username: "goirctest",
		password: "password",
		profile: serverProfile{
			address: "irc.freenode.net:6667",
			nick: login{
				name:     "goirctest",
				password: "",
			},
			realname: "go-get-real",
			altnick: login{
				name:     "goirctest_",
				password: "",
			},
		},
	}
	user2 := iwcUser{
		username: "goirctest2",
		password: "password",
		profile: serverProfile{
			address: "irc.freenode.net:6667",
			nick: login{
				name:     "goirctest2",
				password: "",
			},
			altnick: login{
				name:     "goirctest2_",
				password: "",
			},
			realname: "go-get-real",
		},
	}

	addUser(user)
	addUser(user2)

	log.Print("About to start user sessions...")
	go startUserSessions()
	log.Print("User sessions started.")
}

var templates *template.Template

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

func templateHandler(w http.ResponseWriter, req *http.Request) {
	requestedFile := req.URL.Path[1:]
	if requestedFile == "" || requestedFile[len(requestedFile)-1] == '/' {
		requestedFile += "index"
	}
	template := templates.Lookup(requestedFile + ".html")

	if template != nil {
		template.Execute(w, nil)
	} else {
		w.WriteHeader(404)
	}
}
