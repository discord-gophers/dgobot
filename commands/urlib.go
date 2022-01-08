package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/discord-gophers/dgobot/editor"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var client = &http.Client{Timeout: time.Second * 2}

func InitURLib(domain, pass string) {
	urlib, err := LoadURLib("urlib.json", domain, pass)
	if err != nil {
		panic(err)
	}
	Commands[cmdURL.Name] = &Command{
		ApplicationCommand: cmdURL,
		Handler:            urlib.handleURL,
		Autocomplete:       urlib.handleURLComplete,
	}
	Commands[cmdURLib.Name] = &Command{
		ApplicationCommand: cmdURLib,
		Handler:            urlib.handleURLib,
	}
}

var cmdURL = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "go",
	Description: "Show URLs associated with the keyword.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:         discordgo.ApplicationCommandOptionString,
			Name:         "keyword",
			Description:  "The keyword to show URLs for",
			Autocomplete: true,
			Required:     true,
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
			Name:        "remove",
			Description: "Remove URL",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "The URL to remove",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List URLs",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "edit",
			Description: "Edit URLs",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "apply",
			Description: "Apply updated URLs",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "code",
					Description: "The export code",
					Required:    true,
				},
			},
		},
	},
}

type UResource struct {
	URL      string    `json:"url"`
	Keywords []string  `json:"keywords"`
	Title    string    `json:"title"`
	Added    time.Time `json:"added"`
	Author   string    `json:"author"`
	ID       string    `json:"id"`
}

type URLib struct {
	mx       sync.RWMutex
	fileName string
	keyword  map[string][]*UResource
	resource map[string]*UResource

	editor.UploadApplier

	// whether the code is up to date, or if the resources have been modified
	// since
	dirty bool
	// the last known code.
	code string
	// last time the urlib edit request was made
	lastCreated time.Time
	// last time the urlib was updated via the edit request.
	lastSaved time.Time
}

func (u *URLib) Add(resource *UResource) {
	u.mx.Lock()
	defer u.mx.Unlock()

	u.resource[resource.URL] = resource
	for _, k := range resource.Keywords {
		u.keyword[k] = append(u.keyword[k], resource)
	}
}

func (u *URLib) Remove(url string) bool {
	u.mx.Lock()
	defer u.mx.Unlock()

	before := len(u.resource)
	delete(u.resource, url)
	for k, v := range u.keyword {
		for sk, sv := range v {
			if sv.URL == url {
				u.keyword[k] = append(v[:sk], v[sk+1:]...)
			}
		}
	}
	return before > len(u.resource)
}

func (u *URLib) Save() error {
	u.mx.RLock()
	defer u.mx.RUnlock()

	lit.Debug("Saving repository...")
	data, err := json.MarshalIndent(u.resource, "", "\t")
	if err != nil {
		return fmt.Errorf("marshalling urlib: %v", err)
	}

	if err := os.WriteFile(u.fileName, data, os.ModePerm); err != nil {
		return fmt.Errorf("saving %s: %v", u.fileName, err)
	}

	u.dirty = true
	return nil
}

func LoadURLib(path, domain, pass string) (*URLib, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	urlib := &URLib{
		mx:       sync.RWMutex{},
		fileName: path,
		keyword:  make(map[string][]*UResource),
		resource: make(map[string]*UResource),
		UploadApplier: editor.Filehost{
			Client: &http.Client{
				Timeout: time.Second * 5,
			},
			Domain: domain,
			Pass:   pass,
		},
		dirty: false,
	}

	if err = json.NewDecoder(f).Decode(&urlib.resource); err != nil {
		return nil, fmt.Errorf("could not unmarshal %s: %v", path, err)
	}

	for _, ur := range urlib.resource {
		for _, k := range ur.Keywords {
			urlib.keyword[k] = append(urlib.keyword[k], ur)
		}
	}

	return urlib, nil
}

func (u *URLib) handleURL(_ *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	u.mx.RLock()
	defer u.mx.RUnlock()

	arg := ic.ApplicationCommandData().Options[0].StringValue()

	// Check if we have this keyword...
	urs, ok := u.keyword[arg]
	if !ok {
		return nil, fmt.Errorf("No results for keyword `%s`.", arg)
	}

	embed := discordgo.MessageEmbed{
		Title: "URLs for " + arg,
	}
	for _, ur := range urs {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{
				Name:  fmt.Sprintf("%s - *%s*", ur.Title, ur.Author),
				Value: ur.URL,
			},
		)
	}

	return EmbedResponse(embed), nil
}

