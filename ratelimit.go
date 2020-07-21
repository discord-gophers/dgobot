package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
)

func init() {
	Router.Route("ratelimit", "Good heavens just look at the time!!!11!!", RateLimit)
}

func RateLimit(ds *discordgo.Session, dm *discordgo.Message, ctx *mux.Context) {

	ds.ChannelMessageSend(dm.ChannelID, "http://i.imgur.com/P6bDtR9.gif")
}
