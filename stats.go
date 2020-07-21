package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
)

var StartTime = time.Now()

func init() {
	Router.Route("stats", "Display statistical information for this bot.", Stats)
}

func Stats(ds *discordgo.Session, dm *discordgo.Message, ctx *mux.Context) {

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	msg := "```ruby\n" +
		fmt.Sprintf("Go           : %s\n", runtime.Version()) +
		fmt.Sprintf("Disgord      : %s\n", Version) +
		fmt.Sprintf("DiscordGo    : v%s\n", discordgo.VERSION) +
		fmt.Sprintf("Uptime       : %s\n", time.Now().Sub(StartTime)) +
		fmt.Sprintf("Processes    : %d\n", runtime.NumGoroutine()) +
		fmt.Sprintf("HeapAlloc    : %.2fMB\n", float64(mem.HeapAlloc)/1048576) +
		fmt.Sprintf("Total Sys    : %.2fMB\n", float64(mem.Sys)/1048576)

	if ds.StateEnabled && ds.State != nil {
		guilds := len(ds.State.Guilds)
		channels := 0
		members := 0
		for _, v := range ds.State.Guilds {
			channels += len(v.Channels)
			members += len(v.Members)
		}

		msg += fmt.Sprintf("Guilds       : %d\n", guilds)
		msg += fmt.Sprintf("Channels     : %d\n", channels)
		msg += fmt.Sprintf("Members      : %d\n", members)
	}

	msg += "```"
	ds.ChannelMessageSend(dm.ChannelID, msg)
}
