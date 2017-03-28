package main

import (
	"log"
	"net/http"
	"time"
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
	server := request.PostFormValue("server")
	if server == "" { // key not present
		return
	}
	
	Servers[server].Mux.Lock()
	Servers[server].ActiveThread++
	ThisThreadNummer := Servers[server].ActiveThread
	Servers[server].Mux.Unlock()
	
	switch cmdtype := request.PostFormValue("type"); cmdtype {
		case "init" : // nothing to do, just keep the connection
		case "chat" :
			player := request.PostFormValue("player")
			message := request.PostFormValue("message")
			forwardMessageToDiscord(server, player, message)
		default: return
	}
	
	for {
		select {
		case cmd := <-Servers[server].Outbound :
			// send response with q to game server
			log.Println("Found",cmd)
			js, err := json.Marshal(cmd)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(js)
			log.Println("Sending", js)
			return
		default :
			time.Sleep(time.Duration(100) * time.Millisecond)
			if ThisThreadNummer != Servers[server].ActiveThread {
				return
			}
		}
	}
}
