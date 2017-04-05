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
        MessageStyle string
    }
    Messagestyles struct {
        Rich MessageStyleRichConfig
        Text MessageStyleTextConfig
    }
    Httpserver struct {
        Address string
    }
    Steam struct {
        WebApiKey string
    }
    Servers map[string]ServerConfig
}

type MessageStyleRichConfig struct {
    PlayerJoinColor []int
    PlayerLeaveColor []int
    StatusColor []int
    ChatMessageColor []int
    ReadyRoomChatMessageColor []int
    MarineChatMessageColor []int
    AlienChatMessageColor []int
    SpectatorChatMessageColor []int
}

type MessageStyleTextConfig struct {
    ChatMessageFormat string
}

type ServerConfig struct {
    ChannelID string
    Admins []string
    ChatMessagePrefix string
    StatusMessagePrefix string
    ServerIconUrl string
}

var Config Configuration
var configFile string
var DefaultMessageColor int = 75*256*256 + 78*256 + 82

var Servers map[string]*Server

type Server struct {
    Name string
    ChannelID string
    Admins []string
    Prefix string
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


func UnlinkChannelFromServer(server *Server) (success bool) {
    if server != nil {
        log.Println("Uninked channelID " + server.ChannelID + " from server " + server.Name)
        server.ChannelID = ""
        success = true
    }
    return
}


func IsAdminForServer(user *discordgo.User, server *Server) bool {
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


func GetColorForMessage(messagetype string) int {
    switch messagetype {
        case "chat" :        return getColorFromConfig(Config.Messagestyles.Rich.ChatMessageColor)
        case "playerjoin" :  return getColorFromConfig(Config.Messagestyles.Rich.PlayerJoinColor)
        case "playerleave" : return getColorFromConfig(Config.Messagestyles.Rich.PlayerLeaveColor)
        case "status" :      return getColorFromConfig(Config.Messagestyles.Rich.StatusColor)
        case "adminprint" :  return getColorFromConfig(Config.Messagestyles.Rich.StatusColor)
        default :            return DefaultMessageColor
    }
}

func GetTeamColorForChatMessage(teamNumber int) int {
    println("getcolor", teamNumber)
    switch teamNumber {
        default: fallthrough
        case 0 : return getColorFromConfig(Config.Messagestyles.Rich.ReadyRoomChatMessageColor)
        case 1 : return getColorFromConfig(Config.Messagestyles.Rich.MarineChatMessageColor)
        case 2 : return getColorFromConfig(Config.Messagestyles.Rich.AlienChatMessageColor)
        case 3 : return getColorFromConfig(Config.Messagestyles.Rich.SpectatorChatMessageColor)
    }
}


func getColorFromConfig(color []int) int {
    if len(color) != 3 {
        println("no color specified")
        return DefaultMessageColor
    }
    return color[0]*256*256 + color[1]*256 + color[2]
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
        log.Println("Linked server '"+ serverName +"' to channel", v.ChannelID)
    }
    
    InitSteamBinding()
	startDiscordBot() // non-blocking
	startHTTPServer() // blocking
}
