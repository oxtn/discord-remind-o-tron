package main

import (
	"discord-remind-o-tron/commandhandlers"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Token = os.Getenv("TOKEN")
)

var remindHandler *commandhandlers.Remind
var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
var session *discordgo.Session

func init() {
	var err error
	session, err = discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages
}

func init() {
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "remind",
			Description: "Set a Reminder",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "message",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Reminder message",
					Required:    true,
				},
				{
					Name:        "when",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "When to remind",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "Channel to Remind",
					ChannelTypes: []discordgo.ChannelType{
						discordgo.ChannelTypeGuildText,
					},
					Required: false,
				},
			},
		},
	}

	var err error
	remindHandler, err = commandhandlers.NewRemind(session)
	if err != nil {
		log.Fatalln(err)
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"remind": remindHandler.Handler,
	}
}

func init() {
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	})
}

func main() {
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err := session.Open()
	if err != nil {
		log.Fatalf("error opening connection: %v", err)
	}

	for _, v := range commands {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}

	defer session.Close()
	defer remindHandler.Close()

	go remindHandler.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down...")
}
