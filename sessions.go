package ircwebchat

import (
	"log"

	"github.com/oooska/irc"
)

//Map usernames to channel that looks for new web clients
var sessionNewClientsMap map[string]chan<- *ircClient = make(map[string]chan<- *ircClient)

func getSessionNotifier(username string) chan<- *ircClient {
	sessionNotifier, ok := sessionNewClientsMap[username]
	if ok {
		return sessionNotifier
	} else {
		return nil
	}

}

//Find all user sessions that should be active and start them
func startUserSessions() {
	for _, user := range iwcUserMap {
		err := startSession(user)
		if err != nil {
			log.Println("Trouble starting session for ", user.username, ": ", err)
		} else {
			log.Println("Session started successfully.")
		}
	}
}

func startSession(user iwcUser) error {
	//Start the IRC connection... //TODO: Move this elsewhere.
	conn, err := irc.NewIRCConnection(user.profile.address, false)
	if err != nil {
		return err
	}

	//Channel to accept new http clients logging in
	var newClients chan *ircClient = make(chan *ircClient)

	conn.Write(irc.UserMessage(user.username, "servername", "hostname", user.profile.realname))
	conn.Write(irc.NickMessage(user.profile.nick.name))
	conn.Write(irc.NewMessage("join #gotest"))
	go ircManager(conn, newClients)
	log.Println("Starting session for ", user.username)
	sessionNewClientsMap[user.username] = newClients
	return nil
}
