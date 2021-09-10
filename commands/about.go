package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

func init() {
	Commands[cmdAbout.Name] = &Command{
		ApplicationCommand: cmdAbout,
		Handler:            handleAbout,
	}
}

var cmdAbout = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "about",
	Description: "About this Bot.",
}

func handleAbout(ds *discordgo.Session, ic *discordgo.InteractionCreate) {

	resp := "\n" +
		"Hi, I'm **dgo** the official Discord Google Go library (discordgo) test bot.\n\n" +
		"I provide indispensable stress and bug testing of the discordgo library. " +
		"By allowing me to remain on your server you are directly helping to improve " +
		"both the discordgo library and Discord itself. *Thank you very very much!*\n\n" +
		"You can learn more about me at <http://dgobot.com/>\n\n" +
		"Also, checkout <https://airhorn.solutions/> and <http://septapus.com/> the two largest bots developed with the discordgo library.\n"

	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: resp,
		},
	}); err != nil {
		lit.Error("error responding to about command: %v", err)
	}
}