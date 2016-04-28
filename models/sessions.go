package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"time"
)

const (
	tTL = 6 * time.Hour //Time cookie is valid for
)

//NewSessions returns an object to keep track of web sessions
func NewSessions() Sessions {
	s := sessions{}
	return s
}

//Sessions maintains a list of logged in clients and their sessionid
type Sessions interface {
	Start(Account) (string, time.Time, error)
	Delete(id string) error
	Lookup(id string) (Account, error)
}

type sessions struct {
}

//Start creates a new sessionID hash for the specified user
func (sess sessions) Start(acct Account) (string, time.Time, error) {
	hash := generateHash(acct.Username(), acct.Password())
	s := session{id: hash, account: acct, expires: time.Now().Add(tTL)}
	err := persistenceInstance.saveSession(s)
	log.Printf("Started session: %+v", s)
	return hash, s.expires, err
}

//Delete removes the specified sessionID
func (sess sessions) Delete(id string) error {
	log.Printf("Deleting session %s", id)
	return persistenceInstance.deleteSession(id)
}

//Lookup returns the account associated with a specific sessionID
//Returns an error if session is expired or session is not found
func (sess sessions) Lookup(id string) (Account, error) {
	s, err := persistenceInstance.session(id)

	if err != nil {
		return nil, err
	}

	if time.Now().After(s.expires) {
		sess.Delete(id)
		return nil, errors.New("Session Expired")
	}

	log.Printf("Looked up session %s, returning session %+v", id, s)
	return s.account, nil
}

func newsession(id string, acct account, expires time.Time) session {
	return session{id: id, account: acct, expires: expires}
}

type session struct {
	id      string
	account Account
	expires time.Time
}

func generateHash(username, password string) string {
	h := sha256.New()
	h.Write([]byte(username + password + time.Now().String()))
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
}
