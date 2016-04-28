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

	mdlSettings, err := modelSettings.Settings(account)
	if req.Method == "GET" { //Get saved settings, or default
		if err != nil {
			//No settings saved yet - use defaults
			vsettings = getDefaultViewSettings()
			vsettings.User.Nick = account.Username()
		} else {
			vsettings = modelSettingsToView(mdlSettings)
			vsettings.Enabled = chatManager.ChatStarted(account)
		}
	} else if req.Method == "POST" {
		vsettings = viewsettings{}
		postFormToSettings(req, &vsettings)

		//Update settings
		//TODO: Simplify modelSettings update functions
		s, err := modelSettings.UpdateSettings(account, vsettings.Enabled, vsettings.Name, vsettings.Address, vsettings.Port, vsettings.SSL, vsettings.User, vsettings.AltUser)
		if err != nil {
			log.Printf("Trouble saving settings: %s", err.Error())
		} else {
			//Check to see if we need to start the client
			if s.Enabled() && !chatManager.ChatStarted(account) {
				err := chatManager.StartChat(account, s)
				if err != nil {
					vsettings.ConnectError = err.Error()
					//Unable to connect, update 'Enabled' to false
					modelSettings.UpdateSettings(account, false, vsettings.Name, vsettings.Address, vsettings.Port, vsettings.SSL, vsettings.User, vsettings.AltUser)
					vsettings.Enabled = false
				}
			} else if !s.Enabled() && chatManager.ChatStarted(account) {
				chatManager.StopChat(account)
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
	Enabled      bool
	Name         string
	Address      string
	Port         int
	SSL          bool
	User         models.IRCLogin
	AltUser      models.IRCLogin
	ConnectError string
}

func getDefaultViewSettings() viewsettings {
	return viewsettings{Enabled: true, Name: "Freenode", Address: "irc.freenode.net", Port: 6667, SSL: false}
}

func modelSettingsToView(mdlSettings models.Settings) viewsettings {
	vs := viewsettings{}

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

func postFormToSettings(req *http.Request, settings *viewsettings) {
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

func parseCheckbox(field string, req *http.Request) bool {
	val := req.FormValue(field)
	return val == "on"
}
