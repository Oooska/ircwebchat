package controllers

import (
	"html/template"
	"net/http"

	"golang.org/x/net/websocket"
)

/* ircwebchat provides a web-based IRC client. A user can share the same IRC
session across multiple browsers.
*/

//Register mounts an entry point at / on the supplied http mux.
//staticFiles is the directory that contains the CSS and .js files
//If no mux is supplied, it will be mounted by the default http.Handler
func Register(t *template.Template, staticFiles string, mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	//Instantiate our controllers
	indexController := indexController{template: t.Lookup("index.html")}
	settingsController := settingsController{template: t.Lookup("settings.html")}
	chatController := chatController{template: t.Lookup("chat.html")}
	accountsController := accountsController{template: t.Lookup("register.html")}

	//Associate routes with our controllers
	mux.Handle("/", indexController)
	mux.Handle("/settings/", settingsController)
	mux.Handle("/chat/", chatController)
	mux.Handle("/register/", accountsController)
	mux.Handle("/login", accountsController)
	mux.Handle("/logout", accountsController)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticFiles))))
	mux.Handle("/chat/socket", websocket.Handler(webSocketHandler))

}

//sitedata is used by all pages on the site
type sitedata struct {
	Title    string
	Username string
	Active   string //Common name for the currently loaded page ('Settings', 'Chat')
}
