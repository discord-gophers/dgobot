package commands

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

func handleGopher(ds *discordgo.Session, ic *discordgo.InteractionCreate) {

	if rand.Intn(15) <= 1 {
		ds.ChannelMessageSend(ic.ChannelID, `https://www.youtube.com/watch?v=iay2wUY8uqA`)
		return
	}

	// get channel
	c, err := ds.State.Channel(ic.ChannelID)
	if err != nil {

		// Try fetching via REST API
		c, err = ds.Channel(ic.ChannelID)
		if err != nil {
			lit.Error("getting channel, %s", err)
			ds.ChannelMessageSend(ic.ChannelID, `Looks like all the Gophers are sleeping right now`)
			return
		}
	}

	// Find the guild for that channel.
	g, err := ds.State.Guild(c.GuildID)
	if err != nil {

		// Try fetching via REST API
		g, err = ds.Guild(ic.ChannelID)
		if err != nil {
			lit.Error("getting guild, %s", err)
			ds.ChannelMessageSend(ic.ChannelID, `Looks like all the Gophers are sleeping right now`)
			return
		}
	}

	// Look for the message sender in that guild's current voice states.
	for _, vs := range g.VoiceStates {

		if vs.UserID == ic.Message.Author.ID {
			err = playSound(ds, g.ID, vs.ChannelID)
			if err != nil {
				lit.Error("play sound, %s", err)
				ds.ChannelMessageSend(ic.ChannelID, `Looks like all the Gophers are sleeping right now`)
			}
			return
		}

	}

	ds.ChannelMessageSend(ic.ChannelID, `Sorry, you must be in a voice channel to hear the mighty gopher.`)
}

var gopherlock sync.Mutex

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (err error) {

	gopherlock.Lock()
	defer gopherlock.Unlock()

	var buffer = make([][]byte, 0)
	var opuslen int16

	files, err := ioutil.ReadDir("sounds/")
	if err != nil {
		return err
	}

	gopher := files[rand.Intn(len(files))]
	log.Println("Playing", gopher.Name())

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	defer vc.Disconnect()
	if err != nil {
		fmt.Println("Cannot join voice, ", err)
		return err
	}

	// read the file
	file, err := os.Open("sounds/" + gopher.Name())
	if err != nil {
		fmt.Printf("Error opening dca file %s, %s", file.Name(), err)
		return err
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
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		if opuslen < 5 || opuslen > 500 {
			log.Printf("Something wrong with opuslen : %d\n", opuslen)
			return fmt.Errorf("bad size opuslen")
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
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
