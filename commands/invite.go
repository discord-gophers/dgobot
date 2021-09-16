package commands

import (
	"github.com/bwmarrin/discordgo"
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

const inviteLink = "https://discord.com/oauth2/authorize?client_id=173113690092994561&scope=bot"

func handleInvite(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	return ContentResponse("Please visit " + inviteLink + " to add dgo to your server."), nil
}
