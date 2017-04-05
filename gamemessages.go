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


func initGameMessages() {
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
		return match
	}
}


func formatDiscordMessage(m *discordgo.MessageCreate) string {
	message := mentionPattern.ReplaceAllStringFunc(m.Content, mentionTranslator(m.Mentions) )
	message = channelPattern.ReplaceAllStringFunc(message, channelTranslator(m.Mentions) )
	return message
}