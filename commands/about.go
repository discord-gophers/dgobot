package commands

import (
	"github.com/bwmarrin/discordgo"
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
	Description: "About this bot.",
}

func handleAbout(_ *discordgo.Session, _ *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	resp := "\n" +
		"Hi, I'm **dgo** the official Discord Google Go library (discordgo) test bot.\n\n" +
		"I provide indispensable stress and bug testing of the discordgo library. " +
		"By allowing me to remain on your server you are directly helping to improve " +
		"both the discordgo library and Discord itself. *Thank you very very much!*\n\n" +
		"You can learn more about me at <https://github.com/discord-gophers/dgobot>\n\n" +
		"Also, checkout <https://airhorn.solutions/> and <http://septapus.com/> the two largest bots developed with the discordgo library.\n"

	return ContentResponse(resp), nil
}
