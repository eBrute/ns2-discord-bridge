package main

import (
    "strings"
	"regexp"
    "github.com/bwmarrin/discordgo"
)

var (
	mentionPattern *regexp.Regexp
	channelPattern *regexp.Regexp
)

type Command struct {
    Type    string `json:"type"`
    User    string `json:"user"`
    Content string `json:"content"`
}


func init() {
	mentionPattern, _ = regexp.Compile(`[\\]?<@[!]?\d+>`)
	channelPattern, _ = regexp.Compile(`[\\]?<#\d+>`)
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
	}
}


// formats a discord message so it looks good in-game
func formatDiscordMessage(m *discordgo.MessageCreate) string {
	message := mentionPattern.ReplaceAllStringFunc(m.Content, mentionTranslator(m.Mentions) )
	message = channelPattern.ReplaceAllStringFunc(message, channelTranslator(m.Mentions) )
	return message
}


func createChatMessageCommand(username string, m *discordgo.MessageCreate) *Command {
	return &Command{
		Type: "chat",
		User: username,
		Content: formatDiscordMessage(m),
	}
}


func createRconCommand(username string, command string) *Command {
	return &Command{
		Type: "rcon",
		User: username,
		Content: command,
	}
}
