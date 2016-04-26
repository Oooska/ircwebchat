package ircwebchat

import (
	"html/template"
	"log"
	"net/http"

	"github.com/oooska/ircwebchat/models"

	"golang.org/x/net/websocket"
)

/* ircwebchat provides a web-based IRC client. A user can share the same IRC
session across multiple browsers.

Still in early development stages.
*/

var templates *template.Template
var modelAccounts = models.NewAccounts()
var modelSessions = models.NewSessions()
var modelSettings = models.NewSettingsManager()
var chatManager = models.NewChatManager()

//Register mounts an entry point at / on the supplied http mux.
//staticFiles is the directory that contains the CSS and .js files
//If no mux is supplied, it will be mounted by the default http.Handler
func Register(t *template.Template, staticFiles string, mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}
	log.Println("Register() called...")
	templates = t

	indexController := indexController{template: templates.Lookup("index.html")}
	settingsController := settingsController{template: templates.Lookup("settings.html")}
	chatController := chatController{template: templates.Lookup("chat.html")}
	accountsController := accountsController{template: templates.Lookup("register.html")}

	mux.Handle("/", indexController)
	mux.Handle("/settings", settingsController)
	mux.Handle("/chat", chatController)

	mux.Handle("/register", accountsController)
	mux.Handle("/login", accountsController)
	mux.Handle("/logout", accountsController)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticFiles))))
	mux.Handle("/chat/socket", websocket.Handler(webSocketHandler))

	log.Print("About to start user sessions...")
	chatManager.StartChats(modelAccounts, modelSettings)
	log.Print("User sessions started.")
}

//sitedata is used by all pages on the site
type sitedata struct {
	Title    string
	Username string
	Active   string
}
