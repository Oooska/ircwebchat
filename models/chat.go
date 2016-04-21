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

func NewChatManager() ChatManager {
	return chatManager{}
}

//Empty struct - TODO: Make this a non-empty struct, needs clientListenerMap
//and other pertinent data
type chatManager struct {
	chatmap map[string]Chat
}

func (cm chatManager) StartChats(accts Accounts, settings SettingsManager) {
	log.Printf("Starting chats...")
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

func (cm chatManager) StartChat(acct Account, settings Settings) error {
	chat, ok := cm.chatmap[acct.Username()]
	if ok && chat.Active() {
		return errors.New("Chat already started for " + acct.Username())
	}
	chat = newChat(acct, settings)
	return chat.Start()
}

func (cm chatManager) StopChat(acct Account) {
	chat, ok := cm.chatmap[acct.Username()]
	if ok && chat.Active() {
		chat.Stop()
	}
}

//joinchat
func (cm chatManager) JoinChat(acct Account, sessionID string, webclient net.Conn) error {
	c, ok := cm.chatmap[acct.Username()]
	if !ok {
		return errors.New("Chat is not enabled or active.")
	}
	err := c.Join(sessionID, irc.NewConnectionWrapper(webclient))
	return err
}

func newChat(acct Account, settings Settings) Chat {
	chat := &chat{
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

type Chat interface {
	Start() error
	Stop()
	Join(sessionID string, webclient irc.Conn) error
	Active() bool
}

type chat struct {
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

type clientMessage struct {
	SessionID string
	Message   irc.Message
}

func (c chat) Start() error {
	if !c.running {
		log.Printf("Starting chat...")
		client, err := irc.NewClient(fmt.Sprintf("%s:%d", c.settings.Address(), c.settings.Port()), c.settings.SSL())
		if err != nil {
			log.Printf("Error starting chat: %s", err.Error())
			return err
		}
		c.client = client
		log.Printf("Calling ircManager in its own goroutine...")
		go ircManager(c)
		c.running = true
	}
	return nil
}

func (c chat) Stop() {
	if c.running {
		close(c.quit)
		c.running = false
	}
}

//Coordinates webclient
//Blocks until webclient closes or chat ends
func (c chat) Join(sessionID string, webclient irc.Conn) error {
	if c.Active() {
		webclient.Write(irc.NickMessage(c.settings.Login().Nick))
		//TODO: Send rooms, users, and logs

		//Listens for incoming messages and b
		//c.newClients <- webclient
		c.webClientsLock.Lock()
		c.webclients[sessionID] = webclient
		c.webClientsLock.Unlock()

		var err error
		for err == nil {
			select {
			case <-c.quit:
				fmt.Println("Exiting ircClientListener")
				err = errors.New("IRC Session has ended")
			default:
				msg, err := webclient.Read()
				if err != nil {
					log.Printf("Error in webclient JoiN(): %s", err.Error())
					break
				}
				c.toServer <- clientMessage{SessionID: sessionID, Message: msg}
			}
		}
		//Error has occured, removed from list of connected clients
		c.webClientsLock.Lock()
		delete(c.webclients, sessionID)
		c.webClientsLock.Unlock()
		return err
	}
	return errors.New("The chat session is not active or enabled. Check settings")
}

func (c chat) Active() bool {
	return c.running
}

//ircManager takes the connection to the IRC server and then coordinates the
//communication between the irc server, and the active IRCClients
func ircManager(c chat) { //ircConn irc.Conn, newClients chan irc.Conn
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

		//Received a message from the server
		case msg := <-c.toServer:
			err := c.client.Write(msg.Message)
			if err != nil {
				fmt.Println("Error writing to server: ", err)

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
		}
	}
	//quitChan <- true
}

//ircServerListener indefinitely listens for messages from the IRC server.
//When one is receives, it sends the message into the msg channel.
func serverListener(ircConn irc.Conn, msgChan chan<- irc.Message, errChan chan<- error, quitChan <-chan bool) {
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
