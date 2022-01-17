package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID  = flag.String("guild", "922157549631979570", "Test guild ID. If not passed - bot registers commands globally")
	BotToken = os.Getenv("TOKEN")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	if BotToken == "" {
		log.Fatal("Token cannot be empty")
	}

	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "play a song",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "url of the song or video",
					Required:    true,
				},
			},
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Query user for song url
			url := i.ApplicationCommandData().Options[0].StringValue()

			// Send message to the channel
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "[Now Playing:](" + url + ")",
				},
				// TODO: Add a reaction
				// TODO: Read user reactions for a set amount of time then play the song
			})
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}

	defer s.Close()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Gracefully shutdowning; Cleaning up commands")
	for _, v := range commands {
		s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.Name)
	}
}
