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

	Account(username string) (Account, error)
	SaveAccount(acct Account) error

	//Session(id string) (session, error)
	//SaveSession(s session) error

	//Settings(username string) (Settings, error)
	//SaveSettings(s settings) error

	//Messages(timestamp time.Time, cnt int) ([]irc.Message, error)
	//SaveMessage(acct Account, msg irc.Message) error
}
