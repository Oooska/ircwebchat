package models

import (
	"errors"
	"strings"
)

func NewAccounts() Accounts {
	accts := accounts{acctmap: make(map[string]Account)}

	//dummy data
	accts.Register("goirctest", "password", "a@b.c")

	//accts.Register("goirctest2", "password", "a@b.c")
	return accts
}

//Accounts maintains a list of accounts on the server.
type Accounts interface {
	Account(username string) (Account, error)
	Authenticate(username, pass string) (Account, error)
	Register(username, password, email string) (Account, error)
	accountMap() map[string]Account
}

type accounts struct {
	acctmap map[string]Account
}

func (accts accounts) accountMap() map[string]Account {
	return accts.acctmap
}

func (accs accounts) Account(username string) (Account, error) {
	acct, ok := accs.acctmap[username]
	if !ok {
		return acct, errors.New("Account does not exist")
	}
	if !acct.Active() {
		return account{}, errors.New("This account is no longer active.")
	}
	return acct, nil
}

func (accs accounts) Authenticate(username, pass string) (Account, error) {
	acct, ok := accs.acctmap[username]
	if !ok {
		return acct, errors.New("Invalid username/password")
	}
	if acct.Password() != pass {
		return account{}, errors.New("Invalid username/password")
	}

	if !acct.Active() {
		return account{}, errors.New("This account is no longer active")
	}
	return acct, nil
}

//TODO: Proper validation of values
func (accs accounts) Register(username, password, email string) (Account, error) {
	acct, ok := accs.acctmap[username]
	if ok {
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

	acct = account{username: username, password: password, email: email, active: true}
	accs.acctmap[username] = acct
	return acct, nil
}

//Account represents a user account in the system
type Account interface {
	Username() string
	Password() string
	Email() string
	Active() bool
}

type account struct {
	username string
	password string
	email    string
	active   bool
}

func (a account) Username() string {
	return a.username
}

func (a account) Password() string {
	return a.password
}

func (a account) Email() string {
	return a.email
}

func (a account) Active() bool {
	return a.active
}
