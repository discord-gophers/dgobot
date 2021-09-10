package commands

// Command parser for the Disgord Bot package.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

func init() {
	urlib, err := LoadURLib("urlib.json")
	if err != nil {
		panic(err)
	}
	Commands[cmdURL.Name] = &Command{
		ApplicationCommand: cmdURL,
		Handler:            urlib.handleURL,
	}
	Commands[cmdURLib.Name] = &Command{
		ApplicationCommand: cmdURLib,
		Handler:            urlib.handleURLib,
	}
}

var cmdURL = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "go",
	Description: "URLs",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "keyword",
			Description: "The keyword to show URLs for",
			Required:    true,
		},
	},
}

var cmdURLib = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "urlib",
	Description: "URL changes",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add URL",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "The URL to add",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "keyword",
					Description: "The keywords, delimited by comma",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "The title",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List URLs",
		},
	},
}

type UResource struct {
	URL      *url.URL
	Keywords []string
	Title    string
	Added    time.Time
	Author   discordgo.User
}

// Skippy
var adminUserID = GetEnvDefault("DGOBOT_ADMIN_ID", "109112383011581952")
var herderRoleID = GetEnvDefault("DGOBOT_HERDER_ID", "370280974593818644")

type URLib struct {
	mx       sync.Mutex
	fileName string
	keyword  map[string][]*UResource
	resource map[string]*UResource
}

func (u *URLib) Add(resource *UResource) {
	u.resource[resource.URL.String()] = resource
	for _, k := range resource.Keywords {
		kws, ok := u.keyword[k]
		if !ok {
			kws = []*UResource{}
		}
		kws = append(kws, resource)
		u.keyword[k] = kws
	}
}

func (u *URLib) Remove(url string) bool {
	before := len(u.resource)
	delete(u.resource, url)
	for k, v := range u.keyword {
		for sk, sv := range v {
			if sv.URL.String() == url {
				u.keyword[k] = append(v[:sk], v[sk+1:]...)
			}
		}
	}
	return before > len(u.resource)
}

func (u *URLib) Save() error {
	u.mx.Lock()
	defer u.mx.Unlock()
	lit.Debug("Saving repository...")
	data, err := json.MarshalIndent(u.resource, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal urllib: %v", err)
	}

	if err := os.WriteFile(u.fileName, data, os.ModePerm); err != nil {
		return fmt.Errorf("error saving %s. %v", u.fileName, err)
	}

	return nil
}

func LoadURLib(fileName string) (*URLib, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", fileName, err)
	}

	var urlib *URLib
	err = json.Unmarshal(data, &urlib)
	if err != nil {
		lit.Error("error unmarshalling urlib:", err)
	}

	for _, ur := range urlib.resource {
		for _, k := range ur.Keywords {
			kws, ok := urlib.keyword[k]
			if !ok {
				kws = []*UResource{}
			}
			kws = append(kws, ur)
			urlib.keyword[k] = kws
		}
	}

	return urlib, nil
}

func (u *URLib) handleURL(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	arg := ic.ApplicationCommandData().Options[0].StringValue()
	// Check if we have this keyword...
	urs, ok := u.keyword[strings.TrimSpace(arg)]
	if !ok {
		return
	}

	var msg string
	for _, ur := range urs {
		msg += fmt.Sprintf("**%s**, <%s> - *%s*\n", ur.Title, ur.URL.String(), ur.Author.Username)

	}
	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	}); err != nil {
		lit.Error("error responding to url command: %v", err)
	}
}

func (u *URLib) handleURLib(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	var herder bool
	for _, role := range ic.Member.Roles {
		if role == herderRoleID {
			herder = true
			break
		}
	}
	if !(ic.Member.User.ID == adminUserID || herder) {
		if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "These commands are only for herders and above.",
			},
		}); err != nil {
			lit.Error("error responding to urlib command: %v", err)
		}
	}

	cmd := ic.ApplicationCommandData().Options[0].StringValue()
	switch cmd {
	case "add":
		u.handleURLibAdd(ds, ic)
	case "remove":
		u.handleURLibRemove(ds, ic)
	case "list":
		u.handleURLibList(ds, ic)
	}
}

func (u *URLib) handleURLibAdd(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	ur := ic.ApplicationCommandData().Options[0].Options[0].StringValue()
	urp, err := url.Parse(ur)
	if err != nil {
		lit.Error("%v", err)
		return
	}
	keywordStr := ic.ApplicationCommandData().Options[0].Options[1].StringValue()
	keywords := strings.Split(keywordStr, ",")
	title := ic.ApplicationCommandData().Options[0].Options[2].StringValue()

	resp := fmt.Sprintf("Added %s", ur)
	u.Add(&UResource{
		URL:      urp,
		Keywords: keywords,
		Title:    title,
		Added:    time.Now(),
		Author:   *ic.Member.User,
	})

	if err := u.Save(); err != nil {
		lit.Error("%v", err)
	}

	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: resp,
			Flags:   64, // ephemeral
		},
	}); err != nil {
		lit.Error("error responding to urlib command: %v", err)
	}
}

func (u *URLib) handleURLibRemove(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	arg := ic.ApplicationCommandData().Options[0].Options[0].StringValue()

	resp := fmt.Sprintf("Removed %s", arg)
	if ok := u.Remove(arg); !ok {
		resp = fmt.Sprintf("Could not remove %s", arg)
	}

	if err := u.Save(); err != nil {
		lit.Error("%v", err)
	}

	if err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: resp,
			Flags:   64, // ephemeral
		},
	}); err != nil {
		lit.Error("error responding to urlib command: %v", err)
	}
}

func (u *URLib) handleURLibList(ds *discordgo.Session, ic *discordgo.InteractionCreate) {

}
