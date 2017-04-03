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
    TimeoutSet chan int
    TimeoutReset chan int
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
    if len(server.Admins) == 0 {
        return true
    }
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


func clearOutboundChannel(outbound chan Command) {
    for {
        select {
            case <- outbound:
            default: return
        }
    }
}


func clearOutboundChannelOnInactivity(server *Server) {
    var timeout int = -1
    for {
        select {
        case <- server.TimeoutReset:
            // log.Println("Timer stopped, server retrieved message")
            timeout = -1
            // NOTE here we assume that the whole channel was cleared, although we only now that one value was read
        case timeout = <- server.TimeoutSet:
            timeout = timeout * 100
            // log.Println("Timer set to", timeout/100)
        default:
        }
        if timeout == 0 {
            // log.Println("Timeout reached")
            clearOutboundChannel(server.Outbound)
            timeout = -1
        }
        if timeout > 0 {
            timeout--
            // if timeout % 100 == 0 {
            //     log.Println("Time left for server to retrieve message: ", timeout/100)
            // }
        }
        time.Sleep(10 * time.Millisecond)
    }
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
            TimeoutSet : make(chan int),
            TimeoutReset : make(chan int),
        }
        server.ChannelID = v.ChannelID
        for _, admin := range v.Admins {
            server.Admins = append(server.Admins, admin)
        }
        go clearOutboundChannelOnInactivity(server)
        Servers[serverName] = server
        log.Println(Servers[serverName])
        log.Println("Linked server", serverName, "to channel", v.ChannelID)
    }
    
	startDiscordBot() // non-blocking
	startHTTPServer() // blocking
}
