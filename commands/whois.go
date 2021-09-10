package commands

// Command parser for the Disgord Bot package.

import (
	"fmt"
	"github.com/bwmarrin/lit"
	"log"
	"time"

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

func handleWhois(ds *discordgo.Session, ic *discordgo.InteractionCreate) {

	// Figure out who we're going to pull info on
	lookupID := ic.Member.User.ID
	optionUser := ic.ApplicationCommandData().Options[0].UserValue(ds)
	if optionUser != nil {
		lookupID = optionUser.ID
	}

	m, c, g, ok := memberChannelGuild(ds, lookupID, ic.ChannelID, ic.GuildID)
	if !ok {
		if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry " + ic.Member.Mention() + ", I couldn't find any information to give you :(",
				Flags:   64, // ephemeral
			},
		}); err != nil {
			lit.Error("error responding to whois command: %v", err)
		}
	}

	perms, err := ds.UserChannelPermissions(lookupID, ic.ChannelID)
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
		r, err := ds.State.Role(ic.GuildID, v)
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

	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				&embed,
			},
		},
	}); err != nil {
		lit.Error("error responding to whois command: %v", err)
	}
}

func memberChannelGuild(ds *discordgo.Session, memberID, channelID, guildID string) (*discordgo.Member, *discordgo.Channel, *discordgo.Guild, bool) {
	m, err := ds.State.Member(guildID, memberID)
	if err != nil {
		lit.Error("error getting Member from state: %v", err)
		return nil, nil, nil, false
	}
	fmt.Printf("Member: %#v\nUser: %#v\n", m, m.User)

	c, err := ds.State.Channel(channelID)
	if err != nil {
		lit.Error("error getting Channel from state: %v", err)
		return nil, nil, nil, false
	}

	g, err := ds.State.Guild(guildID)
	if err != nil {
		lit.Error("error getting Guild from state: %v", err)
		return nil, nil, nil, false
	}

	return m, c, g, true
}
