package models

import (
	"fmt"
	"log"

	"github.com/oooska/irc"
)

//ChatManager keeps track of active connections to IRC servers
//This interface will be cleaned up shortly
type ChatManager interface {
	SessionNotifier(Account) chan<- irc.Conn
	StartSessions(Accounts, SettingsManager)
	StartSession(Account, Settings) error
	StopSession(Account)
}

func NewChatManager() ChatManager {
	return chatManager{}
}

//Empty struct - TODO: Make this a non-empty struct, needs clientListenerMap
//and other pertinent data
type chatManager struct{}

func (cm chatManager) SessionNotifier(acct Account) chan<- irc.Conn {
	return getSessionNotifier(acct.Username())
}

func (cm chatManager) StartSessions(accts Accounts, settings SettingsManager) {
	log.Printf("Starting sessions...")
	startUserSessions(accts, settings)
}

func (cm chatManager) StartSession(acct Account, settings Settings) error {
	return startSession(acct, settings)
}

func (cm chatManager) StopSession(acct Account) {
	log.Printf("Called stop session. This still needs to be implemented.")
}

//ircManager takes the connection to the IRC server and then coordinates the
//communication between the irc server, and the active IRCClients
func ircManager(ircConn irc.Conn, newClients chan irc.Conn) {
	fmt.Println("*** Entering ircManager ***")
	defer fmt.Println("*** Leaving ircManager ***")

	var clients []irc.Conn

	fromServer := make(chan irc.Message)
	fromClient := make(chan irc.Message)
	errChan := make(chan error)
	quitChan := make(chan bool)

	//Listen for incoming messages form server
	go ircServerListener(ircConn, fromServer, errChan, quitChan)

	for {
		select {
		case pmsg := <-fromServer:
			msg := pmsg.Message
			//Log it,
			log.Println(msg)

			//Repsond if it's a ping
			if pmsg.Command == "PING" {
				ircConn.Write(irc.NewMessage("PONG " + pmsg.Params[0]))
			}

			//...and send it to all clients
			for k := 0; k < len(clients); k++ {
				client := clients[k]
				err := client.Write(pmsg)

				if err != nil {
					stopClientListener(client)
					client.Close()
					clients = deleteNthItem(clients, k)
					k-- //Account for socket deletion in slice
					fmt.Println("*** Disconnected irc Client. ", len(clients), "remaining.")
				}
			}
		//Received a message from the server
		case msg := <-fromClient:
			err := ircConn.Write(msg)
			if err != nil {
				fmt.Println("Error writing to server: ", err)
			}

		//A new client has connected
		case client := <-newClients:
			clients = append(clients, client)
			startClientListener(client, fromClient)
			fmt.Println("*** Accepted the ", len(clients), " client connection ***")
		}
	}

	//quitChan <- true
}

func deleteNthItem(a []irc.Conn, n int) []irc.Conn {
	a, a[len(a)-1] = append(a[:n], a[n+1:]...), nil
	return a
}

//ircServerListener continuallyu listens for messages from the IRC server.
//When one is receives, it sends the message into the msg channel.
func ircServerListener(ircConn irc.Conn, msgChan chan<- irc.Message, errChan chan<- error, quitChan <-chan bool) {
	fmt.Println("*** Entering ircListenerClient ***")
	defer fmt.Println("*** Leaving ircListenerClient ***")
	for {

		select {
		case <-quitChan:
			return
		default:
			msg, err := ircConn.Read()
			if err != nil {
				errChan <- err
				return
			}

			msgChan <- msg
		}
	}
}

//ircCLientListener will indefinitely listen for input from an ircClient, putting
//it into the supplied channel, where it will be sent on to the server
func ircClientListener(client irc.Conn, toServer chan<- irc.Message, quit <-chan bool) {
	for {
		select {
		case <-quit:
			fmt.Println("Exiting ircClientListener")
			return
		default:
			msg, err := client.Read()
			if err != nil {
				//fmt.Println(fmt.Sprintf("ircClientListener %v, error: %v", client, err))
				//time.Sleep(1000 * time.Millisecond)
				return
			}

			toServer <- msg
		}
	}
}

//startClientListener and stopClientListener start and stop the ircClientListener
//for the particular ircClient
//Keep track of the quit bool channel for each client listener
//TODO Find a better way to implement this
var clientListenerMap = make(map[irc.Conn]chan bool)

func startClientListener(client irc.Conn, toServer chan<- irc.Message) {
	quitCh := make(chan bool, 1)
	clientListenerMap[client] = quitCh
	go ircClientListener(client, toServer, quitCh)
}

func stopClientListener(client irc.Conn) {
	//fmt.Println(fmt.Sprintf("Telling client %+v listener  to quit...", client))
	ch, ok := clientListenerMap[client]
	if ok {
		ch <- true
		//fmt.Println("... Successfully sent quit notice")
	}
}

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
func startUserSessions(accts Accounts, settings SettingsManager) {
	for _, acct := range accts.accountMap() {
		settings, err := settings.Settings(acct)
		log.Printf("Starting session for %s. Settings: %+v", acct.Username(), settings)
		if err == nil && settings.Enabled() {
			err := startSession(acct, settings)
			if err != nil {
				log.Printf("Trouble starting session for %s: %s", acct.Username(), err.Error())
			} else {
				log.Println("Session started successfully.")
			}
		}
	}
}

func startSession(user Account, settings Settings) error {
	log.Printf("Starting session for user %s", user.Username())
	//Start the IRC connection... //TODO: Move this elsewhere.

	addr := fmt.Sprintf("%s:%d", settings.Address(), settings.Port())
	conn, err := irc.NewConnection(addr, settings.SSL())
	if err != nil {
		return err
	}

	//Channel to accept new http clients logging in
	newClients := make(chan irc.Conn)

	conn.Write(irc.UserMessage(user.Username(), "servername", "hostname", "irctestuser"))
	conn.Write(irc.NickMessage(settings.Login().Nick))
	conn.Write(irc.NewMessage("join #gotest"))
	go ircManager(conn, newClients)
	sessionNewClientsMap[user.Username()] = newClients

	log.Printf("Session for user %s started with no errors.", user.Username())
	return nil
}
