package commands

import (
	"fmt"
	"github.com/bwmarrin/lit"
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
}

func handleRoll(ds *discordgo.Session, ic *discordgo.InteractionCreate) {

	var A, X int // A is number of dice to roll, X is faces of dice

	A = 1
	X = 6

	var sum int
	var rn int
	for i := 0; i < A; i++ {
		rn = rand.Intn(X) + 1
		sum += rn
	}
	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("```ruby\n[%dd%d] Rolled : %d```", A, X, sum),
		},
	}); err != nil {
		lit.Error("error responding to roll command: %v", err)
	}

}
