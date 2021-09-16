package commands

import (
	"github.com/bwmarrin/discordgo"
)

func init() {
	Commands[cmdSay.Name] = &Command{
		ApplicationCommand: cmdSay,
		Handler:            handleSay,
	}
}

var cmdSay = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "say",
	Description: "You will never never force me to talk...",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "stuff",
			Description: "Stuff to say",
			Required:    false,
		},
	},
}

func handleSay(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	resp := ic.ApplicationCommandData().Options[0].StringValue()
	if resp == "" {
		resp = "Say what?"
	}

	return ContentResponse(resp), nil
}
