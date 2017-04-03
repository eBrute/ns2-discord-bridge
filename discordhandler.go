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


func getResponseFunction(s *discordgo.Session, m *discordgo.MessageCreate) func(string) {
    return func(text string) {
        _, _ = s.ChannelMessageSend(m.ChannelID, text)
    }
}


func chatEventHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	
	// Ignore all messages created by the bot itself
	author := m.Author
	if author.ID == botID {
		return
	}
	
	commandMatches := commandPattern.FindStringSubmatch(m.Content)
	
	if len(commandMatches) == 0 { // this is a regular message
		
		server, ok := GetServerLinkedToChannel(m.ChannelID)
		if !ok {
			// this channel isnt linked to any server, so just do nothing
			return
		}
		
		cmd := Command{
			Type: "chat",
			User: author.Username,
			Content: m.Content,
		}
		server.Outbound <- cmd
		// TODO either make sure server is listening or have a timer clear the channel after some time
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
			server, ok := GetServerByName(fields[1])
			if !ok {
				respond("The server '" + fields[1] + "' is not configured")
				return
			}
			if !GetIsAdminForServer(author, server) {
				respond("You are not registered as an admin for server '" + server.Name + "'")
				return
			}
			if err := LinkChannelIDToServer(m.ChannelID, server); err != nil {
				respond(err.Error())
			} else {
				respond("This channel is now linked to '" + server.Name + "'")
			}
			return

		case "list":
			listAll := len(fields) > 1  && fields[1] == "all"
			for _, server := range Servers {
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
	server, isServerLinked := GetServerLinkedToChannel(m.ChannelID)
	if !isServerLinked {
		respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}
	
	switch commandMatches[1] {
		case "unlink":
			if !GetIsAdminForServer(author, server) {
				respond("You are not registered as an admin for server '" + server.Name + "'")
				return
			}
			UnlinkChannelFromServer(server)
			respond("Unlinked this channel")
			
		case "rcon":
			if !GetIsAdminForServer(author, server) {
				respond("You are not registered as an admin for server '" + server.Name + "'")
				return
			}
			command := strings.Join(fields[1:], " ")
			cmd := Command{
				Type: "rcon",
				User: m.Author.Username,
				Content: command,
			}
			server.Outbound <- cmd
		
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

func IsAdminOfServer(user *discordgo.User, server *Server) bool {
	userHandle := user.Username + "#" + user.Discriminator
	userID := user.ID
	for _, v := range server.Admins {
		if v == userHandle || v == userID {
			return true
		}
	}
	return false
}


func forwardMessageToDiscord(serverName string, username string, message string) {
	if server, ok := Servers[serverName]; ok {
		_, _ = session.ChannelMessageSend(server.ChannelID, "**" + username + ":** " + message)
	}
}


func forwardGameStatusToDiscord(serverName string, message string) {
	if server, ok := Servers[serverName]; ok {
		_, _ = session.ChannelMessageSend(server.ChannelID, message)
	}
}


func forwardAdminPrintToDiscord(serverName string, message string) {
	if server, ok := Servers[serverName]; ok {
		_, _ = session.ChannelMessageSend(server.ChannelID, message)
	}
}