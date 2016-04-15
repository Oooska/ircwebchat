package ircwebchat

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/oooska/ircwebchat/viewmodels"
)

type settingsController struct {
	template *template.Template
}

func (sc settingsController) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/settings" {
		w.WriteHeader(404)
		return
	}

	if req.Method == "GET" || req.Method == "POST" {
		sc.get(w, req)
	}
}

func (sc settingsController) get(w http.ResponseWriter, req *http.Request) {

	server := viewmodels.GetServer()
	server.Title = "IRC Web Chat - Settings"
	if req.Method == "POST" {
		server.Name = req.FormValue("Name")
		server.Address = req.FormValue("Address")
		port, _ := strconv.Atoi(req.FormValue("Port"))
		server.Port = port
		log.Printf("ssl form value: %s", req.FormValue("SSL"))
		ssl, _ := strconv.ParseBool(req.FormValue("SSL"))
		server.SSL = ssl
		server.User.Nick = req.FormValue("Nick")
		server.User.Password = req.FormValue("Password")
		server.AltUser.Nick = req.FormValue("AltNick")
		server.AltUser.Password = req.FormValue("AltPassword")

	}

	log.Printf("Serving up /settings with server: %+v", server)

	w.Header().Add("Content-Header", "text/html")
	sc.template.Execute(w, server)
}
