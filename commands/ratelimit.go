package commands

import (
	"github.com/bwmarrin/discordgo"
)

func init() {
	Commands[cmdRatelimit.Name] = &Command{
		ApplicationCommand: cmdRatelimit,
		Handler:            handleRatelimit,
	}
}

var cmdRatelimit = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "ratelimit",
	Description: "Good heavens just look at the time!!!11!!",
}

func handleRatelimit(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	return ContentResponse("https://i.imgur.com/P6bDtR9.gif"), nil
}
