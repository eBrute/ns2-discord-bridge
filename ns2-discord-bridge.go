package main

import (
	"flag"
)

var configFile string

// parse command line arguments
func init() {
	flag.StringVar(&configFile, "c", "config.toml", "Configuration File")
	flag.Parse()
}


func main() {
	Config.loadConfig(configFile)

	for _, server := range serverList {
		go server.clearOutboundChannelOnInactivity()
	}
	
	startDiscordBot() // non-blocking
	startHTTPServer() // blocking
}
