package main

import (
	"fmt"
	// "log"
	"strings"
	"errors"
	"github.com/bwmarrin/discordgo"
)

var (
	BotID string
	Token string
	session *discordgo.Session
	channels []Channel
)


type Channel struct {
	channelID string
	server string // the server associated with the channel
}


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

	session.AddHandler(chatCommandHandler)

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	
	channels = make([]Channel, 0, 40)
}


func getChannelForServer(server string) (channelID string, err error) {
	for _, channel := range channels {
		if channel.server == server {
			channelID = channel.channelID
			return
		}
	}
	err =  errors.New("no such server")
	fmt.Printf("%d channels", len(channels))
	return
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
		
		channels = append(channels, Channel{m.ChannelID, fields[1]}) // TODO range check(?)
		_, _ = s.ChannelMessageSend(m.ChannelID, "This channel is now linked to " + fields[1])
		return
	}
}


func forwardMessage(server string, username string, message string) {
		channelID, err := getChannelForServer(server)
		fmt.Println("channelid:", channelID)
		if err != nil {
			fmt.Println("Could not get a channel for", server, ". Link a channel first with '!link <servername>'")
			return
		}
		if session == nil {
			fmt.Println("Could not find a discord session")
			return
		}
		
		
		_, _ = session.ChannelMessageSend(channelID, username + ": " + message)
}
