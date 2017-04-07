package main

import (
    "log"
    "flag"
    "sync"
    "errors"
    "time"
    "github.com/bwmarrin/discordgo"
)

var configFile string
var Servers map[string]*Server

type Server struct {
    Name string
    ChannelID string
    Admins []string
    Prefix string
    Outbound chan *Command
    Mux sync.Mutex
    ActiveThread int
    TimeoutSet chan int
    TimeoutReset chan int
}


func getServerByName(serverName string) (server *Server, ok bool) {
    server, ok = Servers[serverName]
    return
}


func linkChannelIDToServer(channelID string, server *Server) error {
    if linkedServer, ok := getServerLinkedToChannel(channelID); ok {
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


func getServerLinkedToChannel(channelID string) (server *Server, success bool) {
    for _, v := range Servers {
        if v.ChannelID == channelID {
            return v, true
        }
    }
    return
}


func unlinkChannelFromServer(server *Server) (success bool) {
    if server != nil {
        log.Println("Uninked channelID " + server.ChannelID + " from server " + server.Name)
        server.ChannelID = ""
        success = true
    }
    return
}


func isAdminForServer(user *discordgo.User, server *Server) bool {
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


func clearOutboundChannel(outbound chan *Command) {
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
    loadConfig(configFile)
	startDiscordBot() // non-blocking
	startHTTPServer() // blocking
}
