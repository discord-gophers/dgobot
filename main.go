package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/discord-gophers/dgobot/commands"

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
	fs.StringVar(&commands.JobsChannelID, "jobs-channel", "484358165236809748", "Job listings channel ID")
	fs.StringVar(&commands.JobsRoleID, "jobs-role", "1337402297306579006", "Job posting role ID")
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
	session.AddHandler(commands.OnAutocomplete)
	session.AddHandler(commands.OnInteractionOther)
	commands.InitURLib(*domain, *pass)

	if err := session.Open(); err != nil {
		log.Fatalf("error opening connection to Discord: %v", err)
	}
	defer session.Close()

	user, err := session.User("@me")
	if err != nil {
		panic(err)
	}

	cmds := make([]*discordgo.ApplicationCommand, 0, len(commands.Commands))
	for _, v := range commands.Commands {
		cmds = append(cmds, v.ApplicationCommand)
	}
	if _, err := session.ApplicationCommandBulkOverwrite(user.ID, "", cmds); err != nil {
		panic(err)
	}

	log.Printf(`Now running as %s. Press CTRL-C to exit.`, user.ID)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
