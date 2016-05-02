package chat

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"strings"
)

//Account represents a user account in the system
type Account interface {
	ID() int64
	Username() string
	Password() string
	Email() string
	Active() bool
}

//Account returns the account with the specified username, or an error if none is found
func GetAccount(username string) (Account, error) {
	acct, err := persistenceInstance.account(username)
	if err != nil {
		return acct, err
	}
	if !acct.Active() {
		return account{}, errors.New("This account is no longer active.")
	}
	return acct, nil
}

//Authenticate returns an account if the specified username and password are valid,
//or an error if the login details are wrong or the account is no longer active.
func Authenticate(username, pass string) (Account, error) {
	acct, err := persistenceInstance.account(username)

	if err != nil {
		log.Printf("Recieved error while authenticating: %v", err)
		return account{}, errors.New("Invalid username/password")
	}

	if acct.Password() != hashPassword(pass) {
		return account{}, errors.New("Invalid username/password")
	}

	if !acct.Active() {
		return account{}, errors.New("This account is not active")
	}
	log.Printf("Authenticated account %s, returning account %+v", username, acct)
	return acct, nil
}

//Register creates a new account with the specified information.
//TODO: Proper validation of values
func Register(username, password, email string) (Account, error) {
	acct, err := persistenceInstance.account(username)
	if err == nil {
		return nil, errors.New("Username already exists.")
	}
	if username == "" || len(username) < 3 || len(username) > 32 {
		return nil, errors.New("Invalid username")
	}
	if password == "" || len(password) < 5 {
		return nil, errors.New("Invalid password")
	}

	if email == "" || len(email) < 5 || !strings.Contains(email, "@") {
		return nil, errors.New("Invalid email address")
	}

	acct = newaccount(-1, username, hashPassword(password), email, true)
	err = persistenceInstance.saveAccount(&acct)
	if err != nil {
		return nil, err
	}
	log.Printf("Registered account: %+v", acct)
	return acct, nil
}

func newaccount(id int64, username, password, email string, active bool) account {
	return account{id: id, username: username, password: password, email: email, active: active}
}

type account struct {
	id       int64
	username string
	password string
	email    string
	active   bool
}

func (a account) ID() int64 {
	return a.id
}

//Returns the username of the account
func (a account) Username() string {
	return a.username
}

//Returns the password of the account
func (a account) Password() string {
	return a.password
}

//Returns the email of the account
func (a account) Email() string {
	return a.email
}

//Returns true if the account is active.
func (a account) Active() bool {
	return a.active
}

//TODO: Salt passwords
func hashPassword(password string) string {
	b := sha256.Sum256([]byte(password))
	return hex.EncodeToString(b[:32])
}
