package ircwebchat

import (
	"log"

	"github.com/oooska/irc"
)

//Map usernames to channel that looks for new web clients
var sessionNewClientsMap = make(map[string]chan<- irc.Conn)

func getSessionNotifier(username string) chan<- irc.Conn {
	sessionNotifier, ok := sessionNewClientsMap[username]
	if ok {
		return sessionNotifier
	}

	return nil
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
	log.Printf("Starting session for user %s", user.username)
	//Start the IRC connection... //TODO: Move this elsewhere.
	conn, err := irc.NewConnection(user.profile.address, false)
	if err != nil {
		return err
	}

	//Channel to accept new http clients logging in
	newClients := make(chan irc.Conn)

	conn.Write(irc.UserMessage(user.username, "servername", "hostname", user.profile.realname))
	conn.Write(irc.NickMessage(user.profile.nick.name))
	conn.Write(irc.NewMessage("join #gotest"))
	go ircManager(conn, newClients)
	sessionNewClientsMap[user.username] = newClients

	log.Printf("Session for user %s started with no errors.", user.username)
	return nil
}