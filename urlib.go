package main

// Command parser for the Disgord Bot package.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

type UResource struct {
	URL         *url.URL
	Keywords    []string
	Title       string
	Description string
	Added       time.Time
	Author      discordgo.User
}

// Skippy
const adminUserID = `109112383011581952`
const herderRoleID = `370280974593818644`

var URLib map[string]*UResource
var URKeywordMap map[string][]*UResource
var URMutex sync.Mutex

func SaveURLib() {

	URMutex.Lock()
	defer URMutex.Unlock()

	lit.Debug("Saving repository...")

	data, err := json.MarshalIndent(URLib, "", "\t")
	if err != nil {
		lit.Error("Unable to marshal urllib: %v", err)
	}

	if err := ioutil.WriteFile("urlib.json", data, os.ModePerm); err != nil {
		lit.Error("Error saving urlib.json. %v", err)
	}

}

func LoadURLib() {

	URMutex.Lock()
	defer URMutex.Unlock()

	data, err := ioutil.ReadFile("urlib.json")
	if err != nil {
		lit.Error("error reading urlib.json:", err)
		return
	}

	err = json.Unmarshal(data, &URLib)
	if err != nil {
		lit.Error("error unmarshalling urlib:", err)
	}

	// populate keyword map
	for _, ur := range URLib {
		for _, k := range ur.Keywords {
			kws, ok := URKeywordMap[k]
			if !ok {
				URKeywordMap[k] = []*UResource{ur}
				continue
			}

			kws = append(kws, ur)
			URKeywordMap[k] = kws
		}
	}

}

func init() {

	Session.AddHandler(urlib)

	URLib = make(map[string]*UResource)

	URKeywordMap = make(map[string][]*UResource)

	LoadURLib()
}

func urlib(dg *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore self, always.
	if m.Author.ID == dg.State.User.ID {
		return
	}

	// Check for magic prefix
	if !strings.HasPrefix(m.Content, `?go`) {
		return
	}

	arg := strings.TrimPrefix(m.Content, `?go`)
	arg = strings.TrimSpace(arg)

	// check if we're adding content
	// format :
	// ?category +url|keyword,keyword,keyword|title or summary
	// check for Gopher Herder role
	herder := false
	gm, err := dg.State.Member(m.GuildID, m.Author.ID)
	if errors.Is(err, discordgo.ErrStateNotFound) {
		lit.Debug("error fetching user %s from state: %v", m.Author.ID, err)
		gm, err = dg.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			lit.Warn("error fetching user %s: %v", m.Author.ID, err)
		}
	}

	if err == nil {
		for _, v := range gm.Roles {
			if v == herderRoleID {
				herder = true
			}
		}
	}

	if strings.HasPrefix(arg, `?`) && (m.Author.ID == adminUserID || herder) {
		msg := `
?go ? : help (?go ?)
?go + : add (?go +URL|KEYWORD KEYWORD KEYWORD|TITLE)
?go - : del (?go -URL)
?go / : search (?go /keyword)
default is search, spaces are unimportant.
`
		_, err = dg.ChannelMessageSend(m.ChannelID, msg)
		if err != nil {
			lit.Warn("Error sending discord message: %v", err)
		}

	}

	if strings.HasPrefix(arg, `=`) && (m.Author.ID == adminUserID || herder) {
		for k, v := range URLib {
			lit.Debug("[%s] = %#v\n", k, v)
		}
	}

	if strings.HasPrefix(arg, `-`) && (m.Author.ID == adminUserID || herder) {

		arg = strings.TrimPrefix(arg, `-`)
		lit.Debug("[%s]", arg)
		arg = strings.TrimSpace(arg)
		lit.Debug("[%s]", arg)
		delete(URLib, arg)
		SaveURLib()

		// clean up the keyword map
		for k, v := range URKeywordMap {
			for sk, sv := range v {
				if sv.URL.String() == arg {
					URKeywordMap[k] = append(v[:sk], v[sk+1:]...)
				}
			}
		}

		return
	}

	if strings.HasPrefix(arg, `+`) && (m.Author.ID == adminUserID || herder) {

		arg = strings.TrimPrefix(arg, `+`)

		fields := strings.SplitN(arg, `|`, 3)
		fmt.Printf("fields : %#v", fields)
		if len(fields) < 3 {
			fmt.Println("fields < 3")
			return
		}

		fmt.Printf("Adding: %#v", fields)

		ur := &UResource{}
		var err error
		ur.URL, err = url.Parse(strings.TrimSpace(fields[0]))
		if err != nil {
			fmt.Println(err)
			return
		}

		ur.Keywords = strings.Split(strings.TrimSpace(fields[1]), ` `)
		ur.Title = strings.TrimSpace(fields[2])
		ur.Author = *m.Author

		// save to urlib
		URLib[ur.URL.String()] = ur

		fmt.Printf("\nUR: %#v", ur)

		// save keywords
		for _, k := range ur.Keywords {
			kws, ok := URKeywordMap[k]
			if !ok {
				URKeywordMap[k] = []*UResource{ur}
				continue
			}

			kws = append(kws, ur)
			URKeywordMap[k] = kws
		}

		// Save to disk
		SaveURLib()

		return
	}

	if strings.HasPrefix(arg, `/`) {
		arg = strings.TrimPrefix(arg, `/`)
	}

	arg = strings.TrimSpace(arg)

	// Check if we have this keyword..
	urs, ok := URKeywordMap[strings.TrimSpace(arg)]
	if !ok {
		return
	}

	var msg string = "\n"

	for _, ur := range urs {
		msg += fmt.Sprintf("**%s**, <%s> - *%s*\n", ur.Title, ur.URL.String(), ur.Author.Username)

	}

	dg.ChannelMessageSend(m.ChannelID, msg)

}
