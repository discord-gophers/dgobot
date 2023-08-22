package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

func init() {
	macro, err := LoadMacro("macros.json")
	if err != nil {
		panic(err)
	}
	Commands[cmdMacro.Name] = &Command{
		ApplicationCommand: cmdMacro,
		Handler:            macro.handleMacroRaw,
		Autocomplete:       macro.handleMacroAutocomplete,
	}
}

var cmdMacro = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "macro",
	Description: "macros for dumping text",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "set",
			Description: "set a macro",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "key",
					Description:  "macro key",
					Autocomplete: true,
					Required:     true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "get",
			Description: "get a macro",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "key",
					Description:  "macro key",
					Autocomplete: true,
					Required:     true,
				},
			},
		},
	},
}

type Macro struct {
	mu       sync.RWMutex
	fileName string
	Macro    map[string]string
}

func LoadMacro(path string) (*Macro, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		f, err = os.Create(path)
		f.Write([]byte("{}"))
		f.Seek(0, 0)
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	macro := &Macro{
		fileName: path,
		Macro:    make(map[string]string),
	}

	err = json.NewDecoder(f).Decode(&macro.Macro)
	switch err {
	case nil, io.EOF:
		// do nothing
	default:
		return nil, fmt.Errorf("could not unmarshal %s: %v", path, err)
	}

	return macro, nil
}

func (m *Macro) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lit.Debug("Saving repository...")
	data, err := json.MarshalIndent(m.Macro, "", "\t")
	if err != nil {
		return fmt.Errorf("marshalling macro: %v", err)
	}

	if err := os.WriteFile(m.fileName, data, 0o644); err != nil {
		return fmt.Errorf("saving %s: %v", m.fileName, err)
	}

	return nil
}

func (m *Macro) handleMacroRaw(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	switch ic.Type {
	case discordgo.InteractionApplicationCommand:
		return m.handleMacroCmd(ds, ic)
	case discordgo.InteractionModalSubmit:
		return m.handleSubmitMacro(ds, ic)
	}

	return nil, fmt.Errorf("unknown interaction type for macro command")
}

func (m *Macro) handleMacroCmd(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(ic.ApplicationCommandData().Options) < 1 {
		return EphemeralResponse("No macro command provided."), nil
	}

	// command
	u := ic.ApplicationCommandData().Options[0].Name
	switch u {
	case "get":
		return m.handleMacroGet(ds, ic)
	case "set":
		return m.handleMacroSet(ds, ic)
	}

	return EphemeralResponse("Unknown macro command."), nil
}

func (m *Macro) handleMacroGet(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(ic.ApplicationCommandData().Options[0].Options) == 0 {
		return EphemeralResponse("No macro key provided."), nil
	}

	u := ic.ApplicationCommandData().Options[0].Options[0].StringValue()
	v, ok := m.Macro[u]
	if !ok {
		return EphemeralResponse("No macro for " + u), nil
	}

	return ContentResponse(v), nil
}

func (m *Macro) handleMacroSet(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(ic.ApplicationCommandData().Options[0].Options) == 0 {
		return EphemeralResponse("No macro key provided."), nil
	}

	u := ic.ApplicationCommandData().Options[0].Options[0].StringValue()
	v := m.Macro[u]

	return &discordgo.InteractionResponseData{
		Title:    "Set Macro for " + u,
		CustomID: "macro:" + u,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						Label:    "Macro",
						Style:    discordgo.TextInputParagraph,
						CustomID: "macro",
						Value:    v,
					},
				},
			},
		},
	}, nil
}

func (m *Macro) handleSubmitMacro(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	var herder bool
	for _, role := range ic.Member.Roles {
		if role == HerderRoleID {
			herder = true
			break
		}
	}

	if ic.Member.User.ID != AdminUserID && !herder {
		return nil, fmt.Errorf("these commands are only for herders and above")
	}

	data := ic.ModalSubmitData()
	input := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput)

	_, id, ok := strings.Cut(data.CustomID, ":")
	if !ok {
		return nil, fmt.Errorf("Invalid custom ID.")
	}

	m.mu.Lock()
	m.Macro[id] = input.Value
	m.mu.Unlock()
	if err := m.Save(); err != nil {
		lit.Error("macro(submit): saving: %v", err)
		return nil, fmt.Errorf("could not update macro: unable to save")
	}

	return EphemeralResponse("Updated macro for " + id), nil
}

func (m *Macro) handleMacroAutocomplete(ds *discordgo.Session, ic *discordgo.InteractionCreate) ([]*discordgo.ApplicationCommandOptionChoice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.Macro))
	for k := range m.Macro {
		keys = append(keys, k)
	}

	return Autocomplete(keys...), nil
}
