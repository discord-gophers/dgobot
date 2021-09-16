package commands

// Command parser for the Disgord Bot package.

import (
	"fmt"
	"time"

	"github.com/bwmarrin/lit"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Commands[cmdWhois.Name] = &Command{
		ApplicationCommand: cmdWhois,
		Handler:            handleWhois,
	}
}

var cmdWhois = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "whois",
	Description: "If you do not know, I will tell you.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "User to get information on - default self",
		},
	},
}

func handleWhois(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	// Figure out who we're going to pull info on
	lookupID := ic.Member.User.ID
	optionUser := ic.ApplicationCommandData().Options[0].UserValue(ds)
	if optionUser != nil {
		lookupID = optionUser.ID
	}

	m, c, g, err := memberChannelGuild(ds, lookupID, ic.ChannelID, ic.GuildID)
	if err != nil {
		lit.Error("whois: %v", err)
		return nil, fmt.Errorf("Sorry, I couldn't find any information to give you :(")
	}

	perms, err := ds.UserChannelPermissions(lookupID, ic.ChannelID)
	if err != nil {
		lit.Error("whois: channel perms: %v", err)
		return nil, fmt.Errorf("An error occurred.")
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
		r, err := ds.State.Role(ic.GuildID, v)
		if err != nil {
			lit.Error("whois: fetching role %s: %v\n", v, err)
			continue
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
	js := "N/A"
	if string(m.JoinedAt) != "" {
		jd, err := time.Parse(time.RFC3339, string(m.JoinedAt))
		if err != nil {
			lit.Error("whois: parsing date: %v", err)
		} else {
			js = fmt.Sprintf("<t:%d:R>", jd.Unix())
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

	return EmbedResponse(embed), nil
}

func memberChannelGuild(ds *discordgo.Session, memberID, channelID, guildID string) (*discordgo.Member, *discordgo.Channel, *discordgo.Guild, error) {
	m, err := ds.State.Member(guildID, memberID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not get member from state: %v", err)
	}

	c, err := ds.State.Channel(channelID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not get channel from state: %v", err)
	}

	g, err := ds.State.Guild(guildID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not get guild from state: %v", err)
	}

	return m, c, g, nil
}
