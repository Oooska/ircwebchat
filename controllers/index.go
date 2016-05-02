package controllers

import (
	"html/template"
	"net/http"
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
	site := sitedata{}
	site.Title = "IRC Web Chat"
	site.Active = "Index"
	if err == nil {
		site.Username = acct.Username()
	}
	w.Header().Add("Content-Header", "text/html")
	ic.template.Execute(w, site)
}
