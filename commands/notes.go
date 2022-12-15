package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

func init() {
	notes, err := LoadNotes("notes.json")
	if err != nil {
		panic(err)
	}
	Commands[cmdNotes.Name] = &Command{
		ApplicationCommand: cmdNotes,
		Handler:            notes.handleNotesRaw,
	}
	Commands[appNoteUser.Name] = &Command{
		ApplicationCommand: appNoteUser,
		Handler:            notes.handleNotesApp,
	}
}

var appNoteUser = &discordgo.ApplicationCommand{
	Type: discordgo.UserApplicationCommand,
	Name: "See Notes",
}

var cmdNotes = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "notes",
	Description: "Information tracking",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user of the note",
			Required:    true,
		},
	},
}

type Notes struct {
	mu       sync.RWMutex
	fileName string
	Notes    map[string]string
}

func LoadNotes(path string) (*Notes, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		f, err = os.Create(path)
	}
	if err != nil {
		return nil, err
	}

	notes := &Notes{
		fileName: path,
		Notes:    make(map[string]string),
	}

	err = json.NewDecoder(f).Decode(&notes.Notes)
	switch err {
	case nil, io.EOF:
		// do nothing
	default:
		return nil, fmt.Errorf("could not unmarshal %s: %v", path, err)
	}

	return notes, nil
}

func (n *Notes) Save() error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	lit.Debug("Saving repository...")
	data, err := json.MarshalIndent(n.Notes, "", "\t")
	if err != nil {
		return fmt.Errorf("marshalling notes: %v", err)
	}

	if err := os.WriteFile(n.fileName, data, os.ModePerm); err != nil {
		return fmt.Errorf("saving %s: %v", n.fileName, err)
	}

	return nil
}

func (n *Notes) handleNotesRaw(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	switch ic.Type {
	case discordgo.InteractionApplicationCommand:
		return n.handleNotesCmd(ds, ic)
	case discordgo.InteractionModalSubmit:
		return n.handleSubmitNotes(ds, ic)
	}
	return nil, fmt.Errorf("unknown interaction type for notes command")
}

func (n *Notes) handleNotesApp(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var herder bool
	for _, role := range ic.Member.Roles {
		if role == HerderRoleID {
			herder = true
			break
		}
	}

	if !(ic.Member.User.ID == AdminUserID || herder) {
		return nil, fmt.Errorf("These commands are only for herders and above.")
	}

	u, err := ds.User(ic.ApplicationCommandData().TargetID)
	if err != nil {
		return nil, fmt.Errorf("The selected user was unable to be found")
	}

	notes := n.Notes[u.ID]

	return &discordgo.InteractionResponseData{
		Title:    "Viewing notes for " + u.String(),
		CustomID: u.ID,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						Label:    "Note",
						Style:    discordgo.TextInputParagraph,
						CustomID: "note",
						Value:    notes,
					},
				},
			},
		},
	}, nil
}

func (n *Notes) handleNotesCmd(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var herder bool
	for _, role := range ic.Member.Roles {
		if role == HerderRoleID {
			herder = true
			break
		}
	}

	if !(ic.Member.User.ID == AdminUserID || herder) {
		return nil, fmt.Errorf("These commands are only for herders and above.")
	}

	u := ic.ApplicationCommandData().Options[0].UserValue(ds)
	if u == nil {
		return nil, fmt.Errorf("The selected user was unable to be found")
	}

	notes := n.Notes[u.ID]
	return &discordgo.InteractionResponseData{
		Title:    "Viewing notes for " + u.String(),
		CustomID: u.ID,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						Label:    "Note",
						Style:    discordgo.TextInputParagraph,
						CustomID: "note",
						Value:    notes,
					},
				},
			},
		},
	}, nil
}

func (n *Notes) handleSubmitNotes(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	var herder bool
	for _, role := range ic.Member.Roles {
		if role == HerderRoleID {
			herder = true
			break
		}
	}

	if !(ic.Member.User.ID == AdminUserID || herder) {
		return nil, fmt.Errorf("These commands are only for herders and above.")
	}

	data := ic.ModalSubmitData()
	u, err := ds.User(data.CustomID)
	if err != nil {
		lit.Error("notes(submit): fetch user: %v", err)
		return nil, fmt.Errorf("Could not update notes: unable to fetch user of note.")
	}

	input := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput)

	n.mu.Lock()
	n.Notes[data.CustomID] = input.Value
	n.mu.Unlock()
	if err := n.Save(); err != nil {
		lit.Error("notes(submit): saving: %v", err)
		return nil, fmt.Errorf("Could not update notes: unable to save.")
	}

	return EphemeralResponse("Updated notes for " + u.String()), nil
}
