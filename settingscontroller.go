package ircwebchat

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/oooska/ircwebchat/models"
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
	log.Printf("Account: %+v", mdlAcct)
	vmSettings := viewmodels.GetEmptySettings()

	log.Printf("Looking up existing settings...")
	mdlSettings, err := modelSettings.Settings(mdlAcct)
	if req.Method == "GET" { //Get saved settings, or default
		if err != nil {
			log.Printf("No settings found. Getting default settings.")
			vmSettings = viewmodels.GetDefaultSettings()
		} else {
			log.Printf("Found existing settings. Loading...")
			vmSettings = modelSettingsToView(mdlSettings)
			vmSettings.Enabled = chatManager.ChatStarted(mdlAcct)
		}
	} else if req.Method == "POST" {
		postFormToSettings(req, &vmSettings)
		log.Printf("Posted settings: %+v", vmSettings)
		//TODO: Validate form
		modelSettings.UpdateSettings(mdlAcct, vmSettings.Enabled, vmSettings.Name, vmSettings.Address, vmSettings.Port, vmSettings.SSL)
		modelSettings.UpdateLogin(mdlAcct, vmSettings.User.Nick, vmSettings.User.Password)
		s := modelSettings.UpdateAltLogin(mdlAcct, vmSettings.AltUser.Nick, vmSettings.AltUser.Password)
		log.Printf("Managed to update settings without crashing...")
		//Check to see if we need to start the client
		if s.Enabled() && !chatManager.ChatStarted(mdlAcct) {
			log.Printf("Starting chat for %s", mdlAcct.Username())
			err := chatManager.StartChat(mdlAcct, s)
			if err != nil {
				vmSettings.ConnectError = err.Error()
				//Unable to connect, update 'Enabled' to false
				modelSettings.UpdateSettings(mdlAcct, false, vmSettings.Name, vmSettings.Address, vmSettings.Port, vmSettings.SSL)
				vmSettings.Enabled = false
			}
		} else if !s.Enabled() && chatManager.ChatStarted(mdlAcct) {
			log.Printf("Disconnecting chat for %s", mdlAcct.Username())
			chatManager.StopChat(mdlAcct)
		}
	}

	vmSettings.Title = "IRC Web Chat - Settings"
	vmSettings.Username = mdlAcct.Username()

	w.Header().Add("Content-Header", "text/html")
	sc.template.Execute(w, vmSettings)
}

func parseCheckbox(field string, req *http.Request) bool {
	val := req.FormValue(field)
	return val == "on"
}

func modelSettingsToView(mdlSettings models.Settings) viewmodels.Settings {
	vs := viewmodels.GetEmptySettings()

	vs.Enabled = mdlSettings.Enabled()
	vs.Name = mdlSettings.Name()
	vs.Address = mdlSettings.Address()
	vs.Port = mdlSettings.Port()
	vs.SSL = mdlSettings.SSL()
	vs.User.Nick = mdlSettings.Login().Nick
	vs.User.Password = mdlSettings.Login().Password
	vs.AltUser.Nick = mdlSettings.AltLogin().Nick
	vs.AltUser.Password = mdlSettings.AltLogin().Password
	return vs
}

func postFormToSettings(req *http.Request, settings *viewmodels.Settings) {
	settings.Enabled = parseCheckbox("Enabled", req)
	settings.Name = req.FormValue("Name")
	settings.Address = req.FormValue("Address")
	settings.Port, _ = strconv.Atoi(req.FormValue("Port"))
	settings.SSL = parseCheckbox("SSL", req)
	settings.User.Nick = req.FormValue("Nick")
	settings.User.Password = req.FormValue("Password")
	settings.AltUser.Nick = req.FormValue("AltNick")
	settings.AltUser.Password = req.FormValue("AltPassword")
}
