package ircwebchat

import (
	"net/http"

	"golang.org/x/net/websocket"
)

/* ircwebchat provides a web-basede IRC client. A user can share the same IRC
session across multiple browsers.

Still in early development stages.

TODO: Currently only sends data to clients. Need to listen to IRCCLients and pass info on to other clients and server
*/

//Register mounts an entry point at /chat/ on the supplied http mux.
//TODO: We currently start the connection to the IRC server here. This
// should be abstracted away.
func Register(mux http.ServeMux) {
	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("/chat/", http.StripPrefix("/chat/", fs))
	mux.Handle("/chat/socket", websocket.Handler(webSocketHandler))

	user := iwcUser{
		username: "goirctest",
		password: "password",
		profile: serverProfile{
			address: "irc.freenode.net:6667",
			nick: login{
				name:     "goirctest",
				password: "",
			},
			realname: "go-get-real",
			altnick: login{
				name:     "goirctest_",
				password: "",
			},
		},
	}
	user2 := iwcUser{
		username: "goirctest2",
		password: "password",
		profile: serverProfile{
			address: "irc.freenode.net:6667",
			nick: login{
				name:     "goirctest2",
				password: "",
			},
			altnick: login{
				name:     "goirctest2_",
				password: "",
			},
			realname: "go-get-real",
		},
	}

	addUser(user)
	addUser(user2)

	startUserSessions()
}
