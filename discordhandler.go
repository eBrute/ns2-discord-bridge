package main

import (
	"log"
	"strings"
	"strconv"
	"errors"
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
	session *discordgo.Session
	message *discordgo.MessageCreate
	guild *discordgo.Guild
	author *discordgo.Member
	messageContent []string
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

	session.UpdateStatus(0, "")
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
	guild, err := getGuildForChannel(s, m.ChannelID)
	if err != nil {
		panic(err.Error())
	}
	author, _ := s.State.Member(guild.ID, m.Author.ID)
	return &ResponseHandler{
		func(text string) {
			_, _ = s.ChannelMessageSend(m.ChannelID, text)
		},
		s,
		m,
		guild,
		author,
		message,
	}
}


func getGuildForChannel(s *discordgo.Session, channelID string) (*discordgo.Guild, error) {
	for _, guild := range s.State.Guilds {
		channels, _ := s.GuildChannels(guild.ID)
		for _, channel := range channels {
			if channel.ID == channelID {
				return guild, nil
			}
		}
	}
	return nil, errors.New("No guild for channel '" + channelID + "'")
}


func getUserNickname(user *discordgo.User, guild *discordgo.Guild) string {
	if member, err := session.State.Member(guild.ID, user.ID); err == nil {
		return getMemberNickname(member)
	}
	return user.Username
}


func getMemberNickname(member *discordgo.Member) string {
	if member.Nick != "" {
		return member.Nick
	}
	return member.User.Username
}


func chatEventHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	
	// ignore all messages created by the bot itself
	author := m.Author
	if author.ID == botID {
		return
	}
	
	guild, err := getGuildForChannel(s, m.ChannelID)
	if err != nil {
		panic(err.Error())
	}
	authorMember, err := s.State.Member(guild.ID, author.ID)
	if err != nil {
		// ignore non-member messages
		return
	}
	
	commandMatches := commandPattern.FindStringSubmatch(m.Content)
	
	if len(commandMatches) == 0 { // this is a regular message
		server, isServerLinked := serverList.getServerLinkedToChannel(m.ChannelID)
		if !isServerLinked {
			// this channel isnt linked to any server, so just do nothing
			return
		}
		
		if server.isMuted(authorMember) {
			return
		}
		nick := getMemberNickname(authorMember)
		server.TimeoutSet <- 60 // sec
		server.Outbound <- createChatMessageCommand(nick, m)
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
	if len(r.messageContent) < 1 {
		r.respond("You need to specify a server")
		return
	}
	server, ok := serverList.getServerByName(r.messageContent[0])
	if !ok {
		r.respond("The server '" + r.messageContent[0] + "' is not configured")
		return
	}
	if !server.isAdmin(r.author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}
	if err := server.linkChannelID(r.message.ChannelID); err != nil {
		r.respond(err.Error())
	} else {
		r.respond("This channel is now linked to '" + server.Name + "'")
	}
}


func (r *ResponseHandler) unlinkChannel() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	if !server.isAdmin(r.author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}
	server.unlinkChannel()
	r.respond("Unlinked this channel")
}


func (r *ResponseHandler) listChannel() {
	listAll := len(r.messageContent) > 0 && r.messageContent[0] == "all"
	for _, server := range serverList {
		id := server.ChannelID
		if listAll || id == r.message.ChannelID {
			r.respond("Server '" + server.Name + "' is linked to channel <#" + id + "> (" + id + ")")
		}
	}
}


func (r *ResponseHandler) muteUser() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	if !server.isAdmin(r.author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}
	
	count := 0
	for _, mention := range r.message.Mentions {
		mentionedMember, err := r.session.State.Member(r.guild.ID, mention.ID)
		if err == nil && !server.isMuted(mentionedMember) {
			server.Muted = append(server.Muted, DiscordIdentity(mention.ID))
			count++
		}
	}
	r.respond("Muted " + strconv.Itoa(count) + " users")
}


func (r *ResponseHandler) unmuteUser() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	if !server.isAdmin(r.author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}

	count := 0
	for _, mentionedUser := range r.message.Mentions {
		for i, mutedUser := range server.Muted {
			if mutedUser.matchesUser(mentionedUser) {
				server.Muted = append(server.Muted[:i], server.Muted[i+1:]...)
				count++
				log.Println("Muted user", "'" + mentionedUser.Username + "#" + mentionedUser.Discriminator + "'", "id:", mentionedUser.ID)
			}
		}
	}
	r.respond("Unmuted " + strconv.Itoa(count) + " user(s)")
}


func (r *ResponseHandler) requestServerStatus() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	server.TimeoutSet <- 60 // sec
	server.Outbound <- createServerStatusCommand()
}


func (r *ResponseHandler) requestServerInfo() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	server.TimeoutSet <- 60 // sec
	server.Outbound <- createServerInfoCommand()
}


func (r *ResponseHandler) sendRconCommand() {
	server, isServerLinked := serverList.getServerLinkedToChannel(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server. Use !link <servername> first.")
		return
	}

	if !server.isAdmin(r.author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}
	command := strings.Join(r.messageContent[:], " ")
	server.TimeoutSet <- 60 // sec
	server.Outbound <- createRconCommand(r.message.Author.Username, command)
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
