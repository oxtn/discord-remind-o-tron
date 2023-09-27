package commandhandlers

import (
	"context"
	"discord-remind-o-tron/persistence"
	"errors"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

var ErrDBNotOpen = errors.New("database not already opened")

type RemindBackground struct {
	db *persistence.RemindPersistence
	s  *discordgo.Session

	context context.Context
	cancel  context.CancelFunc
}

func NewRemindBackground(db *persistence.RemindPersistence, s *discordgo.Session) (*RemindBackground, error) {
	if !db.IsOpen() {
		return nil, ErrDBNotOpen
	}

	return &RemindBackground{
		s:  s,
		db: db,
	}, nil
}

func (r *RemindBackground) FinishPerformReminders() error {
	if r.context == nil || r.cancel == nil {
		return nil
	}

	r.cancel()
	return nil
}

func (r *RemindBackground) PerformReminders() {
	r.context, r.cancel = context.WithCancel(context.Background())
	defer r.cancel()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Starting Reminders...")

	for {
		select {
		case <-r.context.Done():
			log.Println("Stopping Reminders...")
			return
		case <-ticker.C:
			reminders, err := r.db.FetchScheduledWithinMinutes(5, 20)
			if err != nil {
				log.Print(err)
			}

			for _, reminder := range reminders {
				processReminder(r, reminder)
			}

		}
	}
}

func processReminder(r *RemindBackground, reminder *persistence.Reminder) {
	err := sendChannelMessage(r.s, reminder.TargetID, reminder.Reminder)

	if err == nil {
		reminder.MarkSent()
	} else {
		reminder.MarkError()
		log.Println(err)
	}
}
