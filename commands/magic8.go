package commands

import (
	"fmt"
	"math/rand/v2"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Commands[cmd8Ball.Name] = &Command{
		ApplicationCommand: cmd8Ball,
		Handler:            handle8Ball,
	}
}

var cmd8Ball = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "8ball",
	Description: "I can answer all your [yes/no] questions!",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "question",
			Description: "The life-changing question to ask",
			Required:    true,
		},
	},
}

func handle8Ball(_ *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	query := ic.ApplicationCommandData().Options[0].StringValue()
	resp := fmt.Sprintf("> %s\n%s", query, magicAnswers[rand.IntN(len(magicAnswers))])
	return ContentResponse(resp), nil
}

var magicAnswers = []string{
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
