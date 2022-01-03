package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DiscordGophers/dgobot/commands"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/peterbourgon/ff/v3"
)

func main() {
	fmt.Printf(`
	    .___            ___.           __   
	  __| _/ ____   ____\_ |__   _____/  |_ 
	 / __ | / ___\ /  _ \| __ \ /  _ \   __\
	/ /_/ |/ /_/  >  <_> ) \_\ (  <_> )  |  
	\____ |\___  / \____/|___  /\____/|__|  
	     \/_____/            \/  %s`+"\n\n", commands.Version)
	fs := flag.NewFlagSet("dgobot", flag.ExitOnError)
	token := fs.String("token", "", "Discord Authentication Token")
	domain := fs.String("domain", "https://f.teamortix.com", "Filehost domain")
	pass := fs.String("pass", "", "Filehost upload password (empty if none)")
	fs.StringVar(&commands.AdminUserID, "admin-id", "109112383011581952", "Discord Admin ID")
	fs.StringVar(&commands.HerderRoleID, "herder-id", "370280974593818644", "Discord Herder Role ID")
	fs.IntVar(&lit.LogLevel, "log-level", 0, "LogLevel (0-3)")
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("DG")); err != nil {
		lit.Error("could not parse flags: %v", err)
		return
	}

	// Seed rand for any random commands
	rand.Seed(time.Now().UnixNano())

	session, err := discordgo.New("Bot " + *token)
	if err != nil {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
		log.Println("You must provide a Discord authentication token.")
		return
	}

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMembers
	session.AddHandler(commands.OnInteractionCommand)
	session.AddHandler(commands.OnAutocomplete)
	commands.InitURLib(*domain, *pass)

	if err := session.Open(); err != nil {
		log.Fatalf("error opening connection to Discord: %v", err)
	}
	defer session.Close()

	log.Println(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
