package main

import (
    "log"
    "os"
	"io/ioutil"
    "flag"
    "sync"
    "time"
	"github.com/naoina/toml"
)


type Configuration struct {
    Discord struct {
        Token string
    }
    Httpserver struct {
        Address string
    }
    Servers map[string]ServerConfig
}

type ServerConfig struct {
    ChannelID string
    Admins []string
    Password string
}

var Config Configuration
var configFile string

var Servers map[string]*Server

type Server struct {
    ChannelID string
    Admins []string
    Password string
    Outbound chan Command
    Mux sync.Mutex
    ActiveThread int
    Timeout time.Duration
}

type Command struct {
    Type    string `json:"type"`
    User    string `json:"user"`
    Content string `json:"content"`
}


func CreateServer(serverName string) *Server {
    server := &Server{
        Admins : make([]string, 0),
        Outbound : make(chan Command),
    }
    Servers[serverName] = server
    return server
}


func LinkChannelIDToServer(channelID string, serverName string) {
	server := CreateServer(serverName)
    server.ChannelID = channelID
	log.Println("Linked channelID " + channelID + " to server " + serverName)
}


func GetServerLinkedToChannel(channelID string) (server string, success bool) {
	for k, v := range Servers {
		if v.ChannelID == channelID {
			return k, true
		}
	}
	return
}


func UnlinkChannelFromServer(server string) (success bool) {
	if _, ok := Servers[server]; ok {
		log.Println("Uninked channelID " + Servers[server].ChannelID + " from server " + server)
		delete(Servers, server)
		success = true
	}
	return
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
    
    Servers = make(map[string]*Server)
    for serverName, v := range Config.Servers {
        server := CreateServer(serverName)
        server.ChannelID = v.ChannelID
        for _, admin := range v.Admins {
            Servers[serverName].Admins = append(Servers[serverName].Admins, admin)
        }
        log.Println(Servers[serverName])
        log.Println("Linked server", serverName, "to channel", v.ChannelID)
    }
    
	startDiscordBot()
	startHTTPServer()
}
