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

	// open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}
	
	log.Println("Discord Bot is now running.")
}


func getResponseFunction(s *discordgo.Session, m *discordgo.MessageCreate) func(string) {
    return func(text string) {
        _, _ = s.ChannelMessageSend(m.ChannelID, text)
    }
}


func chatEventHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	
	// ignore all messages created by the bot itself
	author := m.Author
	if author.ID == botID {
		return
	}
	
	commandMatches := commandPattern.FindStringSubmatch(m.Content)
	
	if len(commandMatches) == 0 { // this is a regular message
		server, ok := serverList.getServerLinkedToChannel(m.ChannelID)
		if !ok {
			// this channel isnt linked to any server, so just do nothing
			return
		}
		server.TimeoutSet <- 60 // sec
		server.Outbound <- createChatMessageCommand(author.Username, m)
		return
	}

	// message was a discord command
	fields := strings.Fields(m.Content)
	respond := getResponseFunction(s, m)
	
	// first handle the commands that dont require a linked server
	switch commandMatches[1] {
		case "link":
			if len(fields) < 2 {
				respond("You need to specify a server")
				return
			}
			server, ok := serverList.getServerByName(fields[1])
			if !ok {
				respond("The server '" + fields[1] + "' is not configured")
				return
			}
			if !server.isAdmin(author) {
				respond("You are not registered as an admin for server '" + server.Name + "'")
				return
			}
			if err := server.linkChannelID(m.ChannelID); err != nil {
				respond(err.Error())
			} else {
				respond("This channel is now linked to '" + server.Name + "'")
			}
			return

		case "list":
			listAll := len(fields) > 1  && fields[1] == "all"
			for _, server := range serverList {
				id := server.ChannelID
				if listAll || id == m.ChannelID {
					respond("Server '" + server.Name + "' is linked to channel <#" + id + "> (" + id + ")")
				}
			}
			return
			
		case "help": fallthrough
		case "commands":
			respond(getHelpMessage())
			return
	}
	
	// now handle the commands that require a linked server
	server, isServerLinked := serverList.getServerLinkedToChannel(m.ChannelID)
	if !isServerLinked {
		respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}
	
	switch commandMatches[1] {
		case "unlink":
			if !server.isAdmin(author) {
				respond("You are not registered as an admin for server '" + server.Name + "'")
				return
			}
			server.unlinkChannel()
			respond("Unlinked this channel")
			
		case "rcon":
			if !server.isAdmin(author) {
				respond("You are not registered as an admin for server '" + server.Name + "'")
				return
			}
			command := strings.Join(fields[1:], " ")
			server.TimeoutSet <- 60 // sec
			server.Outbound <- createRconCommand(m.Author.Username, command)
		
		default:
			respond(getHelpMessage())
	}
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
