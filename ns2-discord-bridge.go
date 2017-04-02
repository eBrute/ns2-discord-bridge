package main

import (
    "log"
    "os"
	"io/ioutil"
    "flag"
    "sync"
    "errors"
    "time"
	"github.com/naoina/toml"
    "github.com/bwmarrin/discordgo"
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
    Name string
    ChannelID string
    Admins []string
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


func GetServerByName(serverName string) (server *Server, ok bool) {
    server, ok = Servers[serverName]
    return
}


func LinkChannelIDToServer(channelID string, server *Server) error {
    if linkedServer, ok := GetServerLinkedToChannel(channelID); ok {
        if linkedServer == server {
            return errors.New("This channel was already linked to '" + linkedServer.Name + "'")
        } else {
            return errors.New("This channel is already linked to '" + linkedServer.Name +"'. Use !unlink first.")
        }
    }
    server.ChannelID = channelID
    log.Println("Linked channelID " + channelID + " to server " + server.Name)
    return nil
}


func GetServerLinkedToChannel(channelID string) (server *Server, success bool) {
    for _, v := range Servers {
        if v.ChannelID == channelID {
            return v, true
        }
    }
    return
}

func GetIsAdminForServer(user *discordgo.User, server *Server) bool {
    userName := user.Username + "#" + user.Discriminator
    userID := user.ID
    for _, admin := range server.Admins {
        if admin == userID || admin == userName {
            return true
        }
    }
    return false
}


func UnlinkChannelFromServer(server *Server) (success bool) {
	if server != nil {
		log.Println("Uninked channelID " + server.ChannelID + " from server " + server.Name)
        server.ChannelID = ""
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
        server := &Server{
            Name : serverName,
            Admins : make([]string, 0),
            Outbound : make(chan Command),
        }
        server.ChannelID = v.ChannelID
        for _, admin := range v.Admins {
            server.Admins = append(server.Admins, admin)
        }
        Servers[serverName] = server
        log.Println(Servers[serverName])
        log.Println("Linked server", serverName, "to channel", v.ChannelID)
    }
    
	startDiscordBot()
	startHTTPServer()
}
