package commandhandlers

import (
	"context"
	"discord-remind-o-tron/persistence"
	"errors"
	"log"
	"time"
)

var ErrDBNotOpen = errors.New("database not already opened")

type RemindBackground struct {
	db *persistence.RemindPersistence

	context context.Context
	cancel  context.CancelFunc
}

func NewRemindBackground(db *persistence.RemindPersistence) (*RemindBackground, error) {
	if !db.IsOpen() {
		return nil, ErrDBNotOpen
	}

	return &RemindBackground{
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
			rows, err := r.db.FetchScheduledWithinMinutes(5, 20)
			if err != nil {
				log.Print(err)
			}

			for _, row := range rows {
				log.Println(row)
			}

		}
	}
}
