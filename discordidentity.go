// This file contains functions to identify a discord identity.
// A discord identity might be a user, referenced by id or by snowflake id (username + "#" + discriminator),
// or a role, referenced by id or name

package ns2discordbridge

import (
	"errors"
	"github.com/bwmarrin/discordgo"
)

type DiscordIdentity string
type DiscordIdentityList []DiscordIdentity


func (d *DiscordIdentity) UnmarshalText(data []byte) error {
	*d = DiscordIdentity(data)
    return nil
}


func (d DiscordIdentity) MarshalText() ([]byte, error) {
    return []byte(d), nil
}


func (d *DiscordIdentity) String() string {
    return string(*d)
}


func (list DiscordIdentityList) isInList(member *discordgo.Member) bool {
	for _, entry := range list {
		if entry.matches(member) {
			return true
		}
	}
	return false
}


func (list DiscordIdentityList) toMentionString(guild *discordgo.Guild) (response string) {
	for _, mention := range list {
		if role, err := mention.getRole(guild); err == nil {
			response += "<@&" + role.ID + "> "
		} else if user, err := mention.getUser(guild); err == nil {
			response += "<@!" + user.ID + "> "
		}
	}
	return response
}


func (identity *DiscordIdentity) matches(member *discordgo.Member) bool {
	if identity.matchesUser(member.User) {
		return true
	}
	
	for _, roleID := range member.Roles {
		guildID := member.GuildID
		role, err := session.State.Role(guildID, roleID)
		if err == nil && identity.matchesRole(role) {
			return true
		}
	}
	return false
}


func (identity *DiscordIdentity) matchesUser(user *discordgo.User) bool {
	userName := user.Username + "#" + user.Discriminator
	if string(*identity) == user.ID || string(*identity) == userName {
		return true
	}
	return false
}


func (identity *DiscordIdentity) matchesRole(role *discordgo.Role) bool {
	if string(*identity) == role.ID || string(*identity) == role.Name {
		return true
	}
	return false
}


func (identity *DiscordIdentity) getUser(guild *discordgo.Guild) (*discordgo.User, error) {
	for _, member:= range guild.Members {
		if identity.matchesUser(member.User) {
			return member.User, nil
		}
	}
	return nil, errors.New("No User '" + string(*identity) + "' not found")
}


func (identity *DiscordIdentity) getRole(guild *discordgo.Guild) (*discordgo.Role, error) {
	for _, role  := range guild.Roles {
		if identity.matchesRole(role) {
			return role, nil
		}
	}
	return nil, errors.New("Role '" + string(*identity) + "' not found")
}
