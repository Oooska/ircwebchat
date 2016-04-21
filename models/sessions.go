package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

const (
	TTL = 6 * time.Hour //Time cookie is valid for
)

//NewSessions returns an object to keep track of web sessions
func NewSessions() Sessions {
	s := sessions{sessions: make(map[string]session)}
	return s
}

//Sessions maintains a list of logged in clients and their sessionid
type Sessions interface {
	Start(Account) (string, time.Time)
	Delete(id string)
	Lookup(id string) (Account, error)
}

type sessions struct {
	sessions map[string]session
}

//Start creates a new sessionID hash for the specified user
func (sess sessions) Start(acct Account) (string, time.Time) {
	hash := generateHash(acct.Username(), acct.Password())
	s := session{id: hash, account: acct, expires: time.Now().Add(TTL)}
	sess.sessions[hash] = s
	return hash, s.expires
}

//Delete removes the specified sessionID
func (sess sessions) Delete(id string) {
	delete(sess.sessions, id)
}

//Lookup returns the account associated with a specific sessionID
//Returns an error if session is expired or session is not found
func (sess sessions) Lookup(id string) (Account, error) {
	s, ok := sess.sessions[id]
	if !ok {
		return nil, errors.New("Session Expired")
	}

	if time.Now().After(s.expires) {
		sess.Delete(id)
		return nil, errors.New("Session Expired")
	}
	return s.account, nil
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
