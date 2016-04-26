package viewmodels

//TODO: Allow for multiple servers per account
//type Settings struct {
//	Servers []Server
//}

type Settings struct {
	Site
	Enabled      bool
	Name         string
	Address      string
	Port         int
	SSL          bool
	User         IRCUser
	AltUser      IRCUser
	ConnectError string
}

type IRCUser struct {
	Nick     string
	Password string
}

func GetEmptySettings() Settings {
	return Settings{}
}

func GetDefaultSettings() Settings {
	return Settings{Enabled: true, Name: "Freenode", Address: "irc.freenode.net", Port: 6667, SSL: false}
}
