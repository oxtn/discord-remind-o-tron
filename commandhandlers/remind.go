package commandhandlers

import (
	"discord-remind-o-tron/util"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Remind(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	util.LogJson(i)

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

	targetId := resolveTarget(s, i, optionMap)
	message := optionMap["message"].StringValue()
	instant := time.Now().Add(when)

	time.AfterFunc(when,
		func() {
			sendChannelMessage(s, targetId, message)
		})

	err = s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("I'll remind you about '%v' at %v.", message, instant.Format("January 2, 2006 at 3:04pm (MST)")),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		},
	)

	if err != nil {
		log.Println(err)
	}
}

func sendChannelMessage(s *discordgo.Session, targetId string, message string) {
	log.Printf("Sending %v to %v", message, targetId)

	_, err := s.ChannelMessageSend(targetId, message)

	if err != nil {
		log.Println(err)
	}
}

func resolveTarget(s *discordgo.Session, i *discordgo.InteractionCreate, options map[string]*discordgo.ApplicationCommandInteractionDataOption) string {
	channel, ok := options["channel"]

	if ok {
		return channel.StringValue()
	}

	userChannel, err := s.UserChannelCreate(i.Member.User.ID)
	if err != nil {
		log.Println(err)
	}

	return userChannel.ID
}
