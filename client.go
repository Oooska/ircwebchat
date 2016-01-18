package ircwebchat

import (
	"fmt"
	"log"

	"github.com/oooska/irc"
)

var clientListenerMap map[*IRCClient]chan bool = make(map[*IRCClient]chan bool)

/*IRCClient defines the behavior of a browser-based client.
DONE: Websocket client  (websocket.go)
TODO: Longpoll client
*/
type IRCClient interface {
	SendMessage(irc.Message) error
	ReceiveMessage() (irc.Message, error)
	Close()
}

//ircManager takes the connection to the IRC server and then coordinates the
//communication between the irc server, and the active IRCClients
func ircManager(ircConn irc.IRCConn) {
	fmt.Println("*** Entering ircManager ***")
	defer fmt.Println("*** Leaving ircManager ***")

	var clients []*IRCClient

	fromServer := make(chan irc.Message)
	fromClient := make(chan irc.Message)
	errChan := make(chan error)
	quitChan := make(chan bool)

	//Listen for incoming messages form server
	go ircServerListener(ircConn, fromServer, errChan, quitChan)

	for {
		select {
		case msg := <-fromServer:
			//Log it,
			log.Println(msg)

			//Repsond if it's a ping
			if len(msg) >= 4 && msg[0:4] == "PING" {
				var end string = msg[4:].String()
				ircConn.Write(irc.NewMessage("PONG" + end))
				//break
			}

			//...and send it to all clients
			for k := 0; k < len(clients); k++ {
				client := *clients[k]
				err := client.SendMessage(msg)

				if err != nil {
					stopClientListener(client)
					client.Close()
					clients = deleteNthItem(clients, k)
					k-- //Account for socket deletion in slice
				}
			}
		//Received a message from the server
		case msg := <-fromClient:
			fmt.Println("Receiving message from ircclient: ", msg)
			err := ircConn.Write(msg)
			if err != nil {
				fmt.Println("Error writing to server: ", err)
			}

		//A new client has connected
		case client := <-newClients:
			clients = append(clients, client)
			startClientListener(*client, fromClient)
			fmt.Println("*** Accepted the ", len(clients), " client connection ***")
		}
	}

	quitChan <- true
}

func deleteNthItem(a []*IRCClient, n int) []*IRCClient {
	a, a[len(a)-1] = append(a[:n], a[n+1:]...), nil
	return a
}

//ircServerListener continuallyu listens for messages from the IRC server.
//When one is receives, it sends the message into the msg channel.
func ircServerListener(ircConn irc.IRCConn, msgChan chan<- irc.Message, errChan chan<- error, quitChan <-chan bool) {
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

func startClientListener(client IRCClient, to_server chan<- irc.Message) {
	quitCh := make(chan bool, 1)
	clientListenerMap[&client] = quitCh
	go ircClientListener(client, to_server, quitCh)
}

func stopClientListener(client IRCClient) {
	//fmt.Println(fmt.Sprintf("Telling client %+v listener  to quit...", client))
	ch, ok := clientListenerMap[&client]
	if ok {
		ch <- true
		//fmt.Println("... Successfully sent quit notice")
	}
}
func ircClientListener(client IRCClient, to_server chan<- irc.Message, quit <-chan bool) {
	for {
		select {
		case <-quit:
			fmt.Println("Exiting ircClientListener")
			return
		default:
			msg, err := client.ReceiveMessage()
			if err != nil {
				//fmt.Println(fmt.Sprintf("ircClientListener %v, error: %v", client, err))
				//time.Sleep(1000 * time.Millisecond)
				return
			} else {
				fmt.Println("ircClientListener received message: ", msg)
				to_server <- msg
			}
		}
	}
}
