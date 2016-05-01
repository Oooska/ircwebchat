package models

func NewSettingsManager() SettingsManager {
	sm := settingsMgr{}
	return sm
}

//SettingsManager keeps track of settings for each account
type SettingsManager interface {
	Settings(Account) (Settings, error)
	UpdateSettings(acct Account, settings Settings) (Settings, error)
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
func (sets settingsMgr) UpdateSettings(a Account, settings Settings) (Settings, error) {
	settings.accountid = a.ID()
	err := persistenceInstance.saveSettings(settings)
	return settings, err
}

func newsettings(accountid int64, enabled bool, name, address string, port int, ssl bool, login IRCLogin, altlogin IRCLogin) Settings {
	return Settings{
		accountid: accountid,
		Enabled:   enabled,
		Name:      name,
		Address:   address,
		Port:      port,
		SSL:       ssl,
		Login:     login,
		AltLogin:  altlogin,
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

type Settings struct {
	accountid int64
	Enabled   bool
	Name      string
	Address   string
	Port      int
	SSL       bool
	Login     IRCLogin
	AltLogin  IRCLogin
}
