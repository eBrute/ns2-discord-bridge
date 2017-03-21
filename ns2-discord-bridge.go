package main

import (
    "log"
    "os"
	"io/ioutil"
    "flag"
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
var configFile string


// Parse command line arguments
func init() {
	flag.StringVar(&configFile, "c", "config.toml", "Configuration File")
	flag.Parse()
}


func main() {
    if _, err := os.Stat(configFile); os.IsNotExist(err) {
        log.Println("No configuration file found in", configFile)
        return
    }
    
	f, err := os.Open(configFile)
    if err != nil {
        panic(err)
    }
    defer f.Close()
    log.Println("Reading config file", configFile)
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
