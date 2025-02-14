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
	// AdminUserID is the user id for the server admin.
	AdminUserID string // Skippy
	// HerderRoleID is the "moderator" role equivalent for the server.
	HerderRoleID string
	// JobsChannelID is used for the channel where new job listings are submitted
	// for review. It is *not* for the channel where job listings are shown
	// publicly.
	JobsChannelID string
	// JobsRoleID is used to provide access to post job listings. It is given when
	// approved.
	JobsRoleID string
	Commands   = make(map[string]*Command)
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
	if !ok || cmd.Handler == nil {
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

// OnInteractionOther routes modal submit/message component interactions to the appropriate handler.
// it uses `prefix:` from the custom ID to determine which handler to use.
func OnInteractionOther(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	var customID string
	switch ic.Type {
	case discordgo.InteractionModalSubmit:
		customID = ic.ModalSubmitData().CustomID
	case discordgo.InteractionMessageComponent:
		customID = ic.MessageComponentData().CustomID
	default:
		return
	}

	prefix, _, ok := strings.Cut(customID, ":")
	if !ok {
		lit.Error("Invalid custom ID: %s", customID)
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

	typ := discordgo.InteractionResponseChannelMessageWithSource
	if res.Title != "" {
		typ = discordgo.InteractionResponseModal
	}
	// DIRTY HORRIBLE UGLY HACK: we use TTS to flag to update the source message
	if res.TTS {
		typ = discordgo.InteractionResponseUpdateMessage
		res.TTS = false
	}
	err = ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: typ,
		Data: res,
	})
	if err != nil {
		lit.Error("responding to modal submit %s: %v", customID, err)
	}
}

func ContentResponse(c string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content: c,
	}
}

func UpdateMessageResponse(msg *discordgo.Message) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content:    msg.Content,
		Components: msg.Components,
		Embeds:     msg.Embeds,
		TTS:        true,
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
