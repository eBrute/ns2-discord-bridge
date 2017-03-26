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
    Servers map[string]ChannelConfig
}

type ChannelConfig struct {
    ChannelID string
    Admins []string
}

var Config Configuration
var configFile string

var Servers map[string]*Channel

type Channel struct {
    ChannelID string
    Admins []string
}


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
    
    Servers = make(map[string]*Channel)
    for serverName, v := range Config.Servers {
        Servers[serverName] = &Channel{
            ChannelID : v.ChannelID,
            Admins:make([]string, 0),
        }
        for _, admin := range v.Admins {
            Servers[serverName].Admins = append(Servers[serverName].Admins, admin)
        }
        log.Println(Servers[serverName])
        log.Println("Linked server", serverName, "to channel", v.ChannelID)
    }
    
	startDiscordBot()
	startHTTPServer()
}