func (u *URLib) handleURLComplete(_ *discordgo.Session, ic *discordgo.InteractionCreate) ([]*discordgo.ApplicationCommandOptionChoice, error) {
	arg := ic.ApplicationCommandData().Options[0].StringValue()

	keys := make([]string, 0, len(u.keyword))
	for k := range u.keyword {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ranks := fuzzy.RankFindFold(arg, keys)

	var results []string
	for _, rank := range ranks {
		results = append(results, rank.Target)
	}
	// just show everything
	if len(results) == 0 {
		results = keys
	}

	return Autocomplete(results...), nil
}

func (u *URLib) handleURLib(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	var herder bool
	for _, role := range ic.Member.Roles {
		if role == HerderRoleID {
			herder = true
			break
		}
	}

	var f func(*discordgo.Session, *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error)
	var check bool

	cmd := ic.ApplicationCommandData().Options[0].Name
	switch cmd {
	case "add":
		f, check = u.handleURLibAdd, true
	case "remove":
		f, check = u.handleURLibRemove, true
	case "edit":
		f, check = u.handleURLibEdit, true
	case "apply":
		f, check = u.handleURLibApply, true

	case "list":
		f, check = u.handleURLibList, false
	}
	if f == nil {
		return nil, fmt.Errorf("Invalid option: `%s`.", cmd)
	}

	if !check {
		return f(ds, ic)
	}
	if !(ic.Member.User.ID == AdminUserID || herder) {
		return nil, fmt.Errorf("These commands are only for herders and above.")
	}
	return f(ds, ic)
}

func (u *URLib) handleURLibAdd(_ *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
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
		URL:      urp.String(),
		Keywords: keywords,
		Title:    title,
		Added:    time.Now(),
		Author:   ic.Member.User.Username,
		ID:       ic.Member.User.ID,
	})

	if err := u.Save(); err != nil {
		lit.Error("urlib(add): saving: %v", err)
		return nil, fmt.Errorf("Could not add: unable to save.")
	}

	return EphemeralResponse(resp), nil
}

func (u *URLib) handleURLibRemove(_ *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
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

func (u *URLib) handleURLibList(_ *discordgo.Session, _ *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	if u.dirty || u.code == "" {
		u.mx.Lock()
		defer u.mx.Unlock()

		var err error
		u.code, err = u.Upload(u.resource)
		if err != nil {
			lit.Error("could not make list: %v", err)
			return nil, fmt.Errorf("Could not create listing")
		}
		u.dirty = false
	}

	return EphemeralResponse("https://editor.discordgophers.com/" + u.code), nil
}

func (u *URLib) handleURLibEdit(_ *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	editPayload := struct {
		Type   string                `json:"type"`
		Author string                `json:"author"`
		ID     string                `json:"id"`
		URLs   map[string]*UResource `json:"urls"`
	}{"edit", ic.Member.User.Username, ic.Member.User.ID, u.resource}

	u.mx.Lock()
	defer u.mx.Unlock()

	code, err := u.Upload(editPayload)
	if err != nil {
		lit.Error("could not create edit: %v", err)
		return nil, fmt.Errorf("Could not create listing")
	}

	msg := fmt.Sprintf(
		`Private edit link: <https://editor.discordgophers.com/%s>.
Do not share this link with others.
Access code to apply changes: ||%s||.

Time since last edit request: <t:%d>
Time since last edit apply: <t:%d>`,
		code, u.ApplyCode(), u.lastCreated.Unix(), u.lastSaved.Unix())

	u.lastCreated = time.Now()

	return EphemeralResponse(msg), nil
}

func (u *URLib) handleURLibApply(_ *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	code := ic.ApplicationCommandData().Options[0].Options[0].StringValue()

	body, err := u.Apply(code)
	if err != nil {
		return nil, fmt.Errorf("could not apply code: %v", err)
	}

	u.mx.Lock()
	defer u.mx.Unlock()

	rsc := make(map[string]*UResource)
	if err = json.NewDecoder(body).Decode(&rsc); err != nil {
		return nil, fmt.Errorf("Could not apply: %v", err)
	}
	u.resource = rsc

	// we decode and then reencode to make sure no errenous fields are in the
	// payload
	data, err := json.MarshalIndent(u.resource, "", "\t")
	if err != nil {
		return nil, fmt.Errorf("Could not marshal: %v", err)
	}

	if err := os.WriteFile(u.fileName, data, os.ModePerm); err != nil {
		lit.Error("urlib(apply): saving: %v", err)
		return nil, fmt.Errorf("Could not save: %v", err)
	}

	u.keyword = make(map[string][]*UResource)
	for _, ur := range u.resource {
		for _, k := range ur.Keywords {
			u.keyword[k] = append(u.keyword[k], ur)
		}
	}

	u.lastSaved = time.Now()

	// we can reuse the new updated link
	u.dirty = false
	u.code = code

	return EphemeralResponse("Changes successfully applied."), nil
}
