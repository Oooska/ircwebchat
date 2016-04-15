package ircwebchat

import (
	"html/template"
	"log"
	"net/http"

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
	log.Print("Hitting chatcontroller.get()")
	site := viewmodels.GetSite()
	site.Title = "IRC Web Chat - Client"

	w.Header().Add("Content-Header", "text/html")
	cc.template.Execute(w, site)
}
