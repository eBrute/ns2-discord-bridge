package main

import (
    "strings"
    "github.com/bwmarrin/discordgo"
)

var DefaultMessageColor int = 75*256*256 + 78*256 + 82


func getColorForMessage(messagetype string) int {
    switch messagetype {
        case "chat" :        return getColorFromConfig(Config.Messagestyles.Rich.ChatMessageColor)
        case "playerjoin" :  return getColorFromConfig(Config.Messagestyles.Rich.PlayerJoinColor)
        case "playerleave" : return getColorFromConfig(Config.Messagestyles.Rich.PlayerLeaveColor)
        case "status" :      return getColorFromConfig(Config.Messagestyles.Rich.StatusColor)
        case "adminprint" :  return getColorFromConfig(Config.Messagestyles.Rich.StatusColor)
        default :            return DefaultMessageColor
    }
}


func getTeamColorForChatMessage(teamNumber int) int {
    switch teamNumber {
        default: fallthrough
        case 0 : return getColorFromConfig(Config.Messagestyles.Rich.ChatMessageReadyRoomColor)
        case 1 : return getColorFromConfig(Config.Messagestyles.Rich.ChatMessageMarineColor)
        case 2 : return getColorFromConfig(Config.Messagestyles.Rich.ChatMessageAlienColor)
        case 3 : return getColorFromConfig(Config.Messagestyles.Rich.ChatMessageSpectatorColor)
    }
}


func getTeamSpecifixPrefix(teamNumber int) string {
    messageConfig := Config.Messagestyles.Text
    switch teamNumber {
        case 0: return messageConfig.ChatMessageReadyRoomPrefix
        case 1: return messageConfig.ChatMessageMarinePrefix
        case 2: return messageConfig.ChatMessageAlienPrefix
        case 3: return messageConfig.ChatMessageSpectatorPrefix
        default: return ""
    }
}


func getColorFromConfig(color []int) int {
    if len(color) != 3 {
        return DefaultMessageColor
    }
    return color[0]*256*256 + color[1]*256 + color[2]
}


func buildTextChatMessage(serverName string, username string, teamNumber int, message string) string {
    serverConfig := Config.Servers[serverName]
    messageFormat := Config.Messagestyles.Text.ChatMessageFormat
    teamSpecificString := getTeamSpecifixPrefix(teamNumber)
    serverSpecificString := serverConfig.ServerChatMessagePrefix
    replacer := strings.NewReplacer("%p", username, "%m", message, "%t", teamSpecificString, "%s", serverSpecificString)
	formattedMessage := replacer.Replace(messageFormat)
    return formattedMessage
}


func buildTextPlayerEvent(serverName, cmdType, username, message string) string {
    serverConfig := Config.Servers[serverName]
    messageConfig := Config.Messagestyles.Text
    messageFormat := "%s %p %m"
    switch cmdType {
        case "playerjoin": messageFormat = messageConfig.PlayerJoinFormat
        case "playerleave": messageFormat = messageConfig.PlayerJoinFormat
    }
    serverSpecificString := serverConfig.ServerChatMessagePrefix
    replacer := strings.NewReplacer("%p", username, "%m", message, "%s", serverSpecificString)
	formattedMessage := replacer.Replace(messageFormat)
    return formattedMessage
}


func forwardChatMessageToDiscord(serverName string, username string, steamID SteamID3, teamNumber int, message string) {
	if server, ok := Servers[serverName]; ok {
		
		switch Config.Discord.MessageStyle {
		default: fallthrough
		case "multiline":
			embed := &discordgo.MessageEmbed{
				Description: message,
				Color: getTeamColorForChatMessage(teamNumber),
				Author: &discordgo.MessageEmbedAuthor{
					URL: getSteamProfileLinkForSteamID(steamID),
					Name: username,
					IconURL: getAvatarForSteamID(steamID),
				},
			}
			 _, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
		
		case "inline": fallthrough
		case "oneline":
			embed := &discordgo.MessageEmbed{
				Color: getTeamColorForChatMessage(teamNumber),
				Footer: &discordgo.MessageEmbedFooter{
					Text: username +": " + message,
					IconURL: getAvatarForSteamID(steamID),
				},
			}
			 _, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
		
		case "text":
			_, _ = session.ChannelMessageSend(server.ChannelID, buildTextChatMessage(server.Name, username, teamNumber, message))
		}
	}
}


func forwardPlayerEventToDiscord(serverName string, cmdType string, username string, steamID SteamID3, message string) {
	if server, ok := Servers[serverName]; ok {
		eventText := ""
		switch cmdType {
    		case "playerjoin": eventText = " joined "
		          case "playerleave": eventText = " left "
		}
		
		switch Config.Discord.MessageStyle {
    		default: fallthrough
    		case "multiline": fallthrough
    		case "inline": fallthrough
    		case "oneline":
    			embed := &discordgo.MessageEmbed{
    				Color: getColorForMessage(cmdType),
    				Footer: &discordgo.MessageEmbedFooter{
    					Text: username + eventText + message,
    					IconURL: getAvatarForSteamID(steamID),
    				},
    			}
    			 _, _ = session.ChannelMessageSendEmbed(server.ChannelID, embed)
    		
    		case "text":
    			_, _ = session.ChannelMessageSend(server.ChannelID, buildTextPlayerEvent(server.Name, cmdType, username, message))
		}
	}
}


func forwardGameStatusToDiscord(serverName string, cmdType string, message string) {
	if server, ok := Servers[serverName]; ok {
		
		switch Config.Discord.MessageStyle {
    		default: fallthrough
    		case "multiline": fallthrough
    		case "inline": fallthrough
    		case "oneline":
    			embed := &discordgo.MessageEmbed{
    				Color: getColorForMessage(cmdType),
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
