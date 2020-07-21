package main

// Command parser for the Disgord Bot package.

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
)

func init() {
	Router.Route("invite", "Display an invite link for this bot.", Invite)
}

func Invite(ds *discordgo.Session, dm *discordgo.Message, ex *mux.Context) {

	ds.ChannelMessageSend(dm.ChannelID, "Please visit https://discordapp.com/oauth2/authorize?client_id=173113690092994561&scope=bot to add dgo to your server.")
}
