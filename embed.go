package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
	"github.com/bwmarrin/lit"
)

func init() {
	Router.Route("embed", "Example Embed!", embed)
}

func embed(ds *discordgo.Session, dm *discordgo.Message, ctx *mux.Context) {

	var embed discordgo.MessageEmbed

	embed.Color = 0xf2c5a8
	embed.Author = &discordgo.MessageEmbedAuthor{Name: "Embed Author", URL: "http://discordapp.com", IconURL: "https://cdn.discordapp.com/embed/avatars/0.png"}
	embed.Title = "Embed Title"
	embed.URL = "https://github.com/bwmarrin/disgord"
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: "https://cdn.discordapp.com/embed/avatars/0.png"}
	embed.Description = "This is the ~~embed~~ **description** ```go fmt.Println(`Gopher!`)```"
	embed.Image = &discordgo.MessageEmbedImage{URL: "https://cdn.discordapp.com/embed/avatars/0.png"}

	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "Field Name", Value: "Value", Inline: true},
		{Name: "Disgord", Value: Version, Inline: true},
		{Name: "DiscordGo", Value: discordgo.VERSION, Inline: true},
	}

	embed.Footer = &discordgo.MessageEmbedFooter{Text: "Footer Text", IconURL: "https://cdn.discordapp.com/embed/avatars/0.png"}
	embed.Timestamp = time.Now().UTC().Format(time.RFC3339)

	_, err := ds.ChannelMessageSendEmbed(dm.ChannelID, &embed)
	if err != nil {
		lit.Error("error sending message, %s", err)
	}
}
