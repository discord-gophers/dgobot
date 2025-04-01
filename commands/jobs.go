package commands

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

const (
	jobMessageTitle   = "Want To Post a Job? READ ME!"
	jobMessageContent = `
# Welcome to the Go jobs forum! Here, you will see __Go related__ job listings.

### This is a forum for job listings only. For-hire listings are not permitted.
### If you want make a post here, please make sure to enable direct messages from others on this server.

### NOTE: __All job postings here must have the following information:__

- Duration (examples: indefinite, one time, contract, ...)
- Expected pay (examples: $45/hr, $85,000-$125,000)
- Job description (provide at least 3 sentences/bullets)
- Location (examples: Remote, Hybrid + Location, New York)

### You are __strongly__ encouraged to include this information as well:

- Job responsibilities
- Other desired skills/Nice to haves

### RECEIVING ACCESS TO POST

To avoid excess spam, please click the button to provide information regarding your job listing. **You only have to request access once!**
You are welcome to modify, update, and post as many applications as you wish once accepted. However, please limit re-posting similar postings to once every 2 weeks.

--

-# Job listings may be removed upon discretion of the mod team for any reason.
`
	jobAcceptedMessage = `
## Congratulations, your posting request has been accepted.

You are now able to post listings. However, please limit re-posting similar postings to once every 2 weeks.

Here is your post template. Modify as needed before posting:
### %s
` + "```" + `
**Location:**
%s

**Expected Pay:**
%s

**Description:**
%s

%s
` + "```" + `

### Important notes:
- If your listing has a link to further information, or an listing board with multiple opportunites *with Go*, include it.
- Spamming or misuse will result in permanent revocation to posting.
`

	jobRejectedMessage = `
## Unfortunately, your posting request has been rejected.

### Reason:
%s

For further information, please contact a moderator.

-# You can submit another request 2 weeks after your initial request date.
`
)

var (
	jobSubmitMu          sync.Mutex
	jobSubmitCooldown    = make(map[string]time.Time)
	jobMessageComponents = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    "Request Access",
					Style:    discordgo.SuccessButton,
					CustomID: "jobs:request",
				},
			},
		},
	}
)

func init() {
	Commands[cmdJobs.Name] = &Command{
		ApplicationCommand: cmdJobs,
		Handler:            handleJobsRaw,
	}
}

var cmdJobs = &discordgo.ApplicationCommand{
	Type:                     discordgo.ChatApplicationCommand,
	Name:                     "jobs",
	Description:              "Job Listing System Manager",
	DefaultMemberPermissions: &[]int64{discordgo.PermissionModerateMembers}[0],
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "post-message",
			Description: "Post Job Listing info message",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "target",
					Description: "The URL to add",
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "request-access",
			Description: "Request access to job listing forum",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "approve-user",
			Description: "Give access to post job listings",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user to give access to",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "revoke-user",
			Description: "Revoke access to post job listings",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user to revoke access from",
					Required:    true,
				},
			},
		},
	},
}

func handleJobsRaw(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	switch ic.Type {
	case discordgo.InteractionApplicationCommand:
		return handleJobsCommand(ds, ic)
	case discordgo.InteractionMessageComponent:
		return handleJobsMsg(ds, ic)
	case discordgo.InteractionModalSubmit:
		return handleJobsSubmit(ds, ic)
	}
	return nil, fmt.Errorf("unknown interaction type for notes command")
}

func handleJobsCommand(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	var f func(*discordgo.Session, *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error)
	var check bool

	cmd := ic.ApplicationCommandData().Options[0].Name
	switch cmd {
	case "post-message":
		f, check = handleJobsPost, true
	case "request-access":
		f, check = handleJobsRequest, false
	case "approve-user":
		f, check = handleJobsAdd, true
	case "revoke-user":
		f, check = handleJobsRemove, true
	}
	if f == nil {
		return nil, fmt.Errorf("Invalid option: `%s`.", cmd)
	}

	if check && !isHerder(ic) {
		return nil, fmt.Errorf("These commands are only for herders and above.")
	}
	return f(ds, ic)
}

func handleJobsPost(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	target := ic.ChannelID
	for _, opt := range ic.ApplicationCommandData().Options[0].Options {
		switch opt.Name {
		case "target":
			target = opt.ChannelValue(nil).ID
		}
	}

	ch, err := ds.State.Channel(target)
	if err != nil {
		ch, err = ds.Channel(target)
		if err != nil {
			return nil, fmt.Errorf("could not get channel data for <#%s> (do I have permissions to see it?)", target)
		}
	}

	switch ch.Type {
	case discordgo.ChannelTypeGuildForum:
		_, err = ds.ForumThreadStartComplex(target, &discordgo.ThreadStart{
			Name: jobMessageTitle,
		}, &discordgo.MessageSend{
			Content:    jobMessageContent,
			Components: jobMessageComponents,
		})
	case discordgo.ChannelTypeGuildText:
		_, err = ds.ChannelMessageSendComplex(target, &discordgo.MessageSend{
			Content:    jobMessageContent,
			Components: jobMessageComponents,
		})
	}

	return EphemeralResponse("message sent"), err
}

