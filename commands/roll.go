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
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "faces",
			Description: "Number of dice faces (default 6)",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "bonus",
			Description: "Bonus modifier (default 0)",
			Required:    false,
		},
	},
}

func handleRoll(_ *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	num, faces, bonus := 1, 6, 0

	for _, opt := range ic.ApplicationCommandData().Options {
		switch opt.Name {
		case "num":
			num = int(opt.IntValue())
		case "faces":
			faces = int(opt.IntValue())
		case "bonus":
			bonus = int(opt.IntValue())
		}
	}

	sum := bonus
	var rn int
	for i := 0; i < num; i++ {
		rn = rand.Intn(faces) + 1
		sum += rn
	}

	return ContentResponse(fmt.Sprintf("```\n[%dd%d+%d] Rolled: %d```", num, faces, bonus, sum)), nil
}
