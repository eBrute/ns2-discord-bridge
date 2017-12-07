// This file contains functions for managing servers.
// A server is a connection between the game server and a specific Discord channel
// This includes establishing the connection, muting users and timeouts for messages

package main

import (
	"github.com/bwmarrin/discordgo"
)

type ServerList map[string]*Server
var serverList ServerList

type Server struct {
	Name         string
	Config       ServerConfig
	Admins       DiscordIdentityList
	Muted        DiscordIdentityList
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