func handleJobsRequest(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	if ic.Member == nil {
		return nil, fmt.Errorf("This application can only be used from within the Gophers Server")
	}

	var hasRole bool
	for _, role := range ic.Member.Roles {
		if role == JobsRoleID {
			hasRole = true
		}
	}
	// Herders can test submit as needed.
	if hasRole && !isHerder(ic) {
		return nil, fmt.Errorf("You already have access to post listings!\nSubmit a new post in this forum directly.")
	}

	jobSubmitMu.Lock()
	defer jobSubmitMu.Unlock()

	// Lazily clean all cooldowns, no need for background goroutine.
	// Realistically this map shouldn't ever contain more than a 1,000 entries.
	now := time.Now()
	for k, cd := range jobSubmitCooldown {
		if now.After(cd) {
			delete(jobSubmitCooldown, k)
			continue
		}
		if k == ic.Member.User.ID && !isHerder(ic) {
			return nil, fmt.Errorf("You have already submitted a request recently. Please wait before submitting again, or contact a moderator.")
		}
	}

	return &discordgo.InteractionResponseData{
		Title:    "Job Posting Access Request",
		CustomID: "jobs:submit",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						Label:       "Post Title",
						Style:       discordgo.TextInputShort,
						CustomID:    "title",
						Required:    true,
						Placeholder: "[Onsite] Senior Gopher Herder at Gophers Inc.",
						MaxLength:   256,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						Label:       "Location",
						Style:       discordgo.TextInputShort,
						CustomID:    "location",
						Required:    true,
						Placeholder: "[Central America] [Dense Forests]",
						MaxLength:   256,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						Label:       "Expected Pay",
						Style:       discordgo.TextInputShort,
						CustomID:    "pay",
						Required:    true,
						Placeholder: "10,000 acorns per month (or equivalent in local currency)",
						MaxLength:   256,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						Label:       "Job Description",
						Style:       discordgo.TextInputParagraph,
						CustomID:    "description",
						Required:    true,
						Placeholder: "- Maintain underground infrastructure \n- Work with local wildlife to ensure peaceful coexistence.",
						MaxLength:   2048,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						Label:       "Additional Information",
						Style:       discordgo.TextInputParagraph,
						CustomID:    "additional",
						Required:    false,
						Placeholder: "Must be comfortable working in an environment where everything is either trying to burrow or bite.",
						MaxLength:   2048,
					},
				},
			},
		},
	}, nil
}

func handleJobsSubmit(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	if ic.Member == nil {
		return nil, fmt.Errorf("This application can only be used from within the Gophers Server")
	}

	data := ic.ModalSubmitData()

	parts := strings.Split(data.CustomID, ":")
	if len(parts) == 3 && parts[1] == "reject" {
		reason := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput)
		return handleJobsReject(ds, ic, reason.CustomID, parts[2], reason.Value)
	}

	var title, loc, pay, desc, addl string
	for _, c := range data.Components {
		inp := c.(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput)
		switch inp.CustomID {
		case "title":
			title = inp.Value
		case "location":
			loc = inp.Value
		case "pay":
			pay = inp.Value
		case "description":
			desc = inp.Value
		case "additional":
			addl = inp.Value
		}
	}

	ds.ChannelMessageSendComplex(JobsChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "New Job Posting Access Request",
				Footer: &discordgo.MessageEmbedFooter{
					IconURL: ic.Member.AvatarURL(""),
					Text:    "Pending Decision",
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  "<@" + ic.Member.User.ID + ">",
						Inline: true,
					},
					{
						Name:   "Member Joined",
						Value:  fmt.Sprintf("<t:%d:R>", ic.Member.JoinedAt.Unix()),
						Inline: true,
					},
					{
						Name:  "Title",
						Value: title,
					},
					{
						Name:  "Location",
						Value: loc,
					},
					{
						Name:  "Pay",
						Value: pay,
					},
					{
						Name:  "Description",
						Value: desc,
					},
					{
						Name:  "Additional Info",
						Value: addl,
					},
				},
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Accept",
						Style:    discordgo.SuccessButton,
						CustomID: "jobs:accept:" + ic.Member.User.ID,
					},
					discordgo.Button{
						Label:    "Reject",
						Style:    discordgo.DangerButton,
						CustomID: "jobs:reject:" + ic.Member.User.ID,
					},
				},
			},
		},
	})

	jobSubmitMu.Lock()
	defer jobSubmitMu.Unlock()

	jobSubmitCooldown[ic.Member.User.ID] = time.Now().Add(time.Hour * 24 * 14)
	return EphemeralResponse("Your request has been submitted! You should hear back shortly.\n-# Please make sure your DMs are enabled."), nil
}

