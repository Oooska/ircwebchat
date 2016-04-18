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
		sc.settings(w, req)
	}
}

func (sc settingsController) settings(w http.ResponseWriter, req *http.Request) {
	mdlAcct, err := validateCookie(w, req)
	if err != nil {
		//Not logged in - get user out of here
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}
	mdlSettings, err := modelSettings.Settings(mdlAcct)

	server := viewmodels.GetServer()
	server.Title = "IRC Web Chat - Settings"
	server.Username = mdlAcct.Username()

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

		//TODO: Validate form
		modelSettings.UpdateSettings(mdlAcct, server.Name, server.Address, server.Port, server.SSL)
		modelSettings.UpdateLogin(mdlAcct, server.User.Nick, server.User.Password)
		modelSettings.UpdateAltLogin(mdlAcct, server.AltUser.Nick, server.AltUser.Password)
	} else if err == nil { //Grab previously saved info
		server.Name = mdlSettings.Name()
		server.Address = mdlSettings.Address()
		server.Port = mdlSettings.Port()
		server.SSL = mdlSettings.SSL()
		server.User.Nick = mdlSettings.Login().Nick
		server.User.Password = mdlSettings.Login().Password
		server.AltUser.Nick = mdlSettings.AltLogin().Nick
		server.AltUser.Password = mdlSettings.AltLogin().Password
	}
	w.Header().Add("Content-Header", "text/html")
	sc.template.Execute(w, server)
}
