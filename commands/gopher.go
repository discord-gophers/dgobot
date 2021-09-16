package commands

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

func init() {
	Commands[cmdGopher.Name] = &Command{
		ApplicationCommand: cmdGopher,
		Handler:            handleGopher,
	}
}

var cmdGopher = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "gopher",
	Description: "Hear the call of the Gopher!",
}

var gopherErr = fmt.Errorf("Looks like all gophers are asleep right now")

func handleGopher(ds *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	if rand.Intn(15) <= 1 {
		return ContentResponse("https://www.youtube.com/watch?v=iay2wUY8uqA"), nil
	}

	// get channel
	c, err := ds.State.Channel(ic.ChannelID)
	if err != nil {

		// Try fetching via REST API
		c, err = ds.Channel(ic.ChannelID)
		if err != nil {
			lit.Error("gopher: getting channel: %v", err)
			return nil, gopherErr
		}
	}

	// Find the guild for that channel.
	g, err := ds.State.Guild(c.GuildID)
	if err != nil {

		// Try fetching via REST API
		g, err = ds.Guild(ic.ChannelID)
		if err != nil {
			lit.Error("gopher: getting guild: %s", err)
			return nil, gopherErr
		}
	}

	// Look for the message sender in that guild's current voice states.
	for _, vs := range g.VoiceStates {
		if vs.UserID == ic.Message.Author.ID {
			// Call in goroutine to allow functon to return response
			go func() {
				err = playSound(ds, g.ID, vs.ChannelID)
				if err != nil {
					lit.Error("gopher: playing sound: %s", err)
				}
			}()
			return ContentResponse("Dispatching..."), nil
		}
	}

	return nil, fmt.Errorf("Sorry, you must be in avoice channel to hear the mighty gopher.")
}

var gopherlock sync.Mutex

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (err error) {
	gopherlock.Lock()
	defer gopherlock.Unlock()

	buffer := make([][]byte, 0)
	var opuslen int16

	files, err := ioutil.ReadDir("sounds/")
	if err != nil {
		return err
	}

	gopher := files[rand.Intn(len(files))]
	lit.Debug("Playing %s", gopher.Name())

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	defer vc.Disconnect()
	if err != nil {
		return fmt.Errorf("could not join voice: %w", err)
	}

	// read the file
	file, err := os.Open("sounds/" + gopher.Name())
	if err != nil {
		return fmt.Errorf("could not open dca file %s: %v", file.Name(), err)
	}

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			break
		}

		if err != nil {
			return fmt.Errorf("could not read from dca file: %v", err)
		}

		if opuslen < 5 || opuslen > 500 {
			return fmt.Errorf("bad opuslen size: %v", opuslen)
		}

		// Read encoded pcm from dca file.
		inBuf := make([]byte, opuslen)
		if err = binary.Read(file, binary.LittleEndian, &inBuf); err != nil {
			return fmt.Errorf("could not read from dca file: %v", err)
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, inBuf)
	}

	time.Sleep(500 * time.Millisecond)

	// Send the buffer data.
	vc.Speaking(true)
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	// Works a bit like a rate limiter too
	time.Sleep(3 * time.Second)

	return nil
}
