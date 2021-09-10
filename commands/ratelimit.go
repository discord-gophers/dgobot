package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
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

func handleRatelimit(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "https://i.imgur.com/P6bDtR9.gif",
		},
	}); err != nil {
		lit.Error("error responding to ratelimit command: %v", err)
	}
}
