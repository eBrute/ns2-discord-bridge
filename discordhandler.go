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
	mentionPattern *regexp.Regexp
	channelPattern *regexp.Regexp
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
	mentionPattern, _ = regexp.Compile(`[\\]?<@[!]?\d+>`)
	channelPattern, _ = regexp.Compile(`[\\]?<#\d+>`)

	session.AddHandler(chatEventHandler)

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}
	
	// footer := &discordgo.MessageEmbedFooter{
	// 	Text: "Brute: Long long long long long message",
	// 	IconURL: "https://image.flaticon.com/teams/new/1-freepik.jpg",
	// }
	// image := &discordgo.MessageEmbedImage{
	// 	URL: "https://image.flaticon.com/teams/new/1-freepik.jpg",
	// 	Width: 200,
	// 	Height: 200,
	// }
	// thumb := &discordgo.MessageEmbedThumbnail{
	// 	URL: "https://image.flaticon.com/teams/new/1-freepik.jpg",
	// 	Width: 564,
	// 	Height: 564,
	// }
	// provider := &discordgo.MessageEmbedProvider{
	// 	URL: "https://google.com",
	// 	Name: "providername",
	// }
    // author := &discordgo.MessageEmbedAuthor{
	// 	URL: "https://userurl.com",
	// 	Name: "Brute",
	// 	IconURL: "https://image.flaticon.com/teams/new/1-freepik.jpg",
	// }
    // fields := []*discordgo.MessageEmbedField{{
	//     Name: "embedfieldname1",
	//     Value: "embedfieldvalue1",
	//     Inline: true,
	// },
	// {
	// 	Name: "embedfieldname2",
	// 	Value: "embedfieldvalue2",
	// 	Inline: true,
	// },
	// {
	// 	Name: "embedfieldname3",
	// 	Value: "embedfieldvalue3",
	// 	Inline: false,
	// },
	// {
	// 	Name: "embedfieldname4",
	// 	Value: "embedfieldvalue4",
	// 	Inline: false,
	// }}
	
	// embed := &discordgo.MessageEmbed{
	// 	// URL: "https://google.de",
	// 	// Type: "Type",
	// 	// Title: "Title",
	// 	Description: "Long long long long long message",
	// 	// Timestamp: "2017-01-01T23:59:59",
	// 	// Color: 255*256*256 + 128*256 + 64,
	// 	// Footer: footer,
	// 	// Image: image,
	// 	// Thumbnail: thumb,
	// 	// Provider: provider,
	// 	Author: author,
	// 	// Fields: fields,
	// }
		
	// _, _ = session.ChannelMessageSendEmbed("242940165516034049", embed)
	
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
			Content: formatDiscordMessage(m),
		}
		server.TimeoutSet <- 60 // sec
		server.Outbound <- cmd
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
			if !IsAdminForServer(author, server) {
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
			if !IsAdminForServer(author, server) {
				respond("You are not registered as an admin for server '" + server.Name + "'")
				return
			}
			UnlinkChannelFromServer(server)
			respond("Unlinked this channel")
			
		case "rcon":
			if !IsAdminForServer(author, server) {
				respond("You are not registered as an admin for server '" + server.Name + "'")
				return
			}
			command := strings.Join(fields[1:], " ")
			cmd := Command{
				Type: "rcon",
				User: m.Author.Username,
				Content: command,
			}
			server.TimeoutSet <- 60 // sec
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


func mentionTranslator(mentions []*discordgo.User) (func(string) string) {
	return func(match string) string {
		id := strings.Trim(match, "\\<@!>")
		for _, mention := range mentions {
			if mention.ID == id {
				return "@" + mention.Username
			}
		}
		return match
	}
}


func channelTranslator(mentions []*discordgo.User) (func(string) string) {
	return func(match string) string {
		id := strings.Trim(match, "\\<#>")
		if channel, err := session.State.Channel(id); err == nil {
			return "#" + channel.Name
		} else {
			return "#deleted-channel"
		}
		return match
	}
}


func formatDiscordMessage(m *discordgo.MessageCreate) string {
	message := mentionPattern.ReplaceAllStringFunc(m.Content, mentionTranslator(m.Mentions) )
	message = channelPattern.ReplaceAllStringFunc(message, channelTranslator(m.Mentions) )
	return message
}


func forwardChatMessageToDiscord(serverName string, username string, steamID3 int32, message string) {
	if server, ok := Servers[serverName]; ok {
		
		switch Config.Discord.MessageStyle {
		case "multiline":
			embed := &discordgo.MessageEmbed{
				URL: "https://google.de",
				Description: message,
				// Color: 255*256*256 + 128*256 + 64,
				Author: &discordgo.MessageEmbedAuthor{
					URL: "https://userurl.com",
					Name: username,
					IconURL: GetAvatarForSteamID3(steamID3),
				},
			}
			 _, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
		case "inline":
			embed := &discordgo.MessageEmbed{
				// Color: 255*256*256 + 128*256 + 64,
				Footer: &discordgo.MessageEmbedFooter{
					Text: username +": " + message,
					IconURL: GetAvatarForSteamID3(steamID3),
				},
			}
			 _, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
		
		default: fallthrough
		case "text":
			_, _ = session.ChannelMessageSend(server.ChannelID, server.Prefix + "**" + username + ":** " + message)
		}
	}
}


func forwardGameStatusToDiscord(serverName string, message string) {
	if server, ok := Servers[serverName]; ok {
		_, _ = session.ChannelMessageSend(server.ChannelID, server.Prefix + message)
	}
}


func forwardAdminPrintToDiscord(serverName string, message string) {
	if server, ok := Servers[serverName]; ok {
		_, _ = session.ChannelMessageSend(server.ChannelID, server.Prefix + message)
	}
}