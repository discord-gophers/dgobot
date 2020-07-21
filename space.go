package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
)

func init() {

	Router.Route("space", "Take b1nzy to space!", Space)
}

func Space(ds *discordgo.Session, dm *discordgo.Message, ctx *mux.Context) {

	ds.ChannelMessageSend(dm.ChannelID, "http://takeb1nzyto.space")
}
