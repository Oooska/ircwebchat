package ircwebchat

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/oooska/ircwebchat/models"
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
	account, err := validateCookie(w, req)
	if err != nil {
		//Not logged in - get user out of here
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}
	var vsettings viewsettings

	mdlSettings, err := models.GetSettings(account)
	if req.Method == "GET" { //Get saved settings, or default
		if err == nil {
			//No settings saved yet - use defaults
			vsettings.Settings = getDefaultViewSettings()
			vsettings.Settings.Login.Nick = account.Username()
		} else {
			vsettings.Settings = mdlSettings
			vsettings.Settings.Enabled = models.ChatStarted(account)
		}
	} else if req.Method == "POST" {
		psettings := models.Settings{}
		postFormToSettings(req, &psettings)
		vsettings.Settings = psettings
		//Update settings
		//TODO: Simplify modelSettings update functions
		s, err := models.UpdateSettings(account, psettings)
		if err != nil {
			log.Printf("Trouble saving settings: %s", err.Error())
		} else {
			//Check to see if we need to start the client
			if s.Enabled && !models.ChatStarted(account) {
				err := models.StartChat(account, s)
				if err != nil {
					vsettings.ConnectError = err.Error()
					//Unable to connect, update 'Enabled' to false
					s.Enabled = false
					models.UpdateSettings(account, s)
					vsettings.Settings.Enabled = false
				}
				vsettings.Settings = s
			} else if !s.Enabled && models.ChatStarted(account) {
				models.StopChat(account)
			}
		}
	}

	vsettings.Title = "IRC Web Chat - Settings"
	vsettings.Username = account.Username()
	vsettings.Active = "Settings"

	w.Header().Add("Content-Header", "text/html")
	sc.template.Execute(w, vsettings)
}

//viewsettings represents the
type viewsettings struct {
	sitedata
	Settings     models.Settings
	ConnectError string
}

func getDefaultViewSettings() models.Settings {
	return models.Settings{Enabled: true, Name: "Freenode", Address: "irc.freenode.net", Port: 6667, SSL: false}
}

func postFormToSettings(req *http.Request, settings *models.Settings) {
	settings.Enabled = parseCheckbox("Enabled", req)
	settings.Name = req.FormValue("Name")
	settings.Address = req.FormValue("Address")
	settings.Port, _ = strconv.Atoi(req.FormValue("Port"))
	settings.SSL = parseCheckbox("SSL", req)
	settings.Login.Nick = req.FormValue("Nick")
	settings.Login.Password = req.FormValue("Password")
	settings.AltLogin.Nick = req.FormValue("AltNick")
	settings.AltLogin.Password = req.FormValue("AltPassword")
}

func parseCheckbox(field string, req *http.Request) bool {
	val := req.FormValue(field)
	return val == "on"
}
