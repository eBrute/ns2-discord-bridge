package main

import (
	"log"
	"strings"
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

	session.AddHandler(chatEventHandler)

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	log.Println("Discord Bot is now running.")
}


func chatEventHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	
	// Ignore all messages created by the bot itself
	if m.Author.ID == botID {
		return
	}
	
	commandMatches := commandPattern.FindStringSubmatch(m.Content)
	
	if len(commandMatches) == 0 { // this is a regular message
		
		server, ok := getServerLinkedToChannel(m.ChannelID)
		if !ok {
			// this channel isnt linked to any server, so just do nothing
			return
		}
		
		cmd := Command{
			Type: "chat",
			User: m.Author.Username,
			Content: m.Content,
		}
		Servers[server].Outbound <- cmd
		// TODO either make sure server is listening or have a timer clear the channel after some time
		return
	}

	// message was a discord command
	fields := strings.Fields(m.Content)
	switch commandMatches[1] {
		case "link":
			if len(fields) < 2 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "You need to specify a server")
				return
			}
			server, ok := getServerLinkedToChannel(m.ChannelID)
			if ok {
				 if server == fields[1] {
					 _, _ = s.ChannelMessageSend(m.ChannelID, "This channel was already linked to '" + server + "'")
				 } else {
					 _, _ = s.ChannelMessageSend(m.ChannelID, "This channel is already linked to '" + server +"'. Use !unlink first.")
				 }
				 return
			}
			linkChannelIDToServer(m.ChannelID, fields[1])
			_, _ = s.ChannelMessageSend(m.ChannelID, "This channel is now linked to '" + fields[1] + "'")
			
		case "unlink":
			if unlinkChannelByChannelID(m.ChannelID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Unlinked this channel")
			} else {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Channel is not linked")
			}
		
		case "list":
			listAll := len(fields) > 1  && fields[1] == "all"
			for server, channel := range Servers {
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
		Servers[server] = CreateChannel(server, channelID)
		log.Println("Linked channelID " + channelID + " to server " + server)
}


func getServerLinkedToChannel(channelID string) (server string, success bool) {
	for k, v := range Servers {
		if v.ChannelID == channelID {
			return k, true
		}
	}
	return
}


func unlinkChannelByChannelID(ID string) (success bool) {
	for server, channel := range Servers {
		if channel.ChannelID == ID {
			log.Println("Uninked channelID " + channel.ChannelID + " from server " + server)
			delete(Servers, server)
			return true
		}
	}
	return
}


func getHelpMessage() string {
	return "```" + `
!help                            - prints this help
!commands                        - prints this help
!link <server>                   - links server to this channel
!unlink                          - unlinks this channel
!list                            - prints the server linked to this channel
!list all                        - prints all linked servers
` + "```"
}


func forwardMessageToDiscord(server string, username string, message string) {
		channel, ok := Servers[server]
		if !ok {
			log.Println("Could not get a channel for", server, ". Link a channel first with '!link <servername>'")
			return
		}
	
		_, _ = session.ChannelMessageSend(channel.ChannelID, "**" + username + ":** " + message)
}
