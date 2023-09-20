package persistence

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var ErrDBAlreadyOpen = errors.New("database already opened")

type RemindPersistence struct {
	db *sql.DB

	stInsert *sql.Stmt
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
	UserID     string
	TargetID   string
	Reminder   string
	RemindTime time.Time
}

func (p *RemindPersistence) SaveReminder(r *Reminder) error {
	_, err := p.stInsert.Exec(r.UserID, r.TargetID, r.Reminder, r.RemindTime, 1)
	if err != nil {
		return err
	}

	return nil
}
