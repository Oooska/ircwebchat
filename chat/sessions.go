package chat

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

const (
	tTL = 6 * time.Hour //Time cookie is valid for
)

type session struct {
	id      string
	account Account
	expires time.Time
}

//Start creates a new sessionID hash for the specified user
func NewSession(acct Account) (string, time.Time, error) {
	hash := generateHash(acct.Username(), acct.Password())
	s := session{id: hash, account: acct, expires: time.Now().Add(tTL)}
	err := persistenceInstance.saveSession(s)
	return hash, s.expires, err
}

//Delete removes the specified sessionID
func DeleteSession(id string) error {
	return persistenceInstance.deleteSession(id)
}

//Lookup returns the account associated with a specific sessionID
//Returns an error if session is expired or session is not found
func LookupSession(id string) (Account, error) {
	s, err := persistenceInstance.session(id)

	if err != nil {
		return nil, err
	}

	if time.Now().After(s.expires) {
		DeleteSession(id)
		return nil, errors.New("Session Expired")
	}

	return s.account, nil
}

func newsession(id string, acct account, expires time.Time) session {
	return session{id: id, account: acct, expires: expires}
}

func generateHash(username, password string) string {
	b := sha256.Sum256([]byte(username + password + time.Now().String()))
	return hex.EncodeToString(b[:32])
}
