package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string

var buffers = make([][][]byte, 0)

func main() {

	if token == "" {
		fmt.Println("No token provided. Please run: airhorn -t <bot token>")
		token = os.Getenv("token")
		fmt.Println(token)
	}

	encode()

	// Load the sound file.
	go createDisord()
	startServer()


}

func createDisord(){

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Airhorn is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateStatus(0, "!kmd")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!airhorn"
	if strings.HasPrefix(m.Content, "!kmd") {

		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				err, s := playSound(s, g.ID, vs.ChannelID)
				if err != nil {
					fmt.Println("Here Error playing sound:", err)
					s.Close()
					s.Open()
				}

				return
			}
		}
	}
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}
	/*
	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "Airhorn is ready! Type !airhorn while in a voice channel to play a sound.")
			return
		}
	}
	*/
}

func loadSounds(){
	buffers = make([][][]byte, 0)

	files, err := ioutil.ReadDir("./dca")
	if err != nil {
		fmt.Println("Error loading sound: ", err)
	}
	for _, f := range files {
		err :=loadSound("./dca/"+f.Name())
		if err != nil {
		}
	}

	fmt.Printf("Loaded %d sounds \n", len(buffers))
	if err != nil {
		fmt.Println("Error loading sound: ", err)
		fmt.Println("Please copy $GOPATH/src/github.com/bwmarrin/examples/airhorn/airhorn.dca to this directory.")
	}
}

// loadSound attempts to load an encoded sound file from disk.
func loadSound(path string) error {
	var buffer = make([][]byte, 0)

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			buffers = append(buffers, buffer)
			return nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
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


	return nil
}

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (err error, session *discordgo.Session) {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err, s
	}
	// Sleep for a specified amount of time before playing the sound
	//!time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range buffers[rand.Intn(len(buffers))] {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	//time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()
	return nil,nil
}
