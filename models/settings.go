package models

import "errors"

func NewSettingsManager() SettingsManager {
	sm := settingsMgr{settings: make(map[string]settings)}

	//Dummy data
	sm.settings["goirctest"] = settings{
		enabled:  true,
		name:     "Freenode",
		address:  "irc.freenode.net",
		ssl:      false,
		port:     6667,
		login:    IRCLogin{Nick: "goirctest", Password: ""},
		altlogin: IRCLogin{Nick: "goirctest_", Password: ""},
	}

	/*sm.settings["goirctest2"] = settings{
		enabled:  false,
		name:     "Freenode",
		address:  "irc.freenode.net",
		ssl:      false,
		port:     6667,
		login:    IRCLogin{Nick: "goirctest2", Password: ""},
		altlogin: IRCLogin{Nick: "goirctest2_", Password: ""},
	}*/

	return sm
}

//SettingsManager keeps track of settings for each account
type SettingsManager interface {
	Settings(Account) (Settings, error)
	UpdateSettings(acct Account, enabled bool, name, address string, port int, ssl bool) Settings
	UpdateLogin(acct Account, nick, password string) Settings
	UpdateAltLogin(acct Account, nick, password string) Settings
}

type settingsMgr struct {
	settings map[string]settings
}

func (sets settingsMgr) Settings(a Account) (Settings, error) {
	s, ok := sets.settings[a.Username()]
	if !ok {
		return s, errors.New("No settings found")
	}
	return s, nil
}

func (sets settingsMgr) UpdateSettings(a Account, enabled bool, name, address string, port int, ssl bool) Settings {
	s := settings{enabled: enabled, name: name, address: address, ssl: ssl, port: port}
	sets.settings[a.Username()] = s
	return s
}

func (sets settingsMgr) UpdateLogin(a Account, nick, password string) Settings {
	settings, ok := sets.settings[a.Username()]
	if ok {
		settings.login = IRCLogin{Nick: nick, Password: password}
		sets.settings[a.Username()] = settings
	}
	return settings
}

func (sets settingsMgr) UpdateAltLogin(a Account, nick, password string) Settings {
	settings, ok := sets.settings[a.Username()]
	if ok {
		settings.altlogin = IRCLogin{Nick: nick, Password: password}
		sets.settings[a.Username()] = settings
	}
	return settings
}

//Settings represents the information required to connect to an IRC server
type Settings interface {
	Enabled() bool
	Name() string
	Address() string
	Port() int
	SSL() bool
	Login() IRCLogin
	AltLogin() IRCLogin
}

//IRCLogin is a simple struct containing a nick and associated password
type IRCLogin struct {
	Nick     string
	Password string
}

type settings struct {
	enabled  bool
	name     string
	address  string
	port     int
	ssl      bool
	login    IRCLogin
	altlogin IRCLogin
}

func (s settings) Enabled() bool {
	return s.enabled
}

func (s settings) Name() string {
	return s.name
}

func (s settings) Address() string {
	return s.address
}

func (s settings) Port() int {
	return s.port
}

func (s settings) SSL() bool {
	return s.ssl
}

func (s settings) Login() IRCLogin {
	return s.login
}

func (s settings) AltLogin() IRCLogin {
	return s.altlogin
}
