package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	var (
		Token   = os.Getenv("TOKEN")
		appID   = os.Getenv("APPID")
		guildID = os.Getenv("GUILDID")
	)

	session, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	_, err = session.ApplicationCommandBulkOverwrite(appID, guildID, []*discordgo.ApplicationCommand{
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
	})
	if err != nil {
		log.Fatalf("error creating Discord commands: %v", err)
	}

	session.AddHandler(interactionCreate)

	session.Identify.Intents = discordgo.IntentsGuildMessages

	err = session.Open()
	if err != nil {
		log.Fatalf("error opening connection: %v", err)
	}

	defer session.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down...")
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	str, _ := json.MarshalIndent(data, "", "\t")
	fmt.Println(string(str))
	switch data.Name {
	case "remind":
		err := s.InteractionRespond(
			i.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "K",
				},
			},
		)
		if err != nil {
			fmt.Println(err)
		}
	}
}
