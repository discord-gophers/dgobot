package commands

import (
	"fmt"
	
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

func handleInvite(ds *discordgo.Session, _ *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	inviteLink := fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s&scope=bot%%20application.commands", ds.State.User.ID)
	return ContentResponse("Please visit " + inviteLink + " to add dgo to your server."), nil
}
