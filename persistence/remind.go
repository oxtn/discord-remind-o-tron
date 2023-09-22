package persistence

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ReminderStatus int

const (
	ReminderStatusUnknown   ReminderStatus = 0
	ReminderStatusScheduled ReminderStatus = 1
	ReminderStatusSent      ReminderStatus = 2
	ReminderStatusError     ReminderStatus = 3
)

var ErrReminderNotNew = errors.New("reminder already exists")
var ErrDBAlreadyOpen = errors.New("database already opened")

type RemindPersistence struct {
	db *sql.DB

	stInsert       *sql.Stmt
	stUpdateStatus *sql.Stmt
}

func NewRemindPersistence() *RemindPersistence {

	return &RemindPersistence{}
}

func (p *RemindPersistence) Open() error {
	var err error

	if p.db != nil {
		return ErrDBAlreadyOpen
	}

	p.db, err = sql.Open("sqlite3", "reminders.db")
	if err != nil {
		return err
	}

	_, err = p.db.Exec(`
	CREATE TABLE IF NOT EXISTS reminders (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		UserID TEXT,
		TargetID TEXT,
		Reminder TEXT,
		RemindTime DATETIME,
		Status INTEGER
		);`)
	if err != nil {
		return err
	}

	_, err = p.db.Exec("CREATE INDEX IF NOT EXISTS IDX_RemindTime ON reminders (RemindTime, TargetID, Reminder);")
	if err != nil {
		return err
	}

	p.stInsert, err = p.db.Prepare("INSERT INTO reminders (UserId, TargetID, Reminder, RemindTime, Status) VALUES(?,?,?,?,?)")
	if err != nil {
		return err
	}

	p.stUpdateStatus, err = p.db.Prepare("UPDATE reminders SET Status = ? WHERE Id = ?")
	if err != nil {
		return err
	}

	return nil
}

func (p *RemindPersistence) Close() error {
	err := p.stInsert.Close()
	if err != nil {
		return err
	}

	err = p.db.Close()
	if err != nil {
		return err
	}

	return nil
}

type Reminder struct {
	Id         int64
	UserID     string
	TargetID   string
	Reminder   string
	RemindTime time.Time

	persistence *RemindPersistence
}

func (p *RemindPersistence) NewReminder(UserID string, TargetID string, Text string, RemindTime time.Time) *Reminder {

	return &Reminder{
		Id:         0,
		UserID:     UserID,
		TargetID:   TargetID,
		Reminder:   Text,
		RemindTime: RemindTime,

		persistence: p,
	}
}

func (r *Reminder) Save() error {
	if r.Id != 0 {
		return ErrReminderNotNew
	}

	result, err := r.persistence.stInsert.Exec(r.UserID, r.TargetID, r.Reminder, r.RemindTime, ReminderStatusScheduled)
	if err != nil {
		return err
	}

	r.Id, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (r *Reminder) MarkSent() error {
	_, err := r.persistence.stUpdateStatus.Exec(ReminderStatusSent, r.Id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reminder) MarkError() error {
	_, err := r.persistence.stUpdateStatus.Exec(ReminderStatusError, r.Id)
	if err != nil {
		return err
	}

	return nil
}
