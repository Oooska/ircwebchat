package models

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/oooska/irc"
)

//ChatManager keeps track of active connections to IRC servers
//This interface will be cleaned up shortly
type ChatManager interface {
	StartChats(Accounts, SettingsManager)
	StartChat(Account, Settings) error
	StopChat(Account)
	JoinChat(acct Account, sessionID string, webclient net.Conn) error
}

//NewChatManager returns a new chat manager to manage communication
//between IRC servers and webclients
func NewChatManager() ChatManager {
	return chatManager{chatmap: make(map[string]chat)}
}

type chatManager struct {
	chatmap map[string]chat
}

//StarChats checks all accounts in the system and connects to the IRC server
//if the account is enabled.
func (cm chatManager) StartChats(accts Accounts, settings SettingsManager) {
	for _, acct := range accts.accountMap() {
		settings, err := settings.Settings(acct)
		if err == nil && settings.Enabled() {
			err := cm.StartChat(acct, settings)
			if err != nil {
				log.Printf("Trouble starting chat for %s: %s", acct.Username(), err.Error())
			} else {
				log.Printf("Chat started successfully for %s", acct.Username())
			}
		}
	}
}

//StartChat creates a connection to the IRC server for the specified user
func (cm chatManager) StartChat(acct Account, settings Settings) error {
	chat, ok := cm.chatmap[acct.Username()]
	if ok && chat.Active() {
		return errors.New("Chat already started for " + acct.Username())
	}
	chat = newChat(acct, settings)
	cm.chatmap[acct.Username()] = chat
	return chat.Start()
}

//StopChat ends the connection to the IRC server for the specified user
func (cm chatManager) StopChat(acct Account) {
	chat, ok := cm.chatmap[acct.Username()]
	if ok && chat.Active() {
		chat.Stop()
	}
}

//JoinChat connects a webclient to the Chat if it is active
func (cm chatManager) JoinChat(acct Account, sessionID string, webclient net.Conn) error {
	c, ok := cm.chatmap[acct.Username()]
	if !ok {
		return errors.New("Unable to find chat to join")
	}
	log.Printf("Chat found for %s", acct.Username())
	err := c.Join(sessionID, irc.NewConnectionWrapper(webclient))
	return err
}

func newChat(acct Account, settings Settings) chat {
	chat := &ircchat{
		account:        acct,
		settings:       settings,
		quit:           make(chan bool),
		running:        false,
		toClients:      make(chan irc.Message),
		toServer:       make(chan clientMessage),
		webclients:     make(map[string]irc.Conn),
		webClientsLock: new(sync.RWMutex),
	}

	return chat
}

//chat manages the connection between the IRC server and the web clients
type chat interface {
	Start() error
	Stop()
	Join(sessionID string, webclient irc.Conn) error
	Active() bool
}

type ircchat struct {
	account Account
	client  irc.Client //Gets set by Start()

	settings  Settings
	quit      chan bool
	running   bool
	toClients chan irc.Message
	toServer  chan clientMessage

	webclients     map[string]irc.Conn //key:sessionID
	webClientsLock *sync.RWMutex
}

//clientMessage is a Message that was sent from a specific webclient
type clientMessage struct {
	SessionID string
	Message   irc.Message
}

//Start connects to the IRC server, authenticates, and then starts a goroutine to manage the chat
func (c *ircchat) Start() error {
	if !c.running {
		log.Printf("Starting chat for %s...", c.account.Username())
		client, err := irc.NewClient(fmt.Sprintf("%s:%d", c.settings.Address(), c.settings.Port()), c.settings.SSL())
		if err != nil {
			log.Printf("Error starting chat for %s: %s", c.account.Username(), err.Error())
			return err
		}
		c.client = client
		client.Write(irc.UserMessage(c.account.Username(), "ircwebchathost", "somewhere", "quack"))
		client.Write(irc.NickMessage(c.settings.Login().Nick))

		c.running = true
		go ircManager(*c)

	}
	return nil
}

