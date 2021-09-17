//go:build manage

package main

import (
	"context"
	"flag"
	"os"

	"github.com/DiscordGophers/dgobot/commands"

	"github.com/bwmarrin/discordgo"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	session *discordgo.Session
	GuildID string
)

func main() {
	fs := flag.NewFlagSet("dgbot", flag.ExitOnError)
	token := fs.String("t", "", "Discord Authentication Token")
	fs.StringVar(&GuildID, "g", "", "Discord Guild")
	cmd := ffcli.Command{
		Name:    "dgobot manager",
		FlagSet: fs,
		Options: []ff.Option{ff.WithEnvVarPrefix("DG")},
		Subcommands: []*ffcli.Command{
			{
				Name:       "add",
				ShortUsage: "Add/upsert/overwrite commands",
				Exec:       add,
			},
			{
				Name:       "remove",
				ShortUsage: "Remove/delete commands",
				Exec:       remove,
			},
		},
	}

	err := cmd.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	session, err = discordgo.New("Bot " + *token)
	if err != nil {
		panic(err)
	}
	if err := session.Open(); err != nil {
		panic(err)
	}
	defer session.Close()

	if err := cmd.Run(context.Background()); err != nil {
		panic(err)
	}
}

func add(_ context.Context, _ []string) error {
	cmds := make([]*discordgo.ApplicationCommand, 0, len(commands.Commands))
	for _, cmd := range commands.Commands {
		cmds = append(cmds, cmd.ApplicationCommand)
	}
	_, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, GuildID, cmds)
	return err
}

func remove(_ context.Context, _ []string) error {
	cmds, err := session.ApplicationCommands(session.State.User.ID, GuildID)
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		if err := session.ApplicationCommandDelete(session.State.User.ID, GuildID, cmd.ID); err != nil {
			return err
		}
	}
	return nil
}
