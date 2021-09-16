package commands

import (
	"fmt"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
)

var StartTime = time.Now()

func init() {
	Commands[cmdStats.Name] = &Command{
		ApplicationCommand: cmdStats,
		Handler:            handleStats,
	}
}

var cmdStats = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "stats",
	Description: "Display statistical information for this bot.",
}

func handleStats(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	var embed discordgo.MessageEmbed
	embed.Color = 0xf2c5a8
	embed.Title = "Stats"
	embed.URL = "https://github.com/DiscordGophers/dgobot"
	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "Go", Value: runtime.Version(), Inline: true},
		{Name: "dgobot", Value: Version, Inline: true},
		{Name: "DiscordGo", Value: discordgo.VERSION, Inline: true},
		{Name: "Uptime", Value: fmt.Sprintf("<t:%d:R>", time.Now().Unix()), Inline: true},
		{Name: "Processes", Value: fmt.Sprint(runtime.NumGoroutine()), Inline: true},
		{Name: "HeapAlloc", Value: fmt.Sprintf("%.2fMB", float64(mem.HeapAlloc)/1048576), Inline: true},
		{Name: "Total Sys", Value: fmt.Sprintf("%.2fMB", float64(mem.Sys)/1048576), Inline: true},
	}

	if ds.StateEnabled && ds.State != nil {
		guilds := len(ds.State.Guilds)
		channels := 0
		members := 0
		for _, v := range ds.State.Guilds {
			channels += len(v.Channels)
			members += len(v.Members)
		}
		embed.Fields = append(embed.Fields, []*discordgo.MessageEmbedField{
			{Name: "Guilds", Value: fmt.Sprint(guilds), Inline: true},
			{Name: "Channels", Value: fmt.Sprint(channels), Inline: true},
			{Name: "Members", Value: fmt.Sprint(members), Inline: true},
		}...)
	}
	embed.Timestamp = time.Now().UTC().Format(time.RFC3339)

	return EmbedResponse(embed), nil
}
