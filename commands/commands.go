package commands

import (
	"github.com/bwmarrin/discordgo"
	"os"
)

// Version is a constant that stores the Disgord version information.
const Version = "v0.0.0-alpha"

var Commands map[string]*Command

type Command struct {
	*discordgo.ApplicationCommand
	Handler func(*discordgo.Session, *discordgo.InteractionCreate)
}

func GetEnvDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
