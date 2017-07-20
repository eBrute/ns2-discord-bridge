NS2-Discord-Bridge
========================================================
NS2-Discord-Bridge is a bi-directional link between Discord chat and NS2 ingame chat.
The program creates a Discord Bot that connects to a Discord server through websockets. It also opens a HTTP server, which acts as a connection point for the game.

The game server runs a mod, which collects player messages and sends them to the HTTP server of the NS2-Discord-Bridge, which then forwards the messages to the Discord chat through the Discord bot. The reverse direction of communication is done via HTTP long polling.


## Setting up the Discord bridge

1. Get the exectuable. <br />
   Download the executable for your system architecture from here https://github.com/eBrute/ns2-discord-bridge/releases
   Alternatively you can build it yourself from source.

2. Setup a config file. <br />
   Download the example config and rename it to **config.toml**.

   In the next steps you need to setup your config. <br />
   For the bare minimum, you need to configure a Discord bot token, the port of the http server, the Steam web API key (can be empty), and one server with a Discord channel ID. We will discuss how to get each of these in the steps below.

3. Create a Discord bot and get a token. <br />
   Go to https://discordapp.com/developers/applications/me, click on My Apps and add a new App.
   Name the bot as you like (i.e. game-bot) and add a description (optional) and icon. If you can't think of one, use this:
   ![discord bot avatar](https://github.com/eBrute/ns2-discord-bridge/raw/master/images/discordbot.png) <br />
   All of these can be changed later.
   
   On the next page, create a bot user. Now you have created a bot account, and will be forwarded to the bots configuration site. There you can view the *APP Bot User Token* (not the Client Secret!). Add the token to your config file. You may make your bot public, but that is not required. Also note down the *Client ID*, which you need to add the bot to your Discord server.

4. Add the Discord bot to your Discord server. <br />
   Visit https://discordapi.com/permissions.html#523328 and enter the client id into the textfield on the lower left. Then use the invite link at the bottom to add the bot.
   You can change permissions anytime either by redoing this step or by changing the bot's role permissions in the discord server settings.
   
   After this step, you should see the bot in your Discord server's member list (as offline).
   
5. Setup HTTP server. <br />
   The discord bridge opens it's own HTTP server for communication with the game. The default config uses
   ```toml
   [httpserver]
   address = ":8080"
   ```
   which means the HTTP server runs on port 8080 on all reachable interfaces. If the discord bridge is running on the same machine as the ns2 game server, you might even use `address = "localhost:8080"`, which means the HTTP server is only accessible from the same machine.
   
   You can use any port you like. Make sure that it is not blocked by your firewall.
   
6. (Optional) Get a Steam Web API key. <br />
   The Discord bot has support for looking up a player's steam avatar. In order to do that you need a Steam Web API key. Head over to http://steamcommunity.com/dev/apikey, sign up with your Steam account. Copy the key to your config file. Alternatively you can skip this step and use and empty key in the config. In this case no Steam avatars will be shown.
   
7. Add a game server to the config. <br />
   In the config, locate the `[servers]` section.
   You now need a *server identifier*. Any alphanumeric string will do. This string is what the game server uses to mark it's messages with, so they will show up in the right channel.
   
   Let's say your server identifer is **server1**. In the config create a new subsection named `[servers.server1]`.
   The only required option is the channelID, which tells the discord bot which game server is linked to which discord channel. The the other options are optional and can be found in the example config. They are explained in detail below.
   There are multiple ways to get the ID of a discord channel:
      * Start the discordbrige bot and type `!channelinfo` in the channel you want (make sure the discord bot has access to the channel)
      * In the discord client, enable developer mode in settings -> apearance, then right click on the channel and copy the ID.
      * In the discord browser version, go to the channel and copy the second number in the url.
      
   Your config then should look like this:
   ```toml
   [servers]
         [servers.server1]
         channelID = "242940165516034049"
   ```

8. Start the discord bridge. <br />
   The bot should now come online. Type `!version` or `!help` in any channel that it has access to, it should respond.

9. Add the Discord Bridge Mod to your NS2 Server. <br />
   You can find the mod [here](http://steamcommunity.com/sharedfiles/filedetails/?id=898502581).
   The mod is a plugin for [shine](http://steamcommunity.com/sharedfiles/filedetails/?id=117887554), so you need to add it as well.
10. Configure the mod. <br />
   When you run the mod for the first time, a config file is created in **<configdir>/shine/plugins/DiscordBridge.json** where *<configdir>* is usually located in *%appdata%/Natural Selection 2* (Windows) or *~/.config/Natural Selection 2/* (Linux). Alternatively you may create the file yourself.
   ```json
    {
      "DiscordBridgeURL":"",
      "ServerIdentifier":"",
      "SendPlayerAllChat":true,
      "SendPlayerJoin":true,
      "SendPlayerLeave":true,
      "SendMapChange":true,
      "SendRoundWarmup":false,
      "SendRoundPregame":false,
      "SendRoundStart":true,
      "SendRoundEnd":true,
      "SendAdminPrint":false,
      "SpamMinIntervall":0.5,
      "__Version":"3.0.0"
    }
  ```
  Set **ServerIdentifier** to whatever alphanumeric string you chose in step 7, i.e. **server1**.
  
  In **DiscordBridgeURL** you need to supply the url under which the discord bridge is reachable, i.e. `http://example.com:8080/discordbridge` where example.com is your server or `http://localhost:8080/discordbridge` when you run the bot on the same machine.
  
  **SpamMinIntervall** is the number of seconds the game waits before sending the next message, in case that a lot of messages are spammed. A low number has impact on server performance, 0.5 should be fine.
  
  The other options determine which events should be forwarded to the discord bridge (don't enable SendAdminPrint for now).
  
11. Test it. <br />
  As soon as the mod is loaded with the new config (usually after map change), you should see the ingame chat forwarded to the discord channel, and vice versa.
  
  If it does not work type `!channelinfo` and check the last lines to see if the channel has been setup correctly.
  
  
## Discord Commands
The discord bot reacts to certain commands. All commands begin with a '!' and must be the first word in a message

Command                  | Description
------------------------ | -------------------------
!help                    | prints an overview over all commands
!commands                | prints an overview over all commands
!status                  | prints a short server status
!info                    | prints a long server info
!channelinfo             | prints ids of the current channel, guild and roles
!version                 | prints the version number of the bot
!mute @discorduser(s)    | (admin only) dont forward messages from user(s) to the server
!unmute @discorduser(s)  | (admin only) remove user(s) from being muted
!rcon <console commands> | (admin only) executes console commands directly on the linked server
   
   
## Message Style Options

**message_style** in the `[discord]` section sets the style for the discord messages. Three different output formats are supported:
  * `multiline` : <br /> ![multiline](https://github.com/eBrute/ns2-discord-bridge/raw/master/images/message_styles_multiline.png) <br />
  ➕ supports steam avatars and steam profiles, colors messages by team/type, groups consecutive messages of the same user <br />
  ➖ 2 lines per message, takes a lot of vertical space
  * `oneline` : <br /> ![oneline](https://github.com/eBrute/ns2-discord-bridge/raw/master/images/message_styles_oneline.png) <br />
  ➕ supports steam avatars, colors messages by team/type <br />
  ➖ does not group messages, player names do not link to their steam profile, wasted vertical space between messages <br />
  * `text` : <br /> ![text](https://github.com/eBrute/ns2-discord-bridge/raw/master/images/message_styles_text.png) <br />
  ➕ very dense, supports (custom) emoticons <br />
  ➖ no steam avatars or profiles, no colors, no images, no grouping <br />
    
The colors of the *multiline* and *oneline* styles are configurable in the `[messagestyles.rich]` section.
The formating of the *text* style is configurable in the `[messagestyles.text]` section. The prefixes and the message format accept emoticons in the format `"<:apheriox:298852163759898624> "` the number is the id of the custom emoticon.


## Additional Server Config Options

Certain config options require a discord identity. This is a string with either the name of a role ("my role"), the full id of a role ("164864561277698048"), the full id of a user ("125786284395462656") or the snowflake id of a user ("Brute#9034"). The latter can be aquired by just typing "@Brute" into discord and hitting enter.

Field  | Value | Description
------------- | ------------- | -------------
statusChannelID | channelID| ID of a discord channel where all status messages will be mirrored to
admins | list of discord identities | list of discord identies who have admin rights on that server. Admins can mute players and invoke remote commands on the server
keyword_notifications | list of [keyword strings], [discord identities] | List of keywords that can be used from within the game to notify certain discord identies. I.e. `[ ["cheater", "@admin" ], ["admins", "Brute#9034"], ]` will alert everyone with the "admins" role and user Brute whenever someone writes "cheater" or "@admin" in the in-game chat.
muted | list of discord identities | Discord messages of muted players are not forwarded to the game server. There is no warning (shadow ban). You can mute players on the fly with the `!mute @Brute#9034` discord command, but only the players specified in the config will survive a restart of the bot.
server_chat_message_prefix | string | Server specific prefix for all chat messages (text message style only)
server_status_message_prefix | string | Server specific prefix for all status messages (text message style only)
server_icon_url | url string | Icon that is used for status messages (in multiline and online message style). Will default to the discord channel icon when left empty


## License Information

The project makes use of [discordgo](https://github.com/bwmarrin/discordgo). The copyright lies with their respective owners.

Everything else is free to use as you see fit.