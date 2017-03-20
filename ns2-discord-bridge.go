package main

import (
	"flag"
)


// Parse command line arguments
func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&address, "a", ":8080", "HTTP Server address")
	flag.Parse()
}


func main() {
	startDiscordBot()
	startHTTPServer()
}
