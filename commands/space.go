package commands

import (
	"github.com/bwmarrin/discordgo"
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

func handleSpace(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	return ContentResponse("https://takeb1nzyto.space"), nil
}
