package commandhandlers

import (
	"discord-remind-o-tron/persistence"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/markusmobius/go-dateparser"
)

type Remind struct {
	db         *persistence.RemindPersistence
	background *RemindBackground
	dtCfg      *dateparser.Configuration
}

func NewRemind(s *discordgo.Session) (*Remind, error) {
	db := persistence.NewRemindPersistence()
	err := db.Open()
	if err != nil {
		return nil, err
	}

	background, err := NewRemindBackground(db, s)
	if err != nil {
		return nil, err
	}

	return &Remind{
		db:         db,
		background: background,
		dtCfg:      &dateparser.Configuration{PreferredDateSource: dateparser.Future},
	}, nil
}

func (r *Remind) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(data.Options))
	for _, opt := range data.Options {
		optionMap[opt.Name] = opt
	}

	whenStr := optionMap["when"].StringValue()
	when, err := dateparser.Parse(r.dtCfg, whenStr)
	if err != nil {
		log.Println(err)

		response := fmt.Sprintf(":x: Your provided time of '%v' wasn't recognized.", whenStr)
		respondInteraction(s, i, response, true)
		return
	}

	if when.Time.Before(time.Now().UTC()) {
		response := fmt.Sprintf(":x: Your provided time of '%v' is in the past.", whenStr)
		respondInteraction(s, i, response, true)
		return
	}

	log.Printf("Scheduling for: %v", when)

	target := resolveTarget(s, i, optionMap)
	message := optionMap["message"].StringValue()

	var reminder = r.db.NewReminder(i.Member.User.ID, target.ID, message, when.Time.UTC())
	reminder.Save()

	response := fmt.Sprintf("I'll remind %v about '%v' at %v.", target.FriendlyName, message, when.Time.Format("January 2, 2006 at 3:04pm (MST)"))
	respondInteraction(s, i, response, true)
}

func (r *Remind) Start() {
	go r.background.PerformReminders()
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

func respondInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, content string, ephemeral bool) {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   flags,
			},
		},
	)

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
