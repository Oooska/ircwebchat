package ircwebchat

import (
	"bufio"
	"fmt"
	"time"

	"github.com/oooska/irc"
	"golang.org/x/net/websocket"
)

/*The WSClient struct is an imeplementation of IRCClient using websockets.
 */
type WSClient struct {
	conn   *websocket.Conn
	reader bufio.Reader
}

//Sends a message to the client/user.
func (wsclient WSClient) SendMessage(msg irc.Message) error {
	_, err := wsclient.conn.Write([]byte(msg.String()))
	return err
}

//Receives a message from the client/user
func (wsclient WSClient) ReceiveMessage() (irc.Message, error) {
	msg, err := wsclient.reader.ReadString('\n')
	return irc.NewMessage(msg), err
}

func (wsclient WSClient) Close() {
	wsclient.conn.Close()
}

//Returns a new WebSocket connection
func NewWSClient(conn *websocket.Conn) WSClient {
	return WSClient{conn: conn, reader: *bufio.NewReader(conn)}
}

/*
The socketHandler for the websocket connection.
Accepts the websocket, hands it off through the socketChan, and
waits until the socket is closed before exiting the function.
//TODO: Actually manage disconnections properly.
*/
func webSocketHandler(ws *websocket.Conn) {
	//Notify the irc manager of a new websocket
	fmt.Println("!!!!socketHandler starting!!!!")
	var client IRCClient = NewWSClient(ws)
	newClients <- &client
	for {
		if ws.IsServerConn() {
			time.Sleep(10 * time.Millisecond)
		} else {
			fmt.Println("!!!!socketHandler returning after IsServerConn returned false")
			return
		}
	}
}
