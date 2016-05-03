package controllers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/oooska/ircwebchat/chat"
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

	if req.Method == "GET" { //Get saved settings, or default
		mdlSettings, err := chat.GetSettings(account)
		if err != nil {
			//No settings saved yet - use defaults
			vsettings.Settings = getDefaultViewSettings()
			vsettings.Settings.Login.Nick = account.Username()
		} else {
			vsettings.Settings = mdlSettings
			vsettings.Settings.Enabled = chat.ChatStarted(account)
		}
	} else if req.Method == "POST" {
		psettings := chat.Settings{}
		postFormToSettings(req, &psettings)
		vsettings.Settings = psettings
		_, err := chat.UpdateSettings(account, psettings)
		if err != nil {
			log.Printf("Trouble saving settings: %s", err.Error())
			vsettings.ConnectError = err.Error()
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
	Settings     chat.Settings
	ConnectError string
}

func getDefaultViewSettings() chat.Settings {
	return chat.Settings{Enabled: true, Name: "Freenode", Address: "irc.freenode.net", Port: 6667, SSL: false}
}

func postFormToSettings(req *http.Request, settings *chat.Settings) {
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
