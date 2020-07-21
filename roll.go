package main

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
)

func init() {
	Router.Route("roll", "Roll the dice.", Roll)
}

func Roll(ds *discordgo.Session, dm *discordgo.Message, ctx *mux.Context) {

	var A, X int // A is number of dice to roll, X is faces of dice

	A = 1
	X = 6

	var sum int
	var rn int
	for i := 0; i < A; i++ {
		rn = (rand.Intn(X) + 1)
		sum += rn
	}

	ds.ChannelMessageSend(dm.ChannelID, fmt.Sprintf("```ruby\n[%dd%d] Rolled : %d```", A, X, sum))

}