//Stop causes the IRC server to disconnect, dropping any clients
func (c *ircchat) Stop() {
	if c.running {
		close(c.quit)
		c.running = false
	}
}

//Join adds a webclient to the list of active clients. It blocks until webclient socket closes or chat ends
func (c ircchat) Join(sessionID string, webclient irc.Conn) error {
	log.Printf("User %s /w session ID %s is joining chat.", c.account.Username(), sessionID)

	if c.Active() {
		webclient.Write(irc.NickMessage(c.settings.Login().Nick))
		//TODO: Send rooms, users, and logs to webclient

		//Register as a listener
		c.registerClient(sessionID, webclient)
		for {
			select {
			case <-c.quit:
				fmt.Println("Exiting ircClientListener")
				return errors.New("IRC Session has ended")
			default:
				msg, err := webclient.Read()
				if err != nil {
					log.Printf("Error reading from webclient %s (%s): %s", c.account.Username(), sessionID, err.Error())
					c.unregisterClient(sessionID)
					return err
				}
				c.toServer <- clientMessage{SessionID: sessionID, Message: msg}
			}
		}
	}
	return errors.New("The chat session is not active or enabled. Check settings")
}

//Active returns true if the chat is connected and running
func (c ircchat) Active() bool {
	return c.running
}

//registerClient registers the webclient to recieve messages from the irc server
func (c ircchat) registerClient(sessionID string, webclient irc.Conn) {
	c.webClientsLock.Lock()
	c.webclients[sessionID] = webclient
	c.webClientsLock.Unlock()
}

//unregisterClient removes the webclient from recieving messages from the irc server
func (c ircchat) unregisterClient(sessionID string) {
	c.webClientsLock.Lock()
	delete(c.webclients, sessionID)
	c.webClientsLock.Unlock()
}

//ircManager takes the connection to the IRC server and then coordinates the
//communication between the irc server, and the connected webclients
//It it will block until the connection to the irc server is closed, or Stop() is called
func ircManager(c ircchat) { //ircConn irc.Conn, newClients chan irc.Conn
	fmt.Println("*** Entering ircManager ***")
	defer fmt.Println("*** Leaving ircManager ***")

	fromServer := make(chan irc.Message)
	errChan := make(chan error) //errors from irc server

	//Listen for incoming messages form server
	go serverListener(c.client, fromServer, errChan, c.quit)

	for {
		select {
		case msg := <-fromServer: //Recieved message from irc server
			//Send it to all clients
			//log.Printf("%s: %s", c.account.Username(), msg.Message)
			c.webClientsLock.RLock()
			for _, client := range c.webclients {
				err := client.Write(msg)
				if err != nil {
					fmt.Printf("***Error sending to websocket for %s.", c.account.Username())
				}
			}
			c.webClientsLock.RUnlock()

		//Received a message from a webclient
		case msg := <-c.toServer:
			err := c.client.Write(msg.Message)
			if err != nil {
				fmt.Printf("Error writing to server: %s", err.Error())
			}

			c.webClientsLock.RLock()
			for k, client := range c.webclients {
				if k != msg.SessionID { //Notify webclients other than the one that sent it
					cerr := client.Write(msg.Message)
					if cerr != nil {
						fmt.Printf("***Error sending to websocket for %s.", c.account.Username())
					}
					if err != nil && cerr == nil { //If no error from client, but an error from server, forward server error
						client.Write(irc.NewMessage("Error sending message: " + err.Error()))
					}
				}
			}
			c.webClientsLock.RUnlock()
		case err := <-errChan:
			log.Printf("Recieved error from serverListerner: %s", err.Error())
			close(c.quit)
		case <-c.quit:
			log.Printf("Stopping the chat. Disconnecting client.")
			c.Stop()
			return
		}
	}
}

//serverListener listens for messages from the IRC server.
//When one is receives, it sends the message into the msg channel.
func serverListener(ircConn irc.Conn, msgChan chan<- irc.Message, errChan chan<- error, quitChan <-chan bool) {
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
