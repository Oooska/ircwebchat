package main

import (
	"log"
	"net/http"
    _ "net/http/pprof"
	"github.com/oooska/ircwebchat"
)

//Starts a basic http server with the ircwebchat Handler registered
func main() {
	//mux := http.NewServeMux()
	ircwebchat.Register(nil)//mux)
	go log.Fatal(http.ListenAndServe(":8080", nil))
}
