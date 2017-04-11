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
	Message string `json:"msg"`
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


func getUnicodeToTextTranslator() *strings.Replacer {
	return strings.NewReplacer(
		"😃", ":)",
		"😄", ":D",
		"😦", ":(",
		"😐", ":|",
		"😛", ":P",
		"😉", ";)",
		"😭", ";(",
		"😠", ">:(",
		"😢", ":,(",
		"❤", "<3",
		"💔", "</3",
	)
}


// formats a discord message so it looks good in-game
func formatDiscordMessage(m *discordgo.MessageCreate) string {
	message := mentionPattern.ReplaceAllStringFunc(m.Content, mentionTranslator(m.Mentions) )
	message = channelPattern.ReplaceAllStringFunc(message, channelTranslator(m.Mentions) )
	message = getUnicodeToTextTranslator().Replace(message)
	return message
}


func createChatMessageCommand(username string, m *discordgo.MessageCreate) *Command {
	return &Command{
		Type: "chat",
		User: username,
		Message: formatDiscordMessage(m),
	}
}


func createServerStatusCommand() *Command {
	return &Command{
		Type: "info",
		Message: "status",
	}
}


func createServerInfoCommand() *Command {
	return &Command{
		Type: "info",
		Message: "info",
	}
}


func createRconCommand(username string, command string) *Command {
	return &Command{
		Type: "rcon",
		User: username,
		Message: command,
	}
}
