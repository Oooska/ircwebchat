package ircwebchat

import (
	"html/template"
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
