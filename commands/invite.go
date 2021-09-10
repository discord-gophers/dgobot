package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

func init() {
	Commands[cmdInvite.Name] = &Command{
		ApplicationCommand: cmdInvite,
		Handler:            handleInvite,
	}
}

var cmdInvite = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "invite",
	Description: "Display an invite link for this bot.",
}

func handleInvite(ds *discordgo.Session, ic *discordgo.InteractionCreate) {

	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Please visit https://discordapp.com/oauth2/authorize?client_id=173113690092994561&scope=bot to add dgo to your server.",
		},
	}); err != nil {
		lit.Error("error responding to invite command: %v", err)
	}
}
