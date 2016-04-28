package models

import (
	"errors"
	"log"
	"strings"
)

//NewAccounts returns an Accounts
func NewAccounts() Accounts {
	accts := accounts{}

	//dummy data
	//accts.Register("goirctest", "password", "a@b.c")

	//accts.Register("goirctest2", "password", "a@b.c")
	return accts
}

//Accounts maintains a list of accounts on the server.
type Accounts interface {
	Account(username string) (Account, error)
	Authenticate(username, pass string) (Account, error)
	Register(username, password, email string) (Account, error)
}

type accounts struct {
	//acctmap map[string]Account
}

//Account returns the account with the specified username, or an error if none is found
func (accs accounts) Account(username string) (Account, error) {
	acct, err := persistenceInstance.Account(username)
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
func (accs accounts) Authenticate(username, pass string) (Account, error) {
	acct, err := persistenceInstance.Account(username)
	log.Printf("Recieved error while authenticating: %v", err)
	if err != nil {
		return account{}, errors.New("Invalid username/password")
	}

	if acct.Password() != pass {
		return account{}, errors.New("Invalid username/password")
	}

	if !acct.Active() {
		return account{}, errors.New("This account is not active")
	}
	return acct, nil
}

//Register creates a new account with the specified information.
//TODO: Proper validation of values
func (accs accounts) Register(username, password, email string) (Account, error) {
	acct, err := persistenceInstance.Account(username)
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

	acct = newaccount(-1, username, password, email, true)
	err = persistenceInstance.SaveAccount(acct)
	if err != nil {
		return nil, err
	}
	return acct, nil
}

//Account represents a user account in the system
type Account interface {
	Username() string
	Password() string
	Email() string
	Active() bool
}

func newaccount(id int, username, password, email string, active bool) account {
	return account{id: id, username: username, password: password, email: email, active: active}
}

type account struct {
	id       int
	username string
	password string
	email    string
	active   bool
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
