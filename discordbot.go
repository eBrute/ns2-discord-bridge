package main

import (
	"flag"
	// "fmt"
	// "log"
)


// Parse command line arguments
func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	address = ":8080"
	flag.Parse()
}


func main() {

	startDiscordBot()
	
	startHTTPServer()

	// keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}
