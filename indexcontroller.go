package ircwebchat

import (
	"html/template"
	"log"
	"net/http"

	"github.com/oooska/ircwebchat/viewmodels"
)

type indexController struct {
	template *template.Template
}

func (ic indexController) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Print("Recieved request for /... in indexController")
	if req.Method == "GET" && req.URL.Path == "/" {
		ic.get(w, req)
	} else {
		w.WriteHeader(404)
	}
}

func (ic indexController) get(w http.ResponseWriter, req *http.Request) {
	site := viewmodels.GetSite()
	site.Title = "IRC Web Chat"

	w.Header().Add("Content-Header", "text/html")

	log.Printf("Site: %+v", site)
	ic.template.Execute(w, site)
}
