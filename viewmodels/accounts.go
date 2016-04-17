package viewmodels

import (
	"errors"
	"strings"
)

type Account struct {
	Site
	Username        string
	Password        string
	ConfirmPassword string
	Email           string
	Errors          map[string]error
}

func GetAccount() Account {
	return Account{}
}

func (a *Account) Validate() map[string]error {
	errs := make(map[string]error)

	if len(a.Username) < 3 || len(a.Username) > 32 {
		errs["Username"] = errors.New("Invalid username. Must be between 3 and 32 characters.")
	}
	if len(a.Password) < 5 {
		errs["Password"] = errors.New("Invalid password. Must be longer than 5 characters")
	}

	if a.Password != a.ConfirmPassword {
		errs["ConfirmPassword"] = errors.New("Passwords do not match.")
	}

	if a.Email == "" || len(a.Email) < 5 || !strings.Contains(a.Email, "@") {
		errs["Email"] = errors.New("Invalid email address")
	}

	a.Errors = errs
	return errs
}
