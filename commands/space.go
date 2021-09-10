package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

func init() {
	Commands[cmdSpace.Name] = &Command{
		ApplicationCommand: cmdSpace,
		Handler:            handleSpace,
	}
}

var cmdSpace = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "space",
	Description: "Take b1nzy to space!",
}

func handleSpace(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "https://takeb1nzyto.space",
		},
	}); err != nil {
		lit.Error("error responding to space command: %v", err)
	}
}
