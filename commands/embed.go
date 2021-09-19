package commands

import (
	_ "embed"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Commands[cmdEmbed.Name] = &Command{
		ApplicationCommand: cmdEmbed,
		Handler:            handleEmbed,
	}
}

var cmdEmbed = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "embed",
	Description: "Example Embed!",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "show-code",
			Description: "Show the code used for the embed",
			Required:    false,
		},
	},
}

//go:embed embed_quine.go
var embedQuine string

func handleEmbed(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	var showCode bool
	if len(ic.ApplicationCommandData().Options) > 0 {
		showCode = ic.ApplicationCommandData().Options[0].BoolValue()
	}

	if showCode {
		return FileResponse(discordgo.File{Name: "embed.go", ContentType: "text/plain", Reader: strings.NewReader(embedQuine)}), nil
	}

	return EmbedResponse(embed), nil
}
