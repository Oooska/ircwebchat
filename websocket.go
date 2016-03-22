package ircwebchat

import (
	"bufio"

	"log"
	"strings"
	"time"

	"github.com/oooska/irc"
	"golang.org/x/net/websocket"
)

/*The WSClient struct is an imeplementation of IRCClient using websockets.
 */
type wsClient struct {
	conn   *websocket.Conn
	reader bufio.Reader
}

//Sends a message to the client/user.
func (wsclient wsClient) SendMessage(msg irc.Message) error {
	_, err := wsclient.conn.Write([]byte(msg.String()))
	return err
}

//Receives a message from the client/user
//TODO: Remove terrible parsing rules in parseLine()
func (wsclient wsClient) ReceiveMessage() (irc.Message, error) {
	msg, err := wsclient.reader.ReadString('\n')
	return irc.NewMessage(msg), err
}

func (wsclient wsClient) Close() {
	wsclient.conn.Close()
}

//Returns a new WebSocket connection
func newWSClient(conn *websocket.Conn) wsClient {
	return wsClient{conn: conn, reader: *bufio.NewReader(conn)}
}

/*
The socketHandler for the websocket connection.
Accepts the websocket, hands it off through the socketChan, and
waits until the socket is closed before exiting the function.
//TODO: Actually manage disconnections properly.
*/
func webSocketHandler(ws *websocket.Conn) {
	//Notify the irc manager of a new websocket
	log.Println("socketHandler starting")
	defer log.Println("socketHandler exiting")
	var client ircClient = newWSClient(ws)

	//Authenticate websocket:
	var user *iwcUser

	for user == nil {
		client.SendMessage(irc.NewMessage("CLIENT-MESSAGE :Enter a username."))
		msg, err := client.ReceiveMessage()
		if err != nil {
			return
		}
		username := strings.TrimSpace(msg.String())

		client.SendMessage(irc.NewMessage("CLIENT-MESSAGE :Enter a password."))
		msg, err = client.ReceiveMessage()
		if err != nil {
			return
		}
		password := strings.TrimSpace(msg.String())
		user = authenticate(username, password)

		if user != nil {
			client.SendMessage(irc.NewMessage("CLIENT-MESSAGE :Successfully logged in."))
		} else {
			client.SendMessage(irc.NewMessage("CLIENT-MESSAGE :Invalid username/password."))
		}
	}

	newclients := getSessionNotifier(user.username)
	if newclients == nil {
		client.SendMessage(irc.NewMessage("Unable to find session. Closing..."))
		log.Printf("Unable to find session for %s", user.username)
		return
	}

	//Notify the client what the user's current nick is
	client.SendMessage(irc.NewMessage("NICK " + user.profile.nick.name))
	newclients <- &client

	for {
		if ws.IsServerConn() {
			time.Sleep(100 * time.Millisecond)
		} else {
			log.Println("socketHandler returning after IsServerConn returned false")
			return
		}
	}
}
