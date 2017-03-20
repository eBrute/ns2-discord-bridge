package main

import (
	"fmt"
	// "log"
	"strings"
	"strconv"
	// "errors"
	"github.com/bwmarrin/discordgo"
)

var (
	BotID string
	Token string
	session *discordgo.Session
	channels map[string]string // maps servers to channelIds
)


func startDiscordBot() {
	
	var err error
	session, err = discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Get the account information.
	user, err := session.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	BotID = user.ID
	channels = make(map[string]string)

	session.AddHandler(chatCommandHandler)

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
}


func chatCommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	
	// Ignore all messages created by the bot itself
	if m.Author.ID == BotID {
		return
	}

	if strings.HasPrefix(m.Content, "!link") {
		fields := strings.Fields(m.Content)
		if len(fields) < 2 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "You need to specify a server")
			return
		}
		server := fields[1]
		channels[server] = m.ChannelID
		_, _ = s.ChannelMessageSend(m.ChannelID, "This channel is now linked to " + server)
		return
	}
	
	if strings.HasPrefix(m.Content, "!unlink") {
		fields := strings.Fields(m.Content)
		count := 0
		if len(fields) > 1 {
			for i := 1; i < len(fields); i++ {
				server := fields[i]
				if _, ok := channels[server]; ok {
					delete(channels, server)
					count++
				}
			}
			
			_, _ = s.ChannelMessageSend(m.ChannelID, "Unlinked " + strconv.Itoa(count) +" channel(s)")
		} else {
			for server, channelID := range channels {
				if channelID == m.ChannelID {
					delete(channels, server)
					count++
				}
			}
			if count > 0 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Unlinked this channel")
			} else {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Channel was not linked")
			}
		}
		return
	}
	
	if strings.HasPrefix(m.Content, "!help") || strings.HasPrefix(m.Content, "!commands") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "```" +`
!help                            - prints this help
!commands                        - prints this help
!link <server>                   - links server to this channel
!unlink <server> [<server2> ..]  - unlinks server(s) from this channel
!unlink                          - unlinks all servers from this channel
`+"```")
		return
	}
}


func forwardMessage(server string, username string, message string) {
		channelID, ok := channels[server]
		if !ok {
			fmt.Println("Could not get a channel for", server, ". Link a channel first with '!link <servername>'")
			return
		}
	
		_, _ = session.ChannelMessageSend(channelID, username + ": " + message)
}
