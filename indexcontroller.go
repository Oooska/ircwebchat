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
	if req.URL.Path == "/" {
		ic.get(w, req)
	} else {
		w.WriteHeader(404)
	}
}

func (ic indexController) get(w http.ResponseWriter, req *http.Request) {
	acct, err := validateCookie(w, req)
	site := viewmodels.GetSite()
	site.Title = "IRC Web Chat"
	if err == nil {
		site.Username = acct.Username()
	}
	w.Header().Add("Content-Header", "text/html")

	log.Printf("Site: %+v", site)
	ic.template.Execute(w, site)
}
