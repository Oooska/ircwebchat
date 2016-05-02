package models

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

//IRCLogin is a simple struct containing a nick and associated password
type IRCLogin struct {
	Nick     string
	Password string
}

//Returns the settings for a specified account, if they exist
func GetSettings(a Account) (Settings, error) {
	s, err := persistenceInstance.settings(a)
	if err != nil {
		return s, err
	}
	return s, nil
}

//UpdateSettings updates the settings for the specified account
func UpdateSettings(a Account, settings Settings) (Settings, error) {
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
