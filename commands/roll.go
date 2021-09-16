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

// TODO: make A, X slash command params
var cmdRoll = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "roll",
	Description: "Roll the dice.",
}

func handleRoll(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	// A is number of dice to roll, X is faces of dice
	A, X := 1, 6

	var sum int
	var rn int
	for i := 0; i < A; i++ {
		rn = rand.Intn(X) + 1
		sum += rn
	}

	return ContentResponse(fmt.Sprintf("```\n[%dd%d] Rolled: %d```", A, X, sum)), nil
}
