package commands

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

func init() {
	Commands[cmdEmbed.Name] = &Command{
		ApplicationCommand: cmdEmbed,
		Handler:            handleEmbed,
	}
}

var cmdEmbed = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "embed",
	Description: "Example Embed!",
}

func handleEmbed(ds *discordgo.Session, ic *discordgo.InteractionCreate) {

	var embed discordgo.MessageEmbed

	embed.Color = 0xf2c5a8
	embed.Author = &discordgo.MessageEmbedAuthor{Name: "Embed Author", URL: "https://discordapp.com", IconURL: "https://cdn.discordapp.com/embed/avatars/0.png"}
	embed.Title = "Embed Title"
	embed.URL = "https://github.com/bwmarrin/disgord"
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: "https://cdn.discordapp.com/embed/avatars/0.png"}
	embed.Description = "This is the ~~embed~~ **description**\n```go\ngo fmt.Println(`Gopher!`)\n```"
	embed.Image = &discordgo.MessageEmbedImage{URL: "https://cdn.discordapp.com/embed/avatars/0.png"}

	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "Field Name", Value: "Value", Inline: true},
		{Name: "Disgord", Value: Version, Inline: true},
		{Name: "DiscordGo", Value: discordgo.VERSION, Inline: true},
	}

	embed.Footer = &discordgo.MessageEmbedFooter{Text: "Footer Text", IconURL: "https://cdn.discordapp.com/embed/avatars/0.png"}
	embed.Timestamp = time.Now().UTC().Format(time.RFC3339)

	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				&embed,
			},
		},
	}); err != nil {
		lit.Error("error responding to embed command: %v", err)
	}
}
