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
var (
	adminUserID  = GetEnvDefault("DG_ADMIN_ID", "109112383011581952")
	herderRoleID = GetEnvDefault("DG_HERDER_ID", "370280974593818644")
)

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
		return fmt.Errorf("marshalling urllib: %v", err)
	}

	if err := os.WriteFile(u.fileName, data, os.ModePerm); err != nil {
		return fmt.Errorf("saving %s: %v", u.fileName, err)
	}

	return nil
}

func LoadURLib(path string) (*URLib, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var urlib *URLib
	if err = json.NewDecoder(f).Decode(&urlib); err != nil {
		return nil, fmt.Errorf("could not unmarshal %s: %v", path, err)
	}

	for _, ur := range urlib.resource {
		for _, k := range ur.Keywords {
			// nil slice is okay
			kws := urlib.keyword[k]

			kws = append(kws, ur)
			urlib.keyword[k] = kws
		}
	}

	return urlib, nil
}

func (u *URLib) handleURL(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	arg := ic.ApplicationCommandData().Options[0].StringValue()

	// Check if we have this keyword...
	urs, ok := u.keyword[arg]
	if !ok {
		return nil, fmt.Errorf("No results for keyword `%s`.", arg)
	}

	var msg string
	for _, ur := range urs {
		msg += fmt.Sprintf("**%s**, <%s> - *%s*\n", ur.Title, ur.URL.String(), ur.Author.Username)
	}
	return ContentResponse(msg), nil
}

func (u *URLib) handleURLib(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	var herder bool
	for _, role := range ic.Member.Roles {
		if role == herderRoleID {
			herder = true
			break
		}
	}

	if !(ic.Member.User.ID == adminUserID || herder) {
		return nil, fmt.Errorf("These commands are only for herders and above.")
	}

	cmd := ic.ApplicationCommandData().Options[0].StringValue()
	switch cmd {
	case "add":
		return u.handleURLibAdd(ds, ic)
	case "remove":
		return u.handleURLibRemove(ds, ic)
	case "list":
		return u.handleURLibList(ds, ic)
	}
	return nil, fmt.Errorf("Invalid option: `%s`.", cmd)
}

func (u *URLib) handleURLibAdd(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	ur := ic.ApplicationCommandData().Options[0].Options[0].StringValue()
	urp, err := url.Parse(ur)
	if err != nil {
		lit.Error("urlib(add): parsing URL: %v", err)
		return nil, fmt.Errorf("Could not add: invalid URL provided.")
	}

	keywordStr := ic.ApplicationCommandData().Options[0].Options[1].StringValue()
	keywords := strings.Split(keywordStr, ",")
	title := ic.ApplicationCommandData().Options[0].Options[2].StringValue()

	resp := fmt.Sprintf("Added `%s`.", ur)
	u.Add(&UResource{
		URL:      urp,
		Keywords: keywords,
		Title:    title,
		Added:    time.Now(),
		Author:   *ic.Member.User,
	})

	if err := u.Save(); err != nil {
		lit.Error("urlib(add): saving: %v", err)
		return nil, fmt.Errorf("Could not add: unable to save.")
	}

	return EphemeralResponse(resp), nil
}

func (u *URLib) handleURLibRemove(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	arg := ic.ApplicationCommandData().Options[0].Options[0].StringValue()

	if ok := u.Remove(arg); !ok {
		return nil, fmt.Errorf("Could not remove `%s`: no results found.", arg)
	}

	if err := u.Save(); err != nil {
		lit.Error("urlib(rm): saving url: %v", err)
		return nil, fmt.Errorf("Could not remove: unable to save.")
	}

	resp := fmt.Sprintf("Removed `%s`", arg)
	return EphemeralResponse(resp), nil
}

func (u *URLib) handleURLibList(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	return nil, fmt.Errorf("unimplemented")
}
