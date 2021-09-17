package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

// Version is a constant that stores the Disgord version information.
const Version = "v0.1.0-rewrite"

var (
	AdminUserID  string // Skippy
	HerderRoleID string
	Commands = make(map[string]*Command)
)

type Command struct {
	*discordgo.ApplicationCommand
	Handler func(*discordgo.Session, *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error)
}

func OnInteractionCommand(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	if ic.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := ic.ApplicationCommandData()
	cmd, ok := Commands[data.Name]
	if !ok {
		return
	}

	res, err := cmd.Handler(ds, ic)
	if err != nil {
		res = &discordgo.InteractionResponseData{
			Flags: 64,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error",
					Description: err.Error(),
					Color:       0xEE2211,
				},
			},
		}
	}

	err = ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: res,
	})
	if err != nil {
		lit.Error("responding to interaction %s: %v", data.Name, err)
	}
}

func ContentResponse(c string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content: c,
	}
}

func EphemeralResponse(c string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Flags:   64,
		Content: c,
	}
}

func EmbedResponse(e discordgo.MessageEmbed) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{&e},
	}
}
