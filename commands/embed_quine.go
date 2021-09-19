package commands

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var embed = discordgo.MessageEmbed{
	Title:       "Embed Title",
	Description: "This is the ~~embed~~ **description**\n```go\ngo fmt.Println(`Gopher!`)\n```",
	Color:       0xf2c5a8,
	URL:         "https://github.com/DiscordGophers/dgobot",
	Timestamp:   time.Now().UTC().Format(time.RFC3339),
	Author: &discordgo.MessageEmbedAuthor{
		Name:    "Embed Author",
		URL:     "https://discordapp.com",
		IconURL: "https://cdn.discordapp.com/embed/avatars/0.png",
	},
	Thumbnail: &discordgo.MessageEmbedThumbnail{
		URL: "https://cdn.discordapp.com/embed/avatars/0.png",
	},
	Image: &discordgo.MessageEmbedImage{
		URL: "https://cdn.discordapp.com/embed/avatars/0.png",
	},
	Fields: []*discordgo.MessageEmbedField{
		{Name: "Field Name", Value: "Value", Inline: true},
		{Name: "dgobot", Value: Version, Inline: true},
		{Name: "DiscordGo", Value: discordgo.VERSION, Inline: true},
	},
	Footer: &discordgo.MessageEmbedFooter{
		Text:    "Footer Text",
		IconURL: "https://cdn.discordapp.com/embed/avatars/0.png",
	},
}
