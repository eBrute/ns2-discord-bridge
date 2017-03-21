package main

import (
    "log"
    "os"
	"io/ioutil"
	"github.com/naoina/toml"
)


type Configuration struct {
    Discord struct {
        Token string
    }
    Httpserver struct {
        Address string
    }
    Servers map[string]Channel
}

type Channel struct {
    ChannelID string
}

var Config Configuration

func main() {
	f, err := os.Open("config.toml")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    buf, err := ioutil.ReadAll(f)
    if err != nil {
        panic(err)
    }
    if err := toml.Unmarshal(buf, &Config); err != nil {
        panic(err)
    }
    
    for k, v := range Config.Servers {
        log.Println("Linked server", k, "to channel", v.ChannelID)
    }
        
	startDiscordBot()
	startHTTPServer()
}
