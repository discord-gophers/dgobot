package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type oldFormat struct {
	URL      *url.URL       `json:"url"`
	Keywords []string       `json:"keywords"`
	Title    string         `json:"title"`
	Added    time.Time      `json:"added"`
	Author   discordgo.User `json:"author"`
}

// don't import to prevent calling of init
type newFormat struct {
	URL      string    `json:"url"`
	Keywords []string  `json:"keywords"`
	Title    string    `json:"title"`
	Added    time.Time `json:"added"`
	Author   string    `json:"author"`
	ID       string    `json:"id"`
}

func main() {
	fs := flag.NewFlagSet("dgobot", flag.ExitOnError)
	file := fs.String("f", "urlib.json", "file to convert to new format")
	cmd := ffcli.Command{
		Name:       "convert",
		ShortUsage: `go run ` + os.Args[0] + ` -f <file> [output]`,
		FlagSet:    fs,
		Options:    []ff.Option{ff.WithEnvVarPrefix("DG")},
		Exec: func(ctx context.Context, args []string) (err error) {
			out := os.Stdout
			if len(args) != 0 {
				out, err = os.Create(args[0])
				if err != nil {
					return fmt.Errorf("could not create %s: %w", args[0], err)
				}
			}

			in, err := os.Open(*file)
			if err != nil {
				return fmt.Errorf("could not create %s: %w", *file, err)
			}
			converted, err := convert(in)

			enc := json.NewEncoder(out)
			enc.SetIndent("", "\t")

			return enc.Encode(converted)
		},
	}

	err := cmd.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	if err := cmd.Run(context.Background()); err != nil {
		panic(err)
	}
}

func convert(file io.Reader) (map[string]newFormat, error) {
	var old map[string]oldFormat
	if err := json.NewDecoder(file).Decode(&old); err != nil {
		return nil, err
	}

	updated := make(map[string]newFormat)
	for k, v := range old {
		updated[k] = newFormat{
			URL:      v.URL.String(),
			Keywords: v.Keywords,
			Title:    v.Title,
			Added:    v.Added,
			Author:   v.Author.Username,
			ID:       v.Author.ID,
		}
	}

	return updated, nil
}
