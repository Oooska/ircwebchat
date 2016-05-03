package chat

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
//StartChat or StopChat will be called if the account becomes enabled/disabled
func UpdateSettings(a Account, newsettings Settings) (Settings, error) {
	initSettings, err := persistenceInstance.settings(a)
	if err == nil && initSettings.Enabled && !newsettings.Enabled {
		StopChat(a)
	} else if (err == nil && !initSettings.Enabled && newsettings.Enabled) ||
		(err != nil && newsettings.Enabled) {
		StartChat(a, newsettings)
	}

	newsettings.accountid = a.ID()
	err = persistenceInstance.saveSettings(newsettings)
	return newsettings, err
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
