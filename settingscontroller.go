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

	settings := viewmodels.GetSettings()
	settings.Title = "IRC Web Chat - Settings"
	settings.Username = mdlAcct.Username()

	if req.Method == "POST" {
		settings.Enabled = parseCheckbox("Enabled", req)
		settings.Name = req.FormValue("Name")
		settings.Address = req.FormValue("Address")
		settings.Port, _ = strconv.Atoi(req.FormValue("Port"))
		settings.SSL = parseCheckbox("SSL", req)
		settings.User.Nick = req.FormValue("Nick")
		settings.User.Password = req.FormValue("Password")
		settings.AltUser.Nick = req.FormValue("AltNick")
		settings.AltUser.Password = req.FormValue("AltPassword")
		log.Printf("Posted settings: %+v", settings)
		//TODO: Validate form

		modelSettings.UpdateSettings(mdlAcct, settings.Enabled, settings.Name, settings.Address, settings.Port, settings.SSL)
		modelSettings.UpdateLogin(mdlAcct, settings.User.Nick, settings.User.Password)
		s := modelSettings.UpdateAltLogin(mdlAcct, settings.AltUser.Nick, settings.AltUser.Password)

		//Check to see if we need to start the client (enable toggled)
		if s.Enabled() && (err == nil || (err != nil && !mdlSettings.Enabled())) {
			log.Printf("This is where I should send some kind of signal to tell the chatmanager to connect...")
			chatManager.StartSession(mdlAcct, s)
		} else if err == nil && mdlSettings.Enabled() && !s.Enabled() {
			log.Printf("This is where I should send some kind of signal to tell the chatmanager to disconnect...")
			chatManager.StopSession(mdlAcct)
		}

	} else if err == nil { //Grab previously saved info
		settings.Enabled = mdlSettings.Enabled()
		settings.Name = mdlSettings.Name()
		settings.Address = mdlSettings.Address()
		settings.Port = mdlSettings.Port()
		settings.SSL = mdlSettings.SSL()
		settings.User.Nick = mdlSettings.Login().Nick
		settings.User.Password = mdlSettings.Login().Password
		settings.AltUser.Nick = mdlSettings.AltLogin().Nick
		settings.AltUser.Password = mdlSettings.AltLogin().Password
	}
	w.Header().Add("Content-Header", "text/html")
	sc.template.Execute(w, settings)
}

func parseCheckbox(field string, req *http.Request) bool {
	val := req.FormValue(field)
	log.Printf("form['%s']=%s (evaluates to %v)", field, val, val == "on")
	return val == "on"
}
