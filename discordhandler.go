package main

import (
	// "fmt"
	"log"
	"strings"
	"strconv"
	// "errors"
	"regexp"
	"github.com/bwmarrin/discordgo"
)

var (
	botID string
	session *discordgo.Session
	commandPattern *regexp.Regexp
)


func startDiscordBot() {
	
	var err error
	session, err = discordgo.New("Bot " + Config.Discord.Token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	// Get the account information.
	user, err := session.User("@me")
	if err != nil {
		log.Println("error obtaining account details,", err)
	}

	botID = user.ID
	commandPattern, _ = regexp.Compile(`^!(\w+)(\s|$)`)

	session.AddHandler(chatCommandHandler)

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	log.Println("Discord Bot is now running.")
}


func chatCommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == botID {
		return
	}

	commandMatches := commandPattern.FindStringSubmatch(m.Content)
	
	if len(commandMatches) == 0 {
		// this is a regular message
		return
	}

	fields := strings.Fields(m.Content)
	switch commandMatches[1] {
		
		case "link":
			if len(fields) < 2 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "You need to specify a server")
				return
			}
			for _, server := range fields[1:] {
				linkChannelIDToServer(m.ChannelID, server)
				_, _ = s.ChannelMessageSend(m.ChannelID, "This channel is now linked to " + server)
			}
			
		case "unlink":
			if len(fields) > 1 {
				count := unlinkChannelByServername(fields[1:])
				_, _ = s.ChannelMessageSend(m.ChannelID, "Unlinked " + strconv.Itoa(count) +" channel(s)")
			} else {
				if unlinkChannelByChannelID(m.ChannelID) > 0 {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Unlinked this channel")
				} else {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Channel was not linked")
				}
			}
		
		case "list":
			listAll := len(fields) > 1  && fields[1] == "all"
			for server, channel := range Config.Servers {
				id := channel.ChannelID
				if listAll || id == m.ChannelID {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Server '" + server + "' is linked to channel <#" + id + "> (" + id + ")")
				}
			}
		
		case "help": fallthrough
		case "commands": fallthrough
		default:
			_, _ = s.ChannelMessageSend(m.ChannelID, getHelpMessage())
	}
}


func linkChannelIDToServer(channelID string, server string) {
		Config.Servers[server] = Channel{channelID}
		log.Println("Linked channelID " + channelID + " to server " + server)
}


func unlinkChannelByServername(servers []string) (count int) {
	for _, server := range servers {
		if _, ok := Config.Servers[server]; ok {
			log.Println("Uninked channelID " + Config.Servers[server].ChannelID + " from server " + server)
			delete(Config.Servers, server)
			count++
		}
	}
	return
}


func unlinkChannelByChannelID(ID string) (count int) {
	for server, channel := range Config.Servers {
		if channel.ChannelID == ID {
			log.Println("Uninked channelID " + channel.ChannelID + " from server " + server)
			delete(Config.Servers, server)
			count++
		}
	}
	return
}


func getHelpMessage() string {
	return "```" + `
!help                            - prints this help
!commands                        - prints this help
!link <server>                   - links server to this channel
!unlink <server> [<server2> ..]  - unlinks server(s) from this channel
!unlink                          - unlinks all servers from this channel
!list                            - prints all servers linked to this channel
!list all                        - prints all linked servers
` + "```"
}


func forwardMessageToDiscord(server string, username string, message string) {
		channel, ok := Config.Servers[server]
		if !ok {
			log.Println("Could not get a channel for", server, ". Link a channel first with '!link <servername>'")
			return
		}
	
		_, _ = session.ChannelMessageSend(channel.ChannelID, "**" + username + ":** " + message)
}
