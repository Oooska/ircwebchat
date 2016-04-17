package ircwebchat

//ircWebChat user
//todo: Support multiple profiles
type iwcUser struct {
	username string
	password string
	profile  serverProfile
}

//serverProfile represents the neccesary information to connect to a
//iwcUser's server with their user and nick details
type serverProfile struct {
	address  string
	username string
	realname string
	nick     login
	altnick  login
}

type login struct {
	name     string
	password string
}

func authenticate(username, password string) *iwcUser {
	user, ok := iwcUserMap[username]
	if ok {
		if user.password == password {
			return &user
		}
	}

	return nil
}

func addUser(user iwcUser) {
	iwcUserMap[user.username] = user
}

//A map of usernames to iwcUser objects... should probably define a proper type
var iwcUserMap map[string]iwcUser = make(map[string]iwcUser)
