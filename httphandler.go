// This file contains functions related to handling requests coming in through http

package main

import (
	"log"
	"net/http"
	"time"
	"strconv"
	"encoding/json"
)

const recordSep = ""


func startHTTPServer() {
	http.HandleFunc("/discordbridge", httpHandler)

	log.Println("Listening for messages on", Config.Httpserver.Address)
	log.Println("Press CTRL-C to exit.")
	log.Fatal(http.ListenAndServe(Config.Httpserver.Address, nil))
}


func httpHandler(w http.ResponseWriter, request *http.Request) {
	err := request.ParseForm() 
	if err != nil {
		log.Print(err)
	}

	serverName := request.PostFormValue("id")
	if serverName == "" { // key not present
		return
	}

	server, ok := serverList[serverName]
	if !ok {
		log.Println("Recieved message but could not get a channel for '" + serverName + "'. Link a channel first with '!link <servername>'")
		return
	}
	
	// announce that we are now responsible for the response
	// all other threads will stop themselves
	server.Mux.Lock()
	server.ActiveThread++
	thisThreadNummer := server.ActiveThread
	server.Mux.Unlock()
	
	// handle the incoming request
	groupType := request.PostFormValue("type")
	subType := request.PostFormValue("sub")
	messageType := MessageType{groupType, subType}
	switch messageType.GroupType {
		case "init": // nothing to do, just keep the connection
		
		case "chat":
			player := request.PostFormValue("plyr")
			steamid, _ := strconv.ParseInt(request.PostFormValue("sid"), 10, 32)
			teamNumber, _ := strconv.Atoi(request.PostFormValue("team"))
			message := request.PostFormValue("msg")
			forwardChatMessageToDiscord(server, player, SteamID3(steamid), TeamNumber(teamNumber), message)
			
		case "player": 
			player := request.PostFormValue("plyr")
			steamid, _ := strconv.ParseInt(request.PostFormValue("sid"), 10, 32)
			playerCount := request.PostFormValue("pc")
			forwardPlayerEventToDiscord(server, messageType, player, SteamID3(steamid), playerCount)
			
		case "status": fallthrough
		case "adminprint":
			playerCount := request.PostFormValue("pc")
			message := request.PostFormValue("msg")
			forwardStatusMessageToDiscord(server, messageType, message, playerCount)
			
		case "info":
			serverInfo := ServerInfo{}
			message := request.PostFormValue("msg")
			err := json.Unmarshal([]byte(message), &serverInfo)
			if err != nil {
				log.Println(err.Error())
			}
			forwardServerStatusToDiscord(server, messageType, serverInfo)
			
		case "test":
			forwardStatusMessageToDiscord(server, messageType, "Test successful", "")
			return
			
		default: return
	}
	
	// build a response
	for {
		select {
		case cmd := <-server.Outbound:
			server.TimeoutReset <- 0
			response := cmd.Type + recordSep + cmd.User + recordSep + cmd.Message
			w.Write([]byte(response))
			return
			
		default:
			time.Sleep(time.Duration(100) * time.Millisecond)
			if thisThreadNummer != server.ActiveThread {
				return
			}
		}
	}
}
