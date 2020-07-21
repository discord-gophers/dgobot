package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
)

func init() {
	Router.Route("say", "You will never never force me to talk...", Say)
}

func Say(ds *discordgo.Session, dm *discordgo.Message, ctx *mux.Context) {

	say := strings.SplitN(dm.Content, "say ", 2)

	if len(say) < 2 {
		ds.ChannelMessageSend(dm.ChannelID, "Say what?")
		return
	}

	ds.ChannelMessageSend(dm.ChannelID, say[1])

}
