package main

import (
    "strings"
    "github.com/bwmarrin/discordgo"
)

type TeamNumber int
type MessageType string

var DefaultMessageColor int = 75*256*256 + 78*256 + 82


func (messagetype MessageType) getColor() int {
    switch messagetype {
        case "chat" :        return Config.getColor(Config.Messagestyles.Rich.ChatMessageColor, DefaultMessageColor)
        case "playerjoin" :  return Config.getColor(Config.Messagestyles.Rich.PlayerJoinColor, DefaultMessageColor)
        case "playerleave" : return Config.getColor(Config.Messagestyles.Rich.PlayerLeaveColor, DefaultMessageColor)
        case "status" :      return Config.getColor(Config.Messagestyles.Rich.StatusColor, DefaultMessageColor)
        case "adminprint" :  return Config.getColor(Config.Messagestyles.Rich.StatusColor, DefaultMessageColor)
        default :            return DefaultMessageColor
    }
}


func (teamNumber TeamNumber) getColor() int {
    switch teamNumber {
        default: fallthrough
        case 0 : return Config.getColor(Config.Messagestyles.Rich.ChatMessageReadyRoomColor, DefaultMessageColor)
        case 1 : return Config.getColor(Config.Messagestyles.Rich.ChatMessageMarineColor, DefaultMessageColor)
        case 2 : return Config.getColor(Config.Messagestyles.Rich.ChatMessageAlienColor, DefaultMessageColor)
        case 3 : return Config.getColor(Config.Messagestyles.Rich.ChatMessageSpectatorColor, DefaultMessageColor)
    }
}


func (teamNumber TeamNumber) getText() string {
    messageConfig := Config.Messagestyles.Text
    switch teamNumber {
        case 0: return messageConfig.ChatMessageReadyRoomPrefix
        case 1: return messageConfig.ChatMessageMarinePrefix
        case 2: return messageConfig.ChatMessageAlienPrefix
        case 3: return messageConfig.ChatMessageSpectatorPrefix
        default: return ""
    }
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
    switch messagetype {
        case "playerjoin": messageFormat = messageConfig.PlayerJoinFormat
        case "playerleave": messageFormat = messageConfig.PlayerLeaveFormat
    }
    serverSpecificString := serverConfig.ServerChatMessagePrefix
    replacer := strings.NewReplacer("%p", username, "%m", message, "%s", serverSpecificString)
	formattedMessage := replacer.Replace(messageFormat)
    return formattedMessage
}


func forwardChatMessageToDiscord(serverName string, username string, steamID SteamID3, teamNumber TeamNumber, message string) {
	if server, ok := serverList[serverName]; ok {
		
		switch Config.Discord.MessageStyle {
		default: fallthrough
		case "multiline":
			embed := &discordgo.MessageEmbed{
				Description: message,
				Color: teamNumber.getColor(),
				Author: &discordgo.MessageEmbedAuthor{
					URL: steamID.getSteamProfileLink(),
					Name: username,
					IconURL: steamID.getAvatar(),
				},
			}
			 _, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
		
		case "inline": fallthrough
		case "oneline":
			embed := &discordgo.MessageEmbed{
				Color: teamNumber.getColor(),
				Footer: &discordgo.MessageEmbedFooter{
					Text: username +": " + message,
					IconURL: steamID.getAvatar(),
				},
			}
			 _, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
		
		case "text":
			_, _ = session.ChannelMessageSend(server.ChannelID, buildTextChatMessage(server.Name, username, teamNumber, message))
		}
	}
}


func forwardPlayerEventToDiscord(serverName string, messagetype MessageType, username string, steamID SteamID3, message string) {
	if server, ok := serverList[serverName]; ok {
		eventText := ""
		switch messagetype {
            case "playerjoin": eventText = " joined "
            case "playerleave": eventText = " left "
		}
		
		switch Config.Discord.MessageStyle {
    		default: fallthrough
    		case "multiline": fallthrough
    		case "inline": fallthrough
    		case "oneline":
    			embed := &discordgo.MessageEmbed{
    				Color: messagetype.getColor(),
    				Footer: &discordgo.MessageEmbedFooter{
    					Text: username + eventText + message,
    					IconURL: steamID.getAvatar(),
    				},
    			}
    			 _, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
    		
    		case "text":
    			_, _ = session.ChannelMessageSend(server.ChannelID, buildTextPlayerEvent(server.Name, messagetype, username, message))
		}
	}
}


func forwardGameStatusToDiscord(serverName string, messagetype MessageType, message string) {
	if server, ok := serverList[serverName]; ok {
		
		switch Config.Discord.MessageStyle {
    		default: fallthrough
    		case "multiline": fallthrough
    		case "inline": fallthrough
    		case "oneline":
    			embed := &discordgo.MessageEmbed{
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
