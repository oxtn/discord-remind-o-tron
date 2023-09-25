package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
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
	sync.RWMutex

	db *sql.DB

	stInsert                 *sql.Stmt
	stUpdateStatus           *sql.Stmt
	stScheduledWithinMinutes *sql.Stmt
}

func NewRemindPersistence() *RemindPersistence {

	return &RemindPersistence{}
}

func (p *RemindPersistence) Open() error {
	var err error

	if p.db != nil {
		return ErrDBAlreadyOpen
	}

	p.Lock()
	defer p.Unlock()

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

	_, err = p.db.Exec("CREATE INDEX IF NOT EXISTS IDX_Status_RemindTime ON reminders (Status, RemindTime, TargetID, Reminder, UserID);")
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

	scheduledWithinMinutes := fmt.Sprintf("SELECT Id, UserID, TargetID, Reminder, RemindTime, Status FROM reminders WHERE Status = %d AND RemindTime <= ? LIMIT ?", ReminderStatusScheduled)
	p.stScheduledWithinMinutes, err = p.db.Prepare(scheduledWithinMinutes)
	if err != nil {
		return err
	}

	return nil
}

func (p *RemindPersistence) Close() error {
	p.Lock()
	defer p.Unlock()

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
	Status     ReminderStatus

	persistence *RemindPersistence
}

func (r Reminder) String() string {
	return fmt.Sprintf("%v - %v - %v", r.Reminder, r.RemindTime, r.Status)
}

func (p *RemindPersistence) NewReminder(userID string, targetID string, text string, remindTime time.Time) *Reminder {

	return &Reminder{
		Id:         0,
		UserID:     userID,
		TargetID:   targetID,
		Reminder:   text,
		RemindTime: remindTime,

		persistence: p,
	}
}

func (p *RemindPersistence) FetchScheduledWithinMinutes(minutes int, maxResults int) ([]*Reminder, error) {
	before := time.Now().UTC().Add(time.Duration(minutes) * time.Minute)

	p.RLock()
	defer p.RUnlock()

	rows, err := p.stScheduledWithinMinutes.Query(before, maxResults)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []*Reminder

	for rows.Next() {
		var row Reminder
		if err := rows.Scan(&row.Id, &row.UserID, &row.TargetID, &row.Reminder, &row.RemindTime, &row.Status); err != nil {
			return nil, err
		}

		results = append(results, &row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Reminder) Save() error {
	if r.Id != 0 {
		return ErrReminderNotNew
	}

	r.persistence.Lock()
	defer r.persistence.Unlock()

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
	r.persistence.Lock()
	defer r.persistence.Unlock()

	_, err := r.persistence.stUpdateStatus.Exec(ReminderStatusSent, r.Id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reminder) MarkError() error {
	r.persistence.Lock()
	defer r.persistence.Unlock()

	_, err := r.persistence.stUpdateStatus.Exec(ReminderStatusError, r.Id)
	if err != nil {
		return err
	}

	return nil
}
