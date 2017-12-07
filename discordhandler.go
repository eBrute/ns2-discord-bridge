// This file contains functions related to handling input coming from the Discord side

package main

import (
	"log"
	"strings"
	"strconv"
	"errors"
	"regexp"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"net/url"
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
		server, isServerLinked := serverList.getServerByChannelID(m.ChannelID)
		if !isServerLinked {
			// this channel isn't linked to any server, so just do nothing
			return
		}

		if server.isMuted(authorMember) {
			return
		}
		nick := getMemberNickname(authorMember)
		v := url.Values {}
		v.Set("request", "discordsend")
		v.Set("user", nick)
		v.Set("msg", formatDiscordMessage(m))
		_, err := http.PostForm(server.Config.WebAdmin, v)

		if err != nil {
			log.Println(err.Error())
		}

		return
	}

	// message was a discord command
	messageFields := strings.Fields(m.Content)[1:]
	responseHandler := createResponseHandler(s, m, messageFields)

	// first handle the commands that dont require a linked server
	switch commandMatches[1] {
		case "mute": responseHandler.muteUser()
		case "unmute": responseHandler.unmuteUser()
		case "rcon": responseHandler.sendRconCommand()
		case "info":  responseHandler.requestServerInfo()
		case "status": responseHandler.requestServerStatus()
		case "channelinfo": responseHandler.printChannelInfo()
		case "version": responseHandler.printVersion()
		default : fallthrough
		case "commands": fallthrough
		case "help": responseHandler.printHelpMessage()
	}
}


func (r *ResponseHandler) printChannelInfo() {
	response := make([]string, 6)
	response = append(response, "```")
	channel, _ := r.session.State.Channel(r.message.ChannelID)
	response = append(response, "Channel '" + channel.Name + "' Id: " + channel.ID)
	guild, _ := r.session.Guild(channel.GuildID)
	response = append(response, "Guild '" + guild.Name + "' Id: " + guild.ID)
	for _, role := range guild.Roles {
		response = append(response, "Role '" + role.Name + "' Id: " + role.ID)
	}
	for _, server := range serverList {
		id := server.Config.ChannelID
		linkedChannel, err := r.session.State.Channel(id)
		name := "<unknown channel>"
		if err == nil {
			name = "<#" + linkedChannel.Name + ">"
		}
		response = append(response, "Server '" + server.Name + "' <-> Channel " + name + " (" + id + ")")
	}
	response = append(response, "```")
	r.respond(strings.Join(response, "\n"))
}


func (r *ResponseHandler) printVersion() {
	r.respond("Version " + version)
}


func (r *ResponseHandler) muteUser() {
	server, isServerLinked := serverList.getServerByChannelID(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server.")
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
	server, isServerLinked := serverList.getServerByChannelID(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server.")
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
	server, isServerLinked := serverList.getServerByChannelID(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server.")
		return
	}

	resp, err := http.PostForm(server.Config.WebAdmin, url.Values {
		"request": {"discordinfo"},
	})

	if err != nil {
		log.Println(err.Error())
	}

	serverInfo := ServerInfo {}
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &serverInfo)
	forwardServerStatusToDiscord(server, MessageType{GroupType: "info", SubType: "status"}, serverInfo)
}


func (r *ResponseHandler) requestServerInfo() {
	server, isServerLinked := serverList.getServerByChannelID(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server.")
		return
	}

	resp, err := http.PostForm(server.Config.WebAdmin, url.Values {
		"request": {"discordinfo"},
	})

	if err != nil {
		log.Println(err.Error())
	}

	serverInfo := ServerInfo {}
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &serverInfo)
	forwardServerStatusToDiscord(server, MessageType{GroupType: "info", SubType: "info"}, serverInfo)
}


func (r *ResponseHandler) sendRconCommand() {
	server, isServerLinked := serverList.getServerByChannelID(r.message.ChannelID)
	if !isServerLinked {
		r.respond("Channel is not linked to any server.")
		return
	}

	if !server.isAdmin(r.author) {
		r.respond("You are not registered as an admin for server '" + server.Name + "'")
		return
	}

	_, err := http.PostForm(server.Config.WebAdmin, url.Values {
		"command": {strings.Join(r.messageContent[:], " ")},
	})

	if err != nil {
		log.Println(err.Error())
	}
}


func (r *ResponseHandler) printHelpMessage() {
	r.respond("```" + `
!help					 - prints this help
!commands				 - prints this help
!status					 - prints a short server status
!info					 - prints a long server info
!channelinfo			 - prints ids of the current channel, guild and roles
!version				 - prints the version number

admin commands:
!mute @discorduser(s)	 - dont forward messages from user(s) to the server
!unmute @discorduser(s)  - remove user(s) from being muted
!rcon <console commands> - executes console commands directly on the linked server
` + "```")
}
