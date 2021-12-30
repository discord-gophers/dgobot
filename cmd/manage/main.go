package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/DiscordGophers/dgobot/commands"

	"github.com/bwmarrin/discordgo"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	session *discordgo.Session
	GuildID string
	AppID   string
)

func main() {
	fs := flag.NewFlagSet("dgobot", flag.ExitOnError)
	token := fs.String("token", "", "Discord Authentication Token")
	fs.StringVar(&GuildID, "guild-id", "", "Discord Guild ID")
	cmd := ffcli.Command{
		Name:       "manage",
		ShortUsage: `go run manage.go -token "<token>" -guild-id "<guildID> add|remove"`,
		FlagSet:    fs,
		Options:    []ff.Option{ff.WithEnvVarPrefix("DG")},
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
		Exec: func(ctx context.Context, args []string) error {
			return fmt.Errorf("error: you did not provide command add | remove")
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
	user, err := session.User("@me")
	if err != nil {
		panic(err)
	}
	AppID = user.ID

	if err := cmd.Run(context.Background()); err != nil {
		panic(err)
	}
}

func add(_ context.Context, _ []string) error {
	cmds := make([]*discordgo.ApplicationCommand, 0, len(commands.Commands))
	for _, cmd := range commands.Commands {
		cmds = append(cmds, cmd.ApplicationCommand)
	}
	_, err := session.ApplicationCommandBulkOverwrite(AppID, GuildID, cmds)
	return err
}

func remove(_ context.Context, _ []string) error {
	cmds, err := session.ApplicationCommands(AppID, GuildID)
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		if err := session.ApplicationCommandDelete(AppID, GuildID, cmd.ID); err != nil {
			return err
		}
	}
	return nil
}
