package commandhandlers

import (
	"discord-remind-o-tron/persistence"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Remind struct {
	db         *persistence.RemindPersistence
	background *RemindBackground
}

func NewRemind() (*Remind, error) {
	db := persistence.NewRemindPersistence()
	err := db.Open()
	if err != nil {
		return nil, err
	}

	background, err := NewRemindBackground(db)
	if err != nil {
		return nil, err
	}

	return &Remind{
		db:         db,
		background: background,
	}, nil
}

func (r *Remind) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

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

	var reminder = r.db.NewReminder(i.Member.User.ID, target.ID, message, instant.UTC())
	reminder.Save()

	time.AfterFunc(when,
		func() {
			err = sendChannelMessage(s, target.ID, message)

			if err == nil {
				reminder.MarkSent()
			} else {
				reminder.MarkError()
				log.Println(err)
			}
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

func (r *Remind) Start() {
	r.background.PerformReminders()
}

func (r *Remind) Close() error {
	r.background.FinishPerformReminders()

	err := r.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func sendChannelMessage(s *discordgo.Session, targetId string, message string) error {
	log.Printf("Sending %v to %v", message, targetId)

	_, err := s.ChannelMessageSend(targetId, message)

	return err
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
