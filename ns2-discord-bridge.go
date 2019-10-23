package main

import (
	"flag"
	"log"
)

const version = "v5.5.0"
var configFile string

// parse command line arguments
func init() {
	flag.StringVar(&configFile, "c", "config.toml", "Specify Configuration File")
	flag.Parse()
}


func main() {
	log.Println("Version", version)
	Config.loadConfig(configFile)

	for serverName, v := range Config.Servers {
		serverList[serverName] = &Server{
			Name : serverName,
			Config : v,
			Outbound : make(chan *Command),
			Muted : v.Muted,
			TimeoutSet : make(chan int),
			TimeoutReset : make(chan int),
		}
		log.Println("Linked server '"+ serverName +"' to channel", v.ChannelID)
	}

	for _, server := range serverList {
		go server.clearOutboundChannelOnInactivity()
	}

	startDiscordBot() // non-blocking
	startHTTPServer() // blocking
}
