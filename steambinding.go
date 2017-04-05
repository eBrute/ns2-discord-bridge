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

var myClient = &http.Client{Timeout: 10 * time.Second}

func GetAvatarForSteamID3(steamID3 int32) string {
    steamID := getSteamID(steamID3)
    steamProfile, err := GetSteamProfile(steamID)
    if err == nil {
        return steamProfile.Avatar
    }
    return ""
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


func GetSteamProfile(steamID int64) (*SteamPlayer, error) {
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