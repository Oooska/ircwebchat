package models

import "errors"

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

	//Messages(timestamp time.Time, cnt int) ([]irc.Message, error)
	//SaveMessage(acct Account, msg irc.Message) error
}
