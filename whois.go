package main

// Command parser for the Disgord Bot package.

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
)

func init() {
	Router.Route("whois", "If you do not know, I will tell you.", Whois)
	Router.Route("who", "", Whois)
}

// Whois tells us information about a Discord User/Bot
func Whois(ds *discordgo.Session, dm *discordgo.Message, ctx *mux.Context) {

	// Figure out who we're going to pull info on
	lookupID := dm.Author.ID
	for _, v := range dm.Mentions {
		if v.ID == ds.State.User.ID {
			continue
		}
		lookupID = v.ID
	}

	m, err := ds.State.Member(dm.GuildID, lookupID)
	if err != nil {
		log.Println("error getting Member from state,", err)
		ds.ChannelMessageSend(dm.ChannelID, "Sorry, "+dm.Author.Mention()+" I couldn't find any information to give you :(")
		return
	}
	fmt.Printf("Member: %#v\nUser: %#v\n", m, m.User)

	c, err := ds.State.Channel(dm.ChannelID)
	if err != nil {
		log.Println("error getting Channel from state,", err)
		ds.ChannelMessageSend(dm.ChannelID, "Sorry, "+dm.Author.Mention()+" I couldn't find any information to give you :(")
		return
	}

	g, err := ds.State.Guild(dm.GuildID)
	if err != nil {
		log.Println("error getting Guild from state,", err)
		ds.ChannelMessageSend(dm.ChannelID, "Sorry, "+dm.Author.Mention()+" I couldn't find any information to give you :(")
		return
	}

	perms, err := ds.UserChannelPermissions(lookupID, dm.ChannelID)
	if err != nil {
		fmt.Println(err)
	}
	//if (perms & discordgo.PermissionManageServer) > 0 {
	// This user has Manage Server permission
	//}
	permbits := fmt.Sprintf("%032b", perms)

	var embed discordgo.MessageEmbed
	embed.Type = "rich"
	embed.Color = 0xf2c5a8
	embed.Author = &discordgo.MessageEmbedAuthor{Name: "whois " + m.User.ID, URL: "http://dgobot.com", IconURL: ""}
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: ds.State.User.AvatarURL("")}
	var roles string
	var delim string
	for _, v := range m.Roles {
		r, err := ds.State.Role(dm.GuildID, v)
		if err != nil {
			fmt.Printf("err fetching role %s, %s\n", v, err)
		}
		roles = roles + delim + r.Name + " (" + r.ID + ")"
		delim = ", "
	}
	if roles == "" {
		roles = "none"
	}
	// Little sanity checks, discord doesn't like it when we send an empty
	// Field value
	if m.Nick == "" {
		m.Nick = m.User.Username
	}
	var js = "N/A"
	if string(m.JoinedAt) != "" {
		jd, err := time.Parse(time.RFC3339, string(m.JoinedAt))
		if err != nil {
			log.Println("error parsing date,", err)
		} else {
			tsjd := time.Since(jd)

			js = fmt.Sprintf(
				"%dd, %dh, %dm, %ds ago.",
				int(tsjd.Hours()/24),
				int(tsjd.Hours())%24,
				int(tsjd.Minutes())%60,
				int(tsjd.Seconds())%60,
			)
		}
	}

	embedFields := []*discordgo.MessageEmbedField{
		{Name: "Joined", Value: js, Inline: true},
		{Name: "Bot Account?", Value: fmt.Sprintf("%t", m.User.Bot), Inline: true},
		{Name: "Username", Value: m.User.Username + "#" + m.User.Discriminator, Inline: true},
		{Name: "Nickname", Value: m.Nick, Inline: true},
		{Name: g.Name + " Server Roles", Value: roles, Inline: true},
		{Name: c.Name + " Chan Perms", Value: permbits, Inline: false},
	}
	embed.Fields = embedFields
	embed.Image = &discordgo.MessageEmbedImage{URL: m.User.AvatarURL("")}
	embed.Footer = &discordgo.MessageEmbedFooter{Text: "Provided with gopher love", IconURL: "https://cdn.discordapp.com/embed/avatars/0.png"}
	embed.Timestamp = time.Now().UTC().Format(time.RFC3339)

	_, err = ds.ChannelMessageSendEmbed(dm.ChannelID, &embed)
	if err != nil {
		log.Println(err)
	}
}
