package main

import (
	"log"
	"net/http" 
)

func startHTTPServer() {
	http.HandleFunc("/discordbridge", httpHandler)

	log.Println("Listening for chat messages on", Config.Httpserver.Address)
	log.Println("Press CTRL-C to exit.")
	log.Fatal(http.ListenAndServe(Config.Httpserver.Address, nil))
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