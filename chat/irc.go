package chat

import (
	"fmt"
	"log"

	"github.com/oooska/irc"
)

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
