package commands

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Commands[cmdRoll.Name] = &Command{
		ApplicationCommand: cmdRoll,
		Handler:            handleRoll,
	}
}

var cmdRoll = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "roll",
	Description: "Roll the dice.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "num",
			Description: "Number of dice (default 1)",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "1", Value: 1},
				{Name: "2", Value: 2},
				{Name: "3", Value: 3},
				{Name: "4", Value: 4},
				{Name: "5", Value: 5},
				{Name: "6", Value: 6},
				{Name: "7", Value: 7},
				{Name: "8", Value: 8},
				{Name: "9", Value: 9},
				{Name: "10", Value: 10},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "faces",
			Description: "Number of dice faces (default 6)",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "d4", Value: 4},
				{Name: "d6", Value: 6},
				{Name: "d8", Value: 8},
				{Name: "d10", Value: 10},
				{Name: "d12", Value: 12},
				{Name: "d20", Value: 20},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "modifier",
			Description: "Modifier to add to final result (default 0)",
			Required:    false,
		},
	},
}

func handleRoll(_ *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	num, faces := 1, 6
	var modifier int64

	for _, opt := range ic.ApplicationCommandData().Options {
		switch opt.Name {
		case "num":
			num = int(opt.IntValue())
		case "faces":
			faces = int(opt.IntValue())
		case "modifier":
			modifier = opt.IntValue()
		}
	}

	sum := modifier
	for i := 0; i < num; i++ {
		sum += int64(rand.Intn(faces) + 1)
	}

	return ContentResponse(fmt.Sprintf("```\n[%dd%d%+d] Rolled: %d```", num, faces, modifier, sum)), nil
}
