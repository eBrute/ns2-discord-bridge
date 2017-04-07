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
    ChatMessageReadyRoomColor []int
    ChatMessageMarineColor []int
    ChatMessageAlienColor []int
    ChatMessageSpectatorColor []int
}

type MessageStyleTextConfig struct {
    ChatMessageFormat string
    ChatMessageReadyRoomPrefix string
    ChatMessageMarinePrefix string
    ChatMessageAlienPrefix string
    ChatMessageSpectatorPrefix string
    PlayerJoinFormat string
    PlayerLeaveFormat string
}

type ServerConfig struct {
    ChannelID string
    Admins []string
    ServerChatMessagePrefix string
    ServerStatusMessagePrefix string
    ServerIconUrl string
}

var Config Configuration


func (config *Configuration) getColor(color []int, defaultColor int) int {
    if len(color) != 3 {
        return defaultColor
    }
    return color[0]*256*256 + color[1]*256 + color[2]
}
    
    
func (config *Configuration) loadConfig(configFile string) {
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
    
    for serverName, v := range Config.Servers {
        server := &Server{
            Name : serverName,
            Admins : make([]string, 0),
            Outbound : make(chan *Command),
            TimeoutSet : make(chan int),
            TimeoutReset : make(chan int),
        }
        server.ChannelID = v.ChannelID
        for _, admin := range v.Admins {
            server.Admins = append(server.Admins, admin)
        }
        serverList[serverName] = server
        log.Println("Linked server '"+ serverName +"' to channel", v.ChannelID)
    }
}
