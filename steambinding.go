package main

import (
	"net/http"
	"time"
    "strconv"
    "errors"
	"encoding/json"
)


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
var AvatarCache map[int32]*Avatar


func initSteamBinding() {
    AvatarCache = make(map[int32]*Avatar)
}

func getAvatarForSteamID3(steamID3 int32) string {
    if avatar, ok := AvatarCache[steamID3]; ok {
        if time.Now().Before(avatar.lastUpdated.Add(time.Duration(24) * time.Hour)) {
            return avatar.url
        }
    }
    steamID := getSteamID(steamID3)
    steamProfile, err := getSteamProfile(steamID)
    if err == nil {
        AvatarCache[steamID3] = &Avatar{
            url : steamProfile.Avatar,
            lastUpdated : time.Now(),
        }
        return steamProfile.Avatar
    }
    return ""
}


func getSteamProfileLinkForSteamID3(steamID3 int32) string {
    if steamID3 == 0 {
        return ""
    }
	steamID := getSteamID(steamID3)
	return "http://steamcommunity.com/profiles/" + strconv.FormatInt(steamID, 10)
}


func getSteamID(steamID3 int32) int64 {
    var steamBaseline int64 = 76561197960265728
    return steamBaseline + int64(steamID3)
}


func getJson(url string, target interface{}) error {
    r, err := myClient.Get(url)
    if err != nil {
        return err
    }
    defer r.Body.Close()

    return json.NewDecoder(r.Body).Decode(target)
}


func getSteamProfile(steamID int64) (*SteamPlayer, error) {
    if steamID == 0 {
        return nil, errors.New("Invalid Steamid")
    }
    
    steamResponse := ISteamUser{}
    url := "http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=" + Config.Steam.WebApiKey + "&steamids=" + strconv.FormatInt(steamID, 10)
    if err := getJson(url, &steamResponse); err != nil {
        return nil, err
    }
    if len(steamResponse.Response.Players) == 0 {
        return nil, errors.New("Empty response")
    }
    steamProfile := &steamResponse.Response.Players[0]
    return steamProfile, nil
}