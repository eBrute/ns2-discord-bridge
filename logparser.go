// This file reads the log file of the server to parse messages that are to be shown in Discord

package main

import (
	"os"
	"bufio"
	"time"
	"regexp"
	"strconv"
	"log"
)

const fieldSep = ""

var (
	chat_regexp = regexp.MustCompile("^--DISCORD--\\|chat" +
		fieldSep + "(.*?)" + // name
		fieldSep + "(.*?)" + // steam id
		fieldSep + "(.*?)" + // team number
		fieldSep + "(.*)\n") // message

	status_regexp = regexp.MustCompile("^--DISCORD--\\|status" +
		fieldSep + "(.*?)" + // status
		fieldSep + "(.*?)" + // map
		fieldSep + "(.*)\n") // player count

	changemap_regexp = regexp.MustCompile("^--DISCORD--\\|status" +
		fieldSep + "(.*?)" + // map
		fieldSep + "(.*)\n") // player count

	init_regexp = regexp.MustCompile("^--DISCORD--\\|init" +
		fieldSep + "(.*)\n") // map

	join_regexp = regexp.MustCompile("^--DISCORD--\\|join" +
		fieldSep + "(.*?)" + // name
		fieldSep + "(.*?)" + // steam id
		fieldSep + "(.*)\n") // player count

	leave_regexp = regexp.MustCompile("^--DISCORD--\\|leave" +
		fieldSep + "(.*?)" + // name
		fieldSep + "(.*?)" + // steam id
		fieldSep + "(.*)\n") // player count

	adminprint_regexp = regexp.MustCompile("^--DISCORD--\\|adminprint" +
		fieldSep + "(.*)\n") // message
)

func startLogParser() {
	for _, server := range serverList {
		file, _ := os.Open(server.Config.LogFilePath)
		reader  := bufio.NewReader(file)
		go func() {
			for { // Skip the initial stuff; yes, this isn't the most efficient way
				line, _ := reader.ReadString('\n')
				if len(line) == 0 {
					break
				}
			}

			var slept uint = 0
			for {
				line, _ := reader.ReadString('\n')
				if len(line) != 0 {
					slept = 0
					if        matches := chat_regexp.FindStringSubmatch(line);       matches != nil {
						steamid, _    := strconv.ParseInt(matches[2], 10, 32)
						teamNumber, _ := strconv.Atoi(matches[3])
						forwardChatMessageToDiscord(server, matches[1], SteamID3(steamid), TeamNumber(teamNumber), matches[4])
					} else if matches := status_regexp.FindStringSubmatch(line);     matches != nil {
						gamestate := matches[1]
						currmap   := matches[2]
						players   := matches[3]
						var message string
						var msgtype MessageType
						msgtype.GroupType = "status"
						switch gamestate {
						case "WarmUp":
							message          = "Warm-up started on "
							msgtype.SubType = "warmup"
						case "PreGame":
							message          = "Pregame started on "
							msgtype.SubType = "pregame"
						case "Started":
							message          = "Round started on "
							msgtype.SubType = "roundstart"
						case "Team1Won":
							message          = "Marines won on "
							msgtype.SubType = "marinewin"
						case "Team2Won":
							message          = "Aliens won on "
							msgtype.SubType = "alienwin"
						case "Draw":
							message          = "Draw on "
							msgtype.SubType = "draw"
						default:
							continue
						}
						forwardStatusMessageToDiscord(server, msgtype, message, players, currmap)
					} else if matches := changemap_regexp.FindStringSubmatch(line);  matches != nil {
						nextmap := matches[1]
						players := matches[2]
						message := "Changing map to "
						forwardStatusMessageToDiscord(server, MessageType {GroupType: "status", SubType: "changemap"}, message, players, nextmap)
					} else if matches := init_regexp.FindStringSubmatch(line);       matches != nil {
						currmap := matches[1]
						message := "Loaded "
						forwardStatusMessageToDiscord(server, MessageType {GroupType: "status", SubType: "init"}, message, "", currmap)
					} else if matches := join_regexp.FindStringSubmatch(line);       matches != nil {
						name       := matches[1]
						steamid, _ := strconv.ParseInt(matches[2], 10, 32)
						players    := matches[3]
						msgtype := MessageType {
							GroupType: "player",
							SubType:   "join",
						}
						forwardPlayerEventToDiscord(server, msgtype, name, SteamID3(steamid), players)
					} else if matches := leave_regexp.FindStringSubmatch(line);      matches != nil {
						name       := matches[1]
						steamid, _ := strconv.ParseInt(matches[2], 10, 32)
						players    := matches[3]
						msgtype := MessageType {
							GroupType: "player",
							SubType:   "leave",
						}
						forwardPlayerEventToDiscord(server, msgtype, name, SteamID3(steamid), players)
					} else if matches := adminprint_regexp.FindStringSubmatch(line); matches != nil {
						forwardStatusMessageToDiscord(server, MessageType {GroupType: "adminprint"}, matches[1], "", "")
					}
				} else if slept > 10 { // Check if server has restarted
					slept = 0
					stat, _ := file.Stat()
					newfile, _ := os.Open(server.Config.LogFilePath)
					newstat, _ := newfile.Stat()
					if newstat.Size() != stat.Size() { // It is a new file
						log.Println("Server restarted!")
						file.Close()
						file   = newfile
						reader = bufio.NewReader(file)
						forwardStatusMessageToDiscord(server, MessageType {GroupType: "status", SubType: "init"}, "Server restarted!", "", "")
					} else {
						time.Sleep(500 * time.Millisecond)
						newfile.Close()
					}
				} else {
					slept += 1
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	}
}
