package main

import (
	"strings"
	"strconv"
	"time"
	"github.com/bwmarrin/discordgo"
)

type TeamNumber int
type MessageType struct{
	GroupType string
	SubType string
}


var (
	DefaultMessageColor int = 75*256*256 + 78*256 + 82
	lastMultilineChatMessage *discordgo.Message
)


func (messagetype MessageType) getColor() int {
	msgConfig := Config.Messagestyles.Rich
	switch messagetype.GroupType {
		case "chat":        return Config.getColor(msgConfig.ChatMessageColor, DefaultMessageColor)
		case "player":  
			switch messagetype.SubType {
				case "join":  return Config.getColor(msgConfig.PlayerJoinColor, DefaultMessageColor)
				case "leave": return Config.getColor(msgConfig.PlayerLeaveColor, DefaultMessageColor)
				default:      return Config.getColor(msgConfig.StatusColor, DefaultMessageColor)
			}
		case "adminprint":  return Config.getColor(msgConfig.StatusColor, DefaultMessageColor)
		default:            return DefaultMessageColor
	}
}


func (teamNumber TeamNumber) getColor() int {
	msgConfig := Config.Messagestyles.Rich
	switch teamNumber {
		default: fallthrough
		case 0: return Config.getColor(msgConfig.ChatMessageReadyRoomColor, DefaultMessageColor)
		case 1: return Config.getColor(msgConfig.ChatMessageMarineColor, DefaultMessageColor)
		case 2: return Config.getColor(msgConfig.ChatMessageAlienColor, DefaultMessageColor)
		case 3: return Config.getColor(msgConfig.ChatMessageSpectatorColor, DefaultMessageColor)
	}
}


func (teamNumber TeamNumber) getText() string {
	msgConfig := Config.Messagestyles.Text
	switch teamNumber {
		case 0: return msgConfig.ChatMessageReadyRoomPrefix
		case 1: return msgConfig.ChatMessageMarinePrefix
		case 2: return msgConfig.ChatMessageAlienPrefix
		case 3: return msgConfig.ChatMessageSpectatorPrefix
		default: return ""
	}
}


func getTextToUnicodeTranslator() *strings.Replacer {
	println("yep")
	return strings.NewReplacer(
		"yes", "no",
		":)",  "😃",
		":D",  "😄",
		":(",  "😦",
		":|",  "😐",
		":P",  "😛",
		";)",  "😉",
		";(",  "😭",
		">:(", "😠",
		":,(", "😢",
		"<3",  "❤",
		"</3", "💔",
	)
}


func buildTextChatMessage(serverName string, username string, teamNumber TeamNumber, message string) string {
	serverConfig := Config.Servers[serverName]
	messageFormat := Config.Messagestyles.Text.ChatMessageFormat
	teamSpecificString := teamNumber.getText()
	serverSpecificString := serverConfig.ServerChatMessagePrefix
	replacer := strings.NewReplacer("%p", username, "%m", message, "%t", teamSpecificString, "%s", serverSpecificString)
	formattedMessage := replacer.Replace(messageFormat)
	return formattedMessage
}


func buildTextPlayerEvent(serverName string, messagetype MessageType, username string, message string) string {
	serverConfig := Config.Servers[serverName]
	messageConfig := Config.Messagestyles.Text
	messageFormat := "%s %p %m"
	switch messagetype.SubType {
		case "join": messageFormat = messageConfig.PlayerJoinFormat
		case "leave": messageFormat = messageConfig.PlayerLeaveFormat
	}
	serverSpecificString := serverConfig.ServerChatMessagePrefix
	replacer := strings.NewReplacer("%p", username, "%m", message, "%s", serverSpecificString)
	formattedMessage := replacer.Replace(messageFormat)
	return formattedMessage
}


func getLastMessageID(channelID string) (string, bool) {
	messages, _ := session.ChannelMessages(channelID, 1, "", "")
	if len(messages) == 1 {
		return messages[0].ID, true
	}
	return "", false
}


func forwardChatMessageToDiscord(serverName string, username string, steamID SteamID3, teamNumber TeamNumber, message string) {
	if server, ok := serverList[serverName]; ok {
		translatedMessage := getTextToUnicodeTranslator().Replace(message)
		switch Config.Discord.MessageStyle {
		default: fallthrough
		case "multiline":
			lastMessageID, ok := getLastMessageID(server.ChannelID);
			if ok && lastMultilineChatMessage != nil {
				lastEmbed := lastMultilineChatMessage.Embeds[0]
				lastAuthor := lastEmbed.Author
				if  lastMessageID == lastMultilineChatMessage.ID &&
					lastEmbed.Color == teamNumber.getColor() &&
					lastAuthor.Name == username &&
					lastAuthor.URL == steamID.getSteamProfileLink() {
					
					lastEmbed.Description += "\n" + translatedMessage
					lastMultilineChatMessage, _ = session.ChannelMessageEditEmbed(server.ChannelID, lastMessageID, lastEmbed)
					return
				}
			}
			embed := &discordgo.MessageEmbed{
				Description: translatedMessage,
				Color: teamNumber.getColor(),
				Author: &discordgo.MessageEmbedAuthor{
					URL: steamID.getSteamProfileLink(),
					Name: username,
					IconURL: steamID.getAvatar(),
				},
			}
			lastMultilineChatMessage, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
		
		case "oneline":
			embed := &discordgo.MessageEmbed{
				Color: teamNumber.getColor(),
				Footer: &discordgo.MessageEmbedFooter{
					Text: username +": " + translatedMessage,
					IconURL: steamID.getAvatar(),
				},
			}
			_, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
		
		case "text":
			_, _ = session.ChannelMessageSend(server.ChannelID, buildTextChatMessage(server.Name, username, teamNumber, translatedMessage))
		}
	}
}


func forwardPlayerEventToDiscord(serverName string, messagetype MessageType, username string, steamID SteamID3, playerCount string) {
	if server, ok := serverList[serverName]; ok {
		
		timestamp := ""
		switch messagetype.SubType + strings.Split(playerCount, "/")[0] {
			case "join1":	fallthrough
			case "leave0":	timestamp = time.Now().UTC().Format("2006-01-02T15:04:05")
		}
		
		if playerCount != "" {
			playerCount = " (" + playerCount + ")"
		}
		
		eventText := ""
		switch messagetype.SubType {
			case "join": eventText = username + " joined" + playerCount
			case "leave": eventText = username + " left" + playerCount
		}
		
		switch Config.Discord.MessageStyle {
			default: fallthrough
			case "multiline": fallthrough
			case "oneline":
				embed := &discordgo.MessageEmbed{
					Timestamp: timestamp,
					Color: messagetype.getColor(),
					Footer: &discordgo.MessageEmbedFooter{
						Text: eventText,
						IconURL: steamID.getAvatar(),
					},
				}
				_, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
			
			case "text":
				_, _ = session.ChannelMessageSend(server.ChannelID, buildTextPlayerEvent(server.Name, messagetype, username, playerCount))
		}
	}
}


func forwardStatusMessageToDiscord(serverName string, messagetype MessageType, message string, playerCount string) {
	if server, ok := serverList[serverName]; ok {
		
		if playerCount != "" {
			message += " (" + playerCount + ")"
		}
		
		switch Config.Discord.MessageStyle {
			default: fallthrough
			case "multiline": fallthrough
			case "oneline":
				timestamp := ""
				switch messagetype.SubType {
					case "roundstart": fallthrough
					case "marinewin": fallthrough
					case "alienwin": fallthrough
					case "draw": 
						timestamp = time.Now().UTC().Format("2006-01-02T15:04:05")
				}
				embed := &discordgo.MessageEmbed{
					Timestamp: timestamp,
					Color: messagetype.getColor(),
					Footer: &discordgo.MessageEmbedFooter{
						Text: message,
						IconURL: Config.Servers[serverName].ServerIconUrl,
					},
				}
				_, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
			
			case "text":
				_, _ = session.ChannelMessageSend(server.ChannelID, Config.Servers[server.Name].ServerStatusMessagePrefix + message)
		}
	}
}
