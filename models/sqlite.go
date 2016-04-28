package models

import (
	"database/sql"
	"log"
)

type sqlite3 struct {
	db *sql.DB
}

func (p *sqlite3) Start(filename string) error {
	var err error
	p.db, err = sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}
	return nil
}

func (p *sqlite3) Stop() error {
	return p.db.Close()
}

func (p *sqlite3) Init() error {
	for _, sqlStmt := range createTables {
		_, err := p.db.Exec(sqlStmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *sqlite3) Account(username string) (Account, error) {
	var acct Account
	log.Printf("Account(%s)... p.db: %+v", username, p)
	stmt, err := p.db.Prepare(`SELECT accountid, name, password, email, active ` +
		`FROM accounts WHERE name = ?`)
	log.Printf("Err: %+v", err)

	if err != nil {
		return acct, err
	}
	defer stmt.Close()
	log.Printf("Getting row...")
	row := stmt.QueryRow(username)
	var accountid int
	var name, password, email string
	var active bool
	log.Printf("Scanning row...")
	err = row.Scan(&accountid, &name, &password, &email, &active)
	if err != nil {
		return acct, err
	}
	log.Printf("Retrived account details: %d | %s | %s | %s | %v", accountid, name, password, email, active)
	return newaccount(accountid, name, password, email, active), nil
}

func (p *sqlite3) SaveAccount(acct Account) error {
	log.Printf("Saving account: %+v", acct)
	stmt, err := p.db.Prepare(`INSERT INTO accounts(name, password, email, active) ` +
		`VALUES(?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(acct.Username(), acct.Password(), acct.Email(), true)
	return err
}

//Statements for creating neccesary tables
//TODO: Move statements to their own file
var createTables = []string{`create table if not exists accounts (
		accountid INTEGER not null primary key, 
		name      TEXT,
		password  TEXT,
		email     TEXT,
        active    INTEGER
	);`,
	`create table if not exists sessions (
		account    INTEGER, 
		sessionkey TEXT,
		expires    TIMESTAMP,
		FOREIGN KEY (account) REFERENCES accounts(accountid)
	);`,
	`create table if not exists ircsettings (
		ircaccountid     INTEGER not null primary key,
		active           INTEGER,
		name             TEXT,
		server           TEXT,
		port             INTEGER,
		ssl              INTEGER,
		nick             TEXT,
		pass             TEXT,
		altnick          TEXT,
		altpass          TEXT,
		account          INTEGER,
		FOREIGN KEY (account) REFERENCES accounts(accountid)
	);`,
	`create table if not exists channels (
		ircaccount  INTEGER,
		channel     TEXT
	);`,
	`create table if not exists messages (
		messageid    INTEGER not null primary key,
		ircaccount   INTEGER,
		timestamp    TIMESTAMP,
		command      TEXT,
		channel      TEXT,
		message      TEXT,
		FOREIGN KEY (ircaccount) REFERENCES ircaccounts(ircaccountid)
	);`,
}
