package main

import (
	"log"
	"net/http" 
)

var (
	address string
)

func startHTTPServer() {
	http.HandleFunc("/hello", httpHandler)

	log.Fatal(http.ListenAndServe(address, nil))

}


func httpHandler(w http.ResponseWriter, request *http.Request) {
	err := request.ParseForm() 
	if err != nil{
	       log.Print(err)
	}
	
	server := request.PostFormValue("server")
	username := request.PostFormValue("username")
	message := request.PostFormValue("message")
	
	forwardMessage(server, username, message)
}