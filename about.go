package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
)

func init() {
	Router.Route("about", "About this Bot.", About)
}

func About(ds *discordgo.Session, dm *discordgo.Message, ex *mux.Context) {

	resp := "\n" +
		"Hi, I'm **dgo** the official Discord Google Go library (discordgo) test bot.\n\n" +
		"I provide indispensable stress and bug testing of the discordgo library. " +
		"By allowing me to remain on your server you are directly helping to improve " +
		"both the discordgo library and Discord itself. *Thank you very very much!*\n\n" +
		"You can learn more about me at <http://dgobot.com/>\n\n" +
		"Also, checkout <https://airhorn.solutions/> and <http://septapus.com/> the two largest bots developed with the discordgo library.\n"

	if ex.IsPrivate {
		resp += "\nHint: Try command **help** to see a list of commands this bot supports."
	} else {
		resp += "\nHint: Try command **@dgo help** to see a list of commands this bot supports."
	}

	ds.ChannelMessageSend(dm.ChannelID, resp)

	return
}
