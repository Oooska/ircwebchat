package models

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
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
	ChatStarted(Account) bool
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
	err := c.Join(sessionID, irc.NewConnectionWrapper(webclient))
	return err
}

//ChatStarted returns true if the chat has started, false if it has not
//started or does not exist.
func (cm chatManager) ChatStarted(acct Account) bool {
	c, ok := cm.chatmap[acct.Username()]
	if !ok {
		return false
	}
	return c.Active()
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
	var err error
	if !c.running {
		var client irc.Client
		client, err = irc.NewClient(fmt.Sprintf("%s:%d", c.settings.Address(), c.settings.Port()), c.settings.SSL())
		if err != nil {
			log.Printf("Error starting chat for %s: %s", c.account.Username(), err.Error())
			return err
		}
		c.client = client
		err = client.Write(irc.UserMessage(c.account.Username(), "ircwebchathost", "somewhere", "quack"))

		login := c.settings.Login()
		if err == nil && login.Nick != "" {
			err = client.Write(irc.NickMessage(c.settings.Login().Nick))
			if err == nil && login.Password != "" {
				err = client.Write(irc.PrivMessage("NickServ", "identify "+login.Password))
			}
		}
		//TODO: Auto join rooms

		if err == nil {
			c.running = true
			go ircManager(*c)
		} else {
			c.Stop()
		}
	}
	return err
}

//Stop causes the IRC server to disconnect, dropping any clients
func (c *ircchat) Stop() {
	if c.running {
		c.running = false
		close(c.quit)
	}
}

//Join adds a webclient to the list of active clients. It blocks until webclient socket closes or chat ends
func (c ircchat) Join(sessionID string, webclient irc.Conn) error {
	if !c.Active() {
		return errors.New("The chat session is not active or enabled. Check settings")
	}

	webclient.Write(irc.NickMessage(c.settings.Login().Nick))
	//Send open channels to client
	for _, ch := range c.client.ChannelNames() {
		webclient.Write(irc.JoinMessage(ch))

		//Send users in the rooms as a names reply commands (353 to indicate start, 366 to indicate end)
		//Webclient doesn't care about length, but a traditional IRC client will
		//TODO: Indicate if channel is public, private or secret ( "=" / "*" / "@" ), current sends as pub
		//:tepper.freenode.net 353 goirctest @ #gotest :goirctest @Oooska
		//:tepper.freenode.net 366 goirctest #gotest :End of /NAMES list.
		users, _ := c.client.Users(ch)
		namesRepl := fmt.Sprintf("353 %s = %s :%s", c.settings.Login().Nick, ch, strings.Join(users, " "))
		namesEndRepl := fmt.Sprintf("366 %s %s", c.settings.Login().Nick, ch)
		webclient.Write(irc.NewMessage(namesRepl))
		webclient.Write(irc.NewMessage(namesEndRepl))

		for _, msg := range c.client.Messages(ch) {
			webclient.Write(irc.NewMessage(msg))
		}
	}

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
				c.unregisterClient(sessionID)
				return err
			}
			c.toServer <- clientMessage{SessionID: sessionID, Message: msg}
		}
	}

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
			log.Printf("Recieved error: %s", err.Error())
			c.Stop()
		case <-c.quit:
			c.client.Write(irc.NewMessage("QUIT"))
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
