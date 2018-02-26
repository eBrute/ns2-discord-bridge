// This file reads the log file of the server to parse messages that are to be shown in Discord

package main

import (
	"os"
	"bufio"
	"time"
	"regexp"
	"strconv"
	"log"
	"path/filepath"
	"strings"
)

const fieldSep = ""
const regex_prefix = "^\\[[0-9][0-9]:[0-9][0-9]:[0-9][0-9]\\]--DISCORD--\\|"

var (
	chat_regexp = regexp.MustCompile(regex_prefix + "chat" +
		fieldSep + "(.*?)" + // name
		fieldSep + "(.*?)" + // steam id
		fieldSep + "(.*?)" + // team number
		fieldSep + "(.*)\n") // message

	status_regexp = regexp.MustCompile(regex_prefix + "status" +
		fieldSep + "(.*?)" + // status
		fieldSep + "(.*?)" + // map
		fieldSep + "(.*)\n") // player count

	changemap_regexp = regexp.MustCompile(regex_prefix + "changemap" +
		fieldSep + "(.*?)" + // map
		fieldSep + "(.*)\n") // player count

	init_regexp = regexp.MustCompile(regex_prefix + "init" +
		fieldSep + "(.*)\n") // map

	player_regexp = regexp.MustCompile(regex_prefix + "player" +
		fieldSep + "(.*?)" + // action
		fieldSep + "(.*?)" + // name
		fieldSep + "(.*?)" + // steam id
		fieldSep + "(.*)\n") // player count

	adminprint_regexp = regexp.MustCompile(regex_prefix + "adminprint" +
		fieldSep + "(.*)\n") // message
)

func findLogFile(logpath string) string {
	dir := filepath.Dir(logpath)
	prefix := dir + string(os.PathSeparator) + "log-Server"
	var file string
	var mod_time time.Time
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		switch {
		case path == dir:
		case info.Mode().IsDir():
			return filepath.SkipDir
		case strings.HasPrefix(path, prefix) && info.ModTime().After(mod_time):
			mod_time = info.ModTime()
			file = path
		}
		return nil
	})
	return file
}

func startLogParser() {
	for _, server := range serverList {
		logfile := server.Config.LogFilePath
		currlog := findLogFile(logfile)
		file, _ := os.Open(currlog)
		reader  := bufio.NewReader(file)
		go func() {
			for { // Skip the initial stuff; yes, this isn't the most efficient way
				line, _ := reader.ReadString('\n')
				if len(line) == 0 {
					break
				}
			}

			var slept    uint  = 0
			var filesize int64 = 0
			for {
				line, err := reader.ReadString('\n')
				if err != nil && len(line) != 0 {
					slept    = 0
					filesize = 0
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
						/* These are pretty much useless
						case "WarmUp":
							message          = "Warm-up started on "
							msgtype.SubType = "warmup"
						case "PreGame":
							message          = "Pregame started on "
							msgtype.SubType = "pregame"
						*/
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
					} else if matches := player_regexp.FindStringSubmatch(line);     matches != nil {
						action     := matches[1]
						name       := matches[2]
						steamid, _ := strconv.ParseInt(matches[3], 10, 32)
						players    := matches[4]
						msgtype := MessageType {
							GroupType: "player",
							SubType:   action,
						}
						forwardPlayerEventToDiscord(server, msgtype, name, SteamID3(steamid), players)
					} else if matches := adminprint_regexp.FindStringSubmatch(line); matches != nil {
						forwardStatusMessageToDiscord(server, MessageType {GroupType: "adminprint"}, matches[1], "", "")
					}
				} else if slept >= 5 { // Check if server has restarted
					slept = 0

					newlog := findLogFile(logfile)
					if newlog == currlog {
						newfile, err := os.Open(currlog)
						if err != nil {
							log.Println(err)
							continue
						}
						newstat, err := newfile.Stat()
						if err != nil {
							log.Println(err)
							continue
						}

						if filesize == 0 {
							filesize = newstat.Size()
						} else if newstat.Size() - filesize > 200 || filesize - newstat.Size() > 0 { // It is a new file
							log.Printf("Server restarted! (size changed, %v != %v)\n", filesize, newstat.Size())
							filesize = 0
							file.Close()
							file   = newfile
							reader = bufio.NewReader(file)
							forwardStatusMessageToDiscord(server, MessageType {GroupType: "status", SubType: "init"}, "Server restarted!", "", "")
						} else {
							time.Sleep(500 * time.Millisecond)
							newfile.Close()
						}
					} else {
						currlog = newlog
						newfile, err := os.Open(currlog)
						if err != nil {
							continue
						}
						filesize = 0
						log.Printf("Server restarted! (log file changed, %v != %v)", currlog, newlog)
						file.Close()
						file   = newfile
						reader = bufio.NewReader(file)
						forwardStatusMessageToDiscord(server, MessageType {GroupType: "status", SubType: "init"}, "Server restarted!", "", "")
					}
				} else {
					slept += 1
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	}
}
