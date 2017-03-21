package main

import (
	"log"
	"net/http" 
)

var (
	address string
)

func startHTTPServer() {
	http.HandleFunc("/discordbridge", httpHandler)

	log.Println("Listening for chat messages on", address)
	log.Println("Press CTRL-C to exit.")
	log.Fatal(http.ListenAndServe(address, nil))
}


func httpHandler(w http.ResponseWriter, request *http.Request) {
	err := request.ParseForm() 
	if err != nil{
	       log.Print(err)
	}
	
	server := request.PostFormValue("server")
	player := request.PostFormValue("player")
	message := request.PostFormValue("message")
	
	forwardMessageToDiscord(server, player, message)
}