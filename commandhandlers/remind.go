package commandhandlers

import (
	"discord-remind-o-tron/persistence"
	"discord-remind-o-tron/util"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	db *persistence.RemindPersistence
)

func Remind(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if db == nil {
		db = persistence.NewRemindPersistence()
		err := db.Open()
		if err != nil {
			log.Fatalln(err)
		}
	}

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

	target := resolveTarget(s, i, optionMap)
	message := optionMap["message"].StringValue()
	instant := time.Now().Add(when)

	time.AfterFunc(when,
		func() {
			sendChannelMessage(s, target.ID, message)
		})

	db.SaveReminder(&persistence.Reminder{
		UserID:     i.Member.User.ID,
		TargetID:   target.ID,
		Reminder:   message,
		RemindTime: instant.UTC(),
	})

	err = s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("I'll remind %v about '%v' at %v.", target.FriendlyName, message, instant.Format("January 2, 2006 at 3:04pm (MST)")),
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

func resolveTarget(s *discordgo.Session, i *discordgo.InteractionCreate, options map[string]*discordgo.ApplicationCommandInteractionDataOption) Target {
	channel, ok := options["channel"]

	if ok {
		channel := channel.ChannelValue(nil)
		return Target{ID: channel.ID, FriendlyName: fmt.Sprintf("<#%s>", channel.ID)}
	}

	userChannel, err := s.UserChannelCreate(i.Member.User.ID)
	if err != nil {
		log.Println(err)
	}

	return Target{ID: userChannel.ID, FriendlyName: "you"}
}

type Target struct {
	ID           string
	FriendlyName string
}
