package main

import (
	"log"
	"net/http"

	"github.com/oooska/ircwebchat"
)

//Starts a basic http server with the ircwebchat Handler registered
func main() {
	mux := http.NewServeMux()
	ircwebchat.Register(*mux)
	go log.Fatal(http.ListenAndServe(":8080", mux))
}
