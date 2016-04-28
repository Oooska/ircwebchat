package models

func NewSettingsManager() SettingsManager {
	sm := settingsMgr{}

	//Dummy data
	/*sm.settings["goirctest"] = settings{
		enabled:  true,
		name:     "Freenode",
		address:  "irc.freenode.net",
		ssl:      false,
		port:     6667,
		login:    IRCLogin{Nick: "goirctest", Password: ""},
		altlogin: IRCLogin{Nick: "goirctest_", Password: ""},
	}*/

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
	UpdateSettings(acct Account, enabled bool, name, address string, port int, ssl bool, login IRCLogin, altlogin IRCLogin) (Settings, error)
}

type settingsMgr struct {
}

//Returns the settings for a specified account, if they exist
func (sets settingsMgr) Settings(a Account) (Settings, error) {
	s, err := persistenceInstance.settings(a)
	if err != nil {
		return s, err
	}
	return s, nil
}

//UpdateSettings updates the settings for the specified account
func (sets settingsMgr) UpdateSettings(a Account, enabled bool, name, address string, port int, ssl bool, login IRCLogin, altlogin IRCLogin) (Settings, error) {
	s := settings{accountid: a.ID(), enabled: enabled, name: name, address: address, ssl: ssl, port: port, login: login, altlogin: altlogin}
	err := persistenceInstance.saveSettings(s)
	return s, err
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

func newsettings(accountid int64, enabled bool, name, address string, port int, ssl bool, login IRCLogin, altlogin IRCLogin) settings {
	return settings{
		accountid: accountid,
		enabled:   enabled,
		name:      name,
		address:   address,
		port:      port,
		ssl:       ssl,
		login:     login,
		altlogin:  altlogin,
	}
}

func newirclogin(nick, password string) IRCLogin {
	return IRCLogin{Nick: nick, Password: password}
}

//IRCLogin is a simple struct containing a nick and associated password
type IRCLogin struct {
	Nick     string
	Password string
}

type settings struct {
	accountid int64
	enabled   bool
	name      string
	address   string
	port      int
	ssl       bool
	login     IRCLogin
	altlogin  IRCLogin
}

//Enabled returns true if the IRC server should be connected
func (s settings) Enabled() bool {
	return s.enabled
}

//Name returns the friendly name of the IRC server (e.g. 'freenode')
func (s settings) Name() string {
	return s.name
}

//Address returns the address of the IRC server
func (s settings) Address() string {
	return s.address
}

//Port returns the port of the IRC server
func (s settings) Port() int {
	return s.port
}

//SSL returns true if SSL is enabled between this server and the irc server
func (s settings) SSL() bool {
	return s.ssl
}

//Login returns the login details for the primary nick
func (s settings) Login() IRCLogin {
	return s.login
}

//AltLogin returns the login details for the alternate nick
func (s settings) AltLogin() IRCLogin {
	return s.altlogin
}
