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

const steamBaseline = 76561197960265728

type ISteamUser struct {
    Response SteamPlayerList
}

type SteamPlayerList struct {
    Players []SteamPlayer
}

type SteamPlayer struct {
    Steamid string
    Communityvisibilitystate int
	Profilestate int
	Personaname string
	Lastlogoff int64
    Avatar string
    Avatarmedium string
    Avatarfull string
    Personastate int
	Primaryclanid string
	Timecreated int64
	Personastateflags int
	Loccountrycode string
	Locstatecode string
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


func (id SteamID3) to64() SteamID64 {
	return SteamID64(steamBaseline + uint64(id))
}


func (id SteamID64) String() string {
	return strconv.FormatUint(uint64(id), 10)
}


func getAvatarForSteamID(steamID SteamID3) string {
    if avatar, ok := AvatarCache[steamID]; ok {
        if time.Now().Before(avatar.lastUpdated.Add(time.Duration(24) * time.Hour)) {
            return avatar.url
        }
    }
    steamProfile, err := getSteamProfile(steamID.to64())
    if err == nil {
        AvatarCache[steamID] = &Avatar{
            url : steamProfile.Avatar,
            lastUpdated : time.Now(),
        }
        return steamProfile.Avatar
    }
    return ""
}


func getSteamProfileLinkForSteamID(steamID SteamID3) string {
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


func getSteamProfile(steamID SteamID64) (*SteamPlayer, error) {
    if steamID == 0 {
        return nil, errors.New("Invalid Steamid")
    }
    
    steamResponse := ISteamUser{}
    url := "http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=" + Config.Steam.WebApiKey + "&steamids=" + steamID.String()
    if err := getJson(url, &steamResponse); err != nil {
        return nil, err
    }
    if len(steamResponse.Response.Players) == 0 {
        return nil, errors.New("Empty response")
    }
    steamProfile := &steamResponse.Response.Players[0]
    return steamProfile, nil
}