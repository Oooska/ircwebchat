package viewmodels

//TODO: Allow for multiple servers per account
//type Settings struct {
//	Servers []Server
//}

type Settings struct {
	Site
	Enabled bool
	Name    string
	Address string
	Port    int
	SSL     bool
	User    IRCUser
	AltUser IRCUser
}

type IRCUser struct {
	Nick     string
	Password string
}

func GetSettings() Settings {
	return Settings{}
}
