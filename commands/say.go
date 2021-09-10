package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
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

func handleSay(ds *discordgo.Session, ic *discordgo.InteractionCreate) {

	resp := ic.ApplicationCommandData().Options[0].StringValue()
	if resp == "" {
		resp = "Say what?"
	}

	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: resp,
		},
	}); err != nil {
		lit.Error("error responding to say command: %v", err)
	}

}
