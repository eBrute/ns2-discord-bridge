// This file contains functions for managing servers.
// A server is a connection between the game server and a specific Discord channel
// This includes establishing the connection, muting users and timeouts for messages

package main

import (
	"sync"
	"time"
	"github.com/bwmarrin/discordgo"
)

type ServerList map[string]*Server
var serverList ServerList

type Server struct {
	Name string
	Config ServerConfig
	Admins DiscordIdentityList
	Muted DiscordIdentityList
	Outbound chan *Command
	Mux sync.Mutex
	ActiveThread int
	TimeoutSet chan int
	TimeoutReset chan int
}


func init() {
	serverList = make(map[string]*Server)
}


func (serverList ServerList) getServerByChannelID(channelID string) (server *Server, success bool) {
	for _, v := range serverList {
		if v.Config.ChannelID == channelID {
			return v, true
		}
	}
	return
}


func (server *Server) isAdmin(member *discordgo.Member) bool {
	return server.Config.Admins.isInList(member)
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