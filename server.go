package main

import (
	"log"
	"sync"
	"errors"
	"time"
	"github.com/bwmarrin/discordgo"
)

type ServerList map[string]*Server
var serverList ServerList

type Server struct {
	Name string
	ChannelID string
	Admins DiscordIdentityList
	Muted DiscordIdentityList
	Prefix string
	Outbound chan *Command
	Mux sync.Mutex
	ActiveThread int
	TimeoutSet chan int
	TimeoutReset chan int
}


func init() {
	serverList = make(map[string]*Server)
}


func (serverList ServerList) getServerByName(serverName string) (server *Server, ok bool) {
	server, ok = serverList[serverName]
	return
}


func (serverList ServerList) getServerLinkedToChannel(channelID string) (server *Server, success bool) {
	for _, v := range serverList {
		if v.ChannelID == channelID {
			return v, true
		}
	}
	return
}


func (serverList ServerList) getNumOfLinkedServers() (count int) {
	for _, v := range serverList {
		if v.isLinked() {
			count++
		}
	}
	return
}


func (server *Server) isLinked() bool {
	return server.ChannelID != ""
}


func (server *Server) linkChannelID(channelID string) error {
	if linkedServer, ok := serverList.getServerLinkedToChannel(channelID); ok {
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


func (server *Server) unlinkChannel() (success bool) {
	if server.ChannelID != "" {
		log.Println("Uninked channelID " + server.ChannelID + " from server " + server.Name)
		server.ChannelID = ""
		success = true
	}
	return
}


func (server *Server) isAdmin(member *discordgo.Member) bool {
	if len(Config.Servers[server.Name].Admins) == 0 {
		return true
	}
	return Config.Servers[server.Name].Admins.isInList(member)
}


func (server *Server) isMuted(member *discordgo.Member) bool {
	return server.Muted.isInList(member)
}


func (server *Server) clearOutboundChannel() {
	for {
		select {
			case <- server.Outbound:
			default: return
		}
	}
}


func (server *Server) clearOutboundChannelOnInactivity() {
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
			server.clearOutboundChannel()
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