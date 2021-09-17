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
	"github.com/peterbourgon/ff/v3"
)

func main() {
	fmt.Printf(`
	________  .__                               .___
	\______ \ |__| ______ ____   ___________  __| _/
	||    |  \|  |/  ___// ___\ /  _ \_  __ \/ __ |
	||    '   \  |\___ \/ /_/  >  <_> )  | \/ /_/ |
	||______  /__/____  >___  / \____/|__|  \____ |
	\_______\/        \/_____/   %-16s\/`+"\n\n", commands.Version)

	fs := flag.NewFlagSet("dgobot", flag.ExitOnError)
	token := fs.String("token", "", "Discord Authentication Token")
	fs.StringVar(&commands.AdminUserID, "admin-id", "109112383011581952", "Discord Admin ID")
	fs.StringVar(&commands.HerderRoleID, "herder-id", "370280974593818644", "Discord Herder Role ID")
	fs.IntVar(&lit.LogLevel, "log-level", 0, "LogLevel (0-3)")
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("DG")); err != nil {
		lit.Error("could not parse flags: %v", err)
		return
	}

	session, err := discordgo.New("Bot " + *token)
	if err != nil {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
		log.Println("You must provide a Discord authentication token.")
		return
	}

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMembers
	session.AddHandler(commands.OnInteractionCommand)

	if err := session.Open(); err != nil {
		log.Fatalf("error opening connection to Discord: %v", err)
	}
	defer session.Close()

	log.Println(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
