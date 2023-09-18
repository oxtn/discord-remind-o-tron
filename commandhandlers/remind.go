package commandhandlers

import (
	"discord-remind-o-tron/util"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Remind(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	util.LogJson(data)

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(data.Options))
	for _, opt := range data.Options {
		optionMap[opt.Name] = opt
	}

	when, err := time.ParseDuration(optionMap["when"].StringValue())
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Scheduling for: %v", when)

	time.AfterFunc(when,
		func() {
			log.Println("Timer Elapsed")
			sendChannelMessage(s, i.ChannelID)
		})

	err = s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Got it",
			},
		},
	)

	if err != nil {
		log.Println(err)
	}
}

func sendChannelMessage(s *discordgo.Session, channelId string) {
	_, err := s.ChannelMessageSend(channelId, "Time's up")

	if err != nil {
		log.Println(err)
	}
}
