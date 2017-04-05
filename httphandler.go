package main

import (
	"log"
	"net/http"
	"time"
	"strconv"
	"encoding/json"
)

func startHTTPServer() {
	http.HandleFunc("/discordbridge", httpHandler)

	log.Println("Listening for chat messages on", Config.Httpserver.Address)
	log.Println("Press CTRL-C to exit.")
	log.Fatal(http.ListenAndServe(Config.Httpserver.Address, nil))
}


func httpHandler(w http.ResponseWriter, request *http.Request) {
	err := request.ParseForm() 
	if err != nil {
	       log.Print(err)
	}

	serverName := request.PostFormValue("server")
	if serverName == "" { // key not present
		return
	}

	server, ok := Servers[serverName]
	if !ok {
		log.Println("Recieved message but could not get a channel for '" + serverName + "'. Link a channel first with '!link <servername>'")
		return
	}
	
	// announce that we are now responsible for the response
	// all other threads will stop themselves
	server.Mux.Lock()
	server.ActiveThread++
	ThisThreadNummer := server.ActiveThread
	server.Mux.Unlock()
	
	// handle the incoming request
	switch cmdtype := request.PostFormValue("type"); cmdtype {
		case "init" : // nothing to do, just keep the connection
		case "chat" :
			player := request.PostFormValue("player")
			steamidString := request.PostFormValue("steamid")
			steamid, _ := strconv.ParseInt(steamidString, 10, 32)
			message := request.PostFormValue("message")
			forwardChatMessageToDiscord(serverName, player, int32(steamid), message)
		case "status" :
			message := request.PostFormValue("message")
			forwardGameStatusToDiscord(serverName, message)
		case "adminprint" :
			message := request.PostFormValue("message")
			forwardAdminPrintToDiscord(serverName, message)
		default: return
	}
	
	// build a response
	for {
		select {
		case cmd := <-server.Outbound :
			server.TimeoutReset <- 0
			js, err := json.Marshal(cmd)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(js)
			return
			
		default :
			time.Sleep(time.Duration(100) * time.Millisecond)
			if ThisThreadNummer != server.ActiveThread {
				return
			}
		}
	}
}
