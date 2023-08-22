package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

const (
	// Version is a constant that stores the dgobot version information.
	Version       = "v0.1.0-rewrite"
	ephemeralFlag = 64
)

var (
	AdminUserID  string // Skippy
	HerderRoleID string
	Commands     = make(map[string]*Command)
)

type Command struct {
	*discordgo.ApplicationCommand
	Handler      func(*discordgo.Session, *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error)
	Autocomplete func(*discordgo.Session, *discordgo.InteractionCreate) ([]*discordgo.ApplicationCommandOptionChoice, error)
}

func OnAutocomplete(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	if ic.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return
	}

	data := ic.ApplicationCommandData()
	cmd, ok := Commands[data.Name]
	if !ok || cmd.Autocomplete == nil {
		return
	}

	choices, err := cmd.Autocomplete(ds, ic)
	if err != nil {
		return
	}

	err = ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		lit.Error("responding to autocomplete: %v", err)
	}
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
			Flags: ephemeralFlag,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error",
					Description: err.Error(),
					Color:       0xEE2211,
				},
			},
		}
	}

	typ := discordgo.InteractionResponseChannelMessageWithSource
	if res.Title != "" {
		typ = discordgo.InteractionResponseModal
	}
	err = ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: typ,
		Data: res,
	})
	if err != nil {
		lit.Error("responding to interaction %s: %v", data.Name, err)
	}
}

// OnModalSubmit routes modal submit interactions to the appropriate handler.
// it uses `prefix:` from the custom ID to determine which handler to use.
func OnModalSubmit(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	if ic.Type != discordgo.InteractionModalSubmit {
		return
	}

	data := ic.ModalSubmitData()
	prefix, _, ok := strings.Cut(data.CustomID, ":")
	if !ok {
		lit.Error("Invalid custom ID: %s", data.CustomID)
		EphemeralResponse("Invalid modal submit.")
		return
	}

	cmd := Commands[prefix]
	res, err := cmd.Handler(ds, ic)
	if err != nil {
		res = &discordgo.InteractionResponseData{
			Flags: ephemeralFlag,
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
		lit.Error("responding to modal submit %s: %v", data.CustomID, err)
	}
}

func ContentResponse(c string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content: c,
	}
}

func EphemeralResponse(c string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Flags:   ephemeralFlag,
		Content: c,
	}
}

func EmbedResponse(e discordgo.MessageEmbed) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{&e},
	}
}

func FileResponse(f discordgo.File) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Files: []*discordgo.File{&f},
	}
}

func Autocomplete(options ...string) []*discordgo.ApplicationCommandOptionChoice {
	if len(options) > 25 {
		options = options[:25]
	}

	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, opt := range options {
		if len(opt) > 100 {
			opt = opt[:100]
		}

		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  opt,
			Value: opt,
		})
	}
	return choices
}
