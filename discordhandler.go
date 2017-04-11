package main

import (
	"log"
	"strings"
	"strconv"
	"regexp"
	"github.com/bwmarrin/discordgo"
)

var (
	botID string
	session *discordgo.Session
	commandPattern *regexp.Regexp
)

type ResponseHandler struct{
	respond func(string)
	s *discordgo.Session
	m *discordgo.MessageCreate
	message []string
}


func init() {
	commandPattern, _ = regexp.Compile(`^!(\w+)(\s|$)`)
}


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

	session.AddHandler(chatEventHandler)

	// open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}
	
	log.Println("Discord Bot is now running.")
}


func createResponseHandler(s *discordgo.Session, m *discordgo.MessageCreate, message []string) *ResponseHandler {
	return &ResponseHandler{
		func(text string) {
			_, _ = s.ChannelMessageSend(m.ChannelID, text)
		},
		s,
		m,
		message,
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
		server, isServerLinked := serverList.getServerLinkedToChannel(m.ChannelID)
		if !isServerLinked {
			// this channel isnt linked to any server, so just do nothing
			return
		}
		
		if server.isMuted(author) {
			return
		}
		
		server.TimeoutSet <- 60 // sec
		server.Outbound <- createChatMessageCommand(author.Username, m)
		return
	}

	// message was a discord command
	messageFields := strings.Fields(m.Content)[1:]
	responseHandler := createResponseHandler(s, m, messageFields)
	
	// first handle the commands that dont require a linked server
	switch commandMatches[1] {
		case "link": responseHandler.linkChannel()
		case "list": responseHandler.listChannel()
		case "unlink": responseHandler.unlinkChannel()
		case "mute": responseHandler.muteUser()
		case "unmute": responseHandler.unmuteUser()
		case "rcon": responseHandler.sendRconCommand()
		case "info":  responseHandler.requestServerInfo()
		case "status": responseHandler.requestServerStatus()
		default : fallthrough
		case "commands": fallthrough
		case "help": responseHandler.printHelpMessage()
	}
}


func (r *ResponseHandler) linkChannel() {
	if len(r.message) < 1 {
		r.respond("You need to specify a server")
		return
	}
	server, ok := serverList.getServerByName(r.message[0])
	if !ok {
		r.respond("The server '" + r.message[0] + "' is not configured")
		return
	}
	if !server.isAdmin(r.m.Author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}
	if err := server.linkChannelID(r.m.ChannelID); err != nil {
		r.respond(err.Error())
	} else {
		r.respond("This channel is now linked to '" + server.Name + "'")
	}
}


func (r *ResponseHandler) unlinkChannel() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.m.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	if !server.isAdmin(r.m.Author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}
	server.unlinkChannel()
	r.respond("Unlinked this channel")
}


func (r *ResponseHandler) listChannel() {
	listAll := len(r.message) > 0 && r.message[0] == "all"
	for _, server := range serverList {
		id := server.ChannelID
		if listAll || id == r.m.ChannelID {
			r.respond("Server '" + server.Name + "' is linked to channel <#" + id + "> (" + id + ")")
		}
	}
}


func (r *ResponseHandler) muteUser() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.m.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	if !server.isAdmin(r.m.Author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}
	
	count := 0
	for _, mention := range r.m.Mentions {
		if !server.isMuted(mention) {
			server.Muted = append(server.Muted, mention.ID)
			count++
		}
	}
	r.respond("Muted " + strconv.Itoa(count) + " users")
}


func (r *ResponseHandler) unmuteUser() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.m.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	if !server.isAdmin(r.m.Author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}

	count := 0
	for _, mentionedUser := range r.m.Mentions {
		for i, mutedUser := range server.Muted {
			if isSameUser(mentionedUser, mutedUser) {
				server.Muted = append(server.Muted[:i], server.Muted[i+1:]...)
				count++
				log.Println("Muted user", "'" + mentionedUser.Username + "#" + mentionedUser.Discriminator + "'", "id:", mentionedUser.ID)
			}
		}
	}
	r.respond("Unmuted " + strconv.Itoa(count) + " user(s)")
}


func (r *ResponseHandler) requestServerStatus() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.m.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	server.TimeoutSet <- 60 // sec
	server.Outbound <- createServerStatusCommand()
}


func (r *ResponseHandler) requestServerInfo() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.m.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	server.TimeoutSet <- 60 // sec
	server.Outbound <- createServerInfoCommand()
}


func (r *ResponseHandler) sendRconCommand() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.m.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	if !server.isAdmin(r.m.Author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}
	command := strings.Join(r.message[:], " ")
	server.TimeoutSet <- 60 // sec
	server.Outbound <- createRconCommand(r.m.Author.Username, command)
}


func (r *ResponseHandler) printHelpMessage() {
	r.respond("```" + `
!help                    - prints this help
!commands                - prints this help
!list                    - prints the server linked to this channel
!list all                - prints all linked servers
!status                  - prints a short server status
!info                    - prints a long server info

admin commands:
!link <server>           - links server to this channel
!unlink                  - unlinks this channel
!mute @discorduser(s)    - dont forward messages from user(s) to the server
!unmute @discorduser(s)  - remove user(s) from being muted
!rcon <console commands> - executes console commands directly on the linked server
` + "```")
}
