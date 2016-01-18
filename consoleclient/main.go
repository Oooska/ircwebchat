package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/oooska/irc"
	"golang.org/x/net/websocket"
)

//Code adapted from https://github.com/husio/irc/blob/master/examples/echobot.go

//A basic irc client that communicates over a websocket.
func main() {
	origin := "http://127.0.0.1/"
	url := "ws://127.0.0.1:8080/chat/socket"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		panic(err)
	}
	conn := irc.IRCConnectionWrapper(ws)

	go func() { //Read from stdin
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Cannot read from stdin: %s", err)
			}

			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			msg, err := parseLine(line)
			if err != nil {
				log.Println("Err: ", err)
			} else {
				log.Println("YOU: ", msg)
				conn.Write(msg)
			}
		}
	}()

	var msg irc.Message

	// handle incomming messages
	for {

		msg, err = conn.Read()
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return
		}

		if len(msg) >= 4 && msg[0:4] == "PING" {
			var end string = msg[4:].String()
			conn.Write(irc.NewMessage("PONG" + end))
		}

		fmt.Println(msg)
	}
}

//parseLine returns an irc.Message object. If the line starts with a forward
//slash, everything after the '/' is converted directly to a server command
//If there is no slash, the first word is taken to be the channel or user to
//send a PRIVMSG to
func parseLine(line string) (msg irc.Message, err error) {
	if line[0] == '/' {
		msg = irc.NewMessage(line[1:]) //TODO Parse actual command
	} else {
		splitlines := strings.SplitN(line, " ", 2)
		if len(splitlines) > 1 {
			msg = irc.PrivMessage(splitlines[0], splitlines[1])
		} else {
			err = errors.New("Unable to parse input")
		}
	}
	return
}