func handleJobsMsg(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	data := ic.MessageComponentData()
	parts := strings.Split(data.CustomID, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("An internal error has occured. Please contact the mods.")
	}

	if parts[1] == "request" {
		return handleJobsRequest(ds, ic)
	}

	if !isHerder(ic) {
		return nil, fmt.Errorf("This action is only for herders and above.")
	}

	switch parts[1] {
	case "accept":
		return handleJobsAccept(ds, ic, parts[2])
	case "reject":
		return &discordgo.InteractionResponseData{
			Title:    "Reject Request",
			CustomID: data.CustomID,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							Label:       "Reason",
							Style:       discordgo.TextInputParagraph,
							CustomID:    ic.Message.ID,
							Required:    true,
							Placeholder: "Lol get rekt crypto bro",
							MaxLength:   1024,
						},
					},
				},
			},
		}, nil
	}

	return EphemeralResponse(data.CustomID), nil
}

func handleJobsAccept(ds *discordgo.Session, ic *discordgo.InteractionCreate, userID string) (*discordgo.InteractionResponseData, error) {
	if ic.Member == nil {
		return nil, fmt.Errorf("This application can only be used from within the Gophers Server")
	}

	if err := ds.GuildMemberRoleAdd(ic.GuildID, userID, JobsRoleID); err != nil {
		lit.Error("could not assign role: %v", err)
		return nil, fmt.Errorf("Could not assign user with role.")
	}

	if _, err := ds.ChannelMessageSendReply(ic.ChannelID, "Accepted by <@"+ic.Member.User.ID+">", &discordgo.MessageReference{
		GuildID:   ic.GuildID,
		ChannelID: ic.ChannelID,
		MessageID: ic.Message.ID,
	}); err != nil {
		lit.Error("could not reply to message: %v", err)
	}

	dm, err := ds.UserChannelCreate(userID)
	if err == nil {
		fields := ic.Message.Embeds[0].Fields
		if _, err := ds.ChannelMessageSend(dm.ID, fmt.Sprintf(jobAcceptedMessage,
			fields[2].Value, fields[3].Value, fields[4].Value, fields[5].Value, fields[6].Value)); err != nil {
			lit.Error("could not send dm: %v", err)
		}
	}

	ic.Message.Components = nil
	ic.Message.Embeds[0].Footer.Text = "Accepted by " + ic.Member.User.Username
	return UpdateMessageResponse(ic.Message), nil
}

func handleJobsReject(ds *discordgo.Session, ic *discordgo.InteractionCreate, messageID, userID, reason string) (*discordgo.InteractionResponseData, error) {
	if _, err := ds.ChannelMessageSendReply(ic.ChannelID, "Rejected by <@"+ic.Member.User.ID+">", &discordgo.MessageReference{
		GuildID:   ic.GuildID,
		ChannelID: ic.ChannelID,
		MessageID: messageID,
	}); err != nil {
		lit.Error("could not reply to message: %v", err)
	}

	dm, err := ds.UserChannelCreate(userID)
	if err == nil {
		if _, err := ds.ChannelMessageSend(dm.ID, fmt.Sprintf(jobRejectedMessage, reason)); err != nil {
			lit.Error("could not send dm: %v", err)
		}
	}

	ic.Message.Components = nil
	ic.Message.Embeds[0].Footer.Text = "Rejected by " + ic.Member.User.Username
	return UpdateMessageResponse(ic.Message), nil
}

func handleJobsAdd(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	userID := ic.ApplicationCommandData().Options[0].Options[0].UserValue(nil).ID
	if err := ds.GuildMemberRoleAdd(ic.GuildID, userID, JobsRoleID); err != nil {
		lit.Error("could not assign role: %v", err)
		return nil, fmt.Errorf("Could not assign user with role.")
	}
	return EphemeralResponse("Roles updated"), nil
}

func handleJobsRemove(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	userID := ic.ApplicationCommandData().Options[0].Options[0].UserValue(nil).ID
	if err := ds.GuildMemberRoleRemove(ic.GuildID, userID, JobsRoleID); err != nil {
		lit.Error("could not assign role: %v", err)
		return nil, fmt.Errorf("Could not assign user with role.")
	}
	return EphemeralResponse("Role updated"), nil
}
