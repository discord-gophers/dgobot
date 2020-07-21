package main

import (
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
)

func init() {
	Router.Route("magic8", "I can answer all your [yes/no] questions!.", Magic8)
	Router.Route("8ball", "", Magic8)
}

func Magic8(ds *discordgo.Session, dm *discordgo.Message, ctx *mux.Context) {

	answers := []string{
		"It is certain",
		"It is decidedly so",
		"Without a doubt",
		"Yes definitely",
		"You may rely on it",
		"As I see it yes",
		"Most likely",
		"Outlook good",
		"Yes",
		"Signs point to yes",
		"Reply hazy try again",
		"Ask again later",
		"Better not tell you now",
		"Cannot predict now",
		"Concentrate and ask again",
		"Don't count on it",
		"My reply is no",
		"My sources say no",
		"Outlook not so good",
		"Very doubtful",
	}

	resp := answers[rand.Intn(len(answers))]
	ds.ChannelMessageSend(dm.ChannelID, resp)

	return
}
