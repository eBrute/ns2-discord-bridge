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
	BotID string
	Token string
	session *discordgo.Session
	channels map[string]string // maps servers to channelIds
	commandPattern *regexp.Regexp
)


func startDiscordBot() {
	
	var err error
	session, err = discordgo.New("Bot " + Token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	// Get the account information.
	user, err := session.User("@me")
	if err != nil {
		log.Println("error obtaining account details,", err)
	}

	BotID = user.ID
	channels = make(map[string]string)
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
	if m.Author.ID == BotID {
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
		
		case "help": fallthrough
		case "commands": fallthrough
		default:
			_, _ = s.ChannelMessageSend(m.ChannelID, getHelpMessage())
	}
}


func linkChannelIDToServer(channelID string, server string) {
		channels[server] = channelID
		log.Println("Linked channelID " + channelID + " to server " + server)
}


func unlinkChannelByServername(servers []string) (count int) {
	for _, server := range servers {
		if _, ok := channels[server]; ok {
			log.Println("Uninked channelID " + channels[server] + " from server " + server)
			delete(channels, server)
			count++
		}
	}
	return
}


func unlinkChannelByChannelID(ID string) (count int) {
	for server, channelID := range channels {
		if channelID == ID {
			log.Println("Uninked channelID " + channelID + " from server " + server)
			delete(channels, server)
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
` + "```"
}


func forwardMessageToDiscord(server string, username string, message string) {
		channelID, ok := channels[server]
		if !ok {
			log.Println("Could not get a channel for", server, ". Link a channel first with '!link <servername>'")
			return
		}
	
		_, _ = session.ChannelMessageSend(channelID, "**" + username + ":** " + message)
}
