package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DiscordGophers/dgobot/commands"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

// Session is declared in the global space so it can be easily used
// throughout this program.
// In this use case, there is no error that would be returned.
var Session, _ = discordgo.New()

// Read in all configuration options from both environment variables and
// command line arguments.
func init() {
	Session.Token = os.Getenv("DG_TOKEN")
	if Session.Token == "" {
		flag.StringVar(&Session.Token, "t", "", "Discord Authentication Token")
	}
	flag.IntVar(&lit.LogLevel, "l", 0, "LogLevel  (0-3)")
}

func main() {
	fmt.Printf(`
	________  .__                               .___
	\______ \ |__| ______ ____   ___________  __| _/
	||    |  \|  |/  ___// ___\ /  _ \_  __ \/ __ |
	||    '   \  |\___ \/ /_/  >  <_> )  | \/ /_/ |
	||______  /__/____  >___  / \____/|__|  \____ |
	\_______\/        \/_____/   %-16s\/`+"\n\n", commands.Version)

	flag.Parse()
	if Session.Token == "" {
		flag.Usage()
		log.Println("You must provide a Discord authentication token.")
		return
	}

	Session.AddHandler(commands.OnInteractionCommand)

	err := Session.Open()
	if err != nil {
		log.Fatalf("error opening connection to Discord: %v", err)
	}

	commands.LoadCommands(Session, false)

	log.Println(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	Session.Close()
}
