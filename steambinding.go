package main

import (
	"net/http"
	"time"
	"strconv"
	"errors"
	"encoding/json"
)

type SteamID3  uint32
type SteamID64 uint64

const steamBaseline uint64 = 76561197960265728

type ISteamUser struct {
	Response SteamPlayerList     `json:"response"`
}

type SteamPlayerList struct {
	Players []SteamPlayer        `json:"players"`
}

type SteamPlayer struct {
	SteamID string               `json:"steamid"`
	CommunityVisibilityState int `json:"communityvisibilitystate"`
	ProfileState int             `json:"profilestate"`
	PersonaName string           `json:"personaname"`
	LastLogoff int64             `json:"lastlogoff"`
	ProfileURL string            `json:"profileurl"`
	Avatar string                `json:"avatar"`
	AvatarMedium string          `json:"avatarmedium"`
	AvatarFull string            `json:"avatarfull"`
	PersonaState int             `json:"personastate"`
	PrimaryClanID string         `json:"primaryclanid"`
	TimeCreated int64            `json:"timecreated"`
	PersonaStateFlags int        `json:"personastateflags"`
	LocCountryCode string        `json:"loccountrycode"`
	LocStateCode string          `json:"locstatecode"`
}

type Avatar struct {
	url string
	lastUpdated time.Time
}

var myClient = &http.Client{Timeout: 10 * time.Second}
var AvatarCache map[SteamID3]*Avatar


func init() {
	AvatarCache = make(map[SteamID3]*Avatar)
}


func (steamID SteamID3) to64() SteamID64 {
	return SteamID64(steamBaseline + uint64(steamID))
}


func (steamID SteamID64) String() string {
	return strconv.FormatUint(uint64(steamID), 10)
}


func (steamID SteamID3) getAvatar() string {
	if avatar, ok := AvatarCache[steamID]; ok {
		if time.Now().Before( avatar.lastUpdated.Add( time.Duration(24) * time.Hour)) {
			return avatar.url
		}
	}
	steamProfile, err := steamID.getSteamProfile()
	if err == nil {
		AvatarCache[steamID] = &Avatar{
			url : steamProfile.Avatar,
			lastUpdated : time.Now(),
		}
		return steamProfile.Avatar
	}
	return ""
}


func (steamID SteamID3) getSteamProfileLink() string {
	if steamID == 0 {
		return ""
	}
	return "http://steamcommunity.com/profiles/" + steamID.to64().String()
}



func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}


func (steamID SteamID3) getSteamProfile() (*SteamPlayer, error) {
	if steamID == 0 {
		return nil, errors.New("Invalid Steamid")
	}
	
	steamResponse := ISteamUser{}
	url := "http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=" + Config.Steam.WebApiKey + "&steamids=" + steamID.to64().String()
	if err := getJson(url, &steamResponse); err != nil {
		return nil, err
	}
	if len(steamResponse.Response.Players) == 0 {
		return nil, errors.New("Empty response")
	}
	steamProfile := &steamResponse.Response.Players[0]
	return steamProfile, nil
}