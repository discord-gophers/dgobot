package commands

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

// Version is a constant that stores the Disgord version information.
const Version = "v0.1.0-rewrite"

var Commands = map[string]*Command{}

type Command struct {
	Loaded bool
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

func LoadCommands(ds *discordgo.Session, all bool) error {
	commands, err := ds.ApplicationCommands(ds.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("loading commands: %v", err)
	}

	// Load active commands.
	for _, c := range commands {
		cmd, ok := Commands[c.Name]
		if !ok {
			continue
		}
		cmd.Loaded = true
		lit.Debug("Loaded %s", cmd.Name)
	}

	for name, c := range Commands {
		// Ignore already loaded when not updating all.
		if c.Loaded && !all {
			continue
		}

		_, err := ds.ApplicationCommandCreate(ds.State.User.ID, "", c.ApplicationCommand)
		if err != nil {
			return err
		}
		lit.Debug("Creating %s", name)
		c.Loaded = true
	}

	return nil
}

func GetEnvDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
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
