package models

import (
	"errors"
	"time"

	"github.com/oooska/irc"
)

var persistenceInstance Persistence

func NewPersistenceInstance(driver string) (Persistence, error) {
	var p Persistence
	if driver == "sqlite3" {
		persistenceInstance = &sqlite3{}
		return persistenceInstance, nil
	}

	//TODO: Support an in-memory persistence object
	return p, errors.New("SQL Driver not supported")
}

type Persistence interface {
	Start(filename string) error //Opens db and connects to it
	Stop() error                 //Closes db
	Init() error                 //Creates tables

	account(username string) (account, error)
	saveAccount(acct *account) error
	activeAccounts() ([]account, error)

	session(id string) (session, error)
	saveSession(s session) error
	deleteSession(id string) error

	settings(account Account) (Settings, error)
	saveSettings(s settings) error

	messages(acct Account, channel string, timestamp time.Time, count int) ([]irc.Message, error)
	saveMessage(acct Account, msg irc.Message) error
}
