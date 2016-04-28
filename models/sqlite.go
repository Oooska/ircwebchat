package models

import (
	"database/sql"
	"log"
	"time"
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

func (p *sqlite3) account(username string) (account, error) {
	var acct account
	log.Printf("Account(%s)... p.db: %+v", username, p)
	stmt, err := p.db.Prepare(`SELECT accountid, username, password, email, active ` +
		`FROM accounts WHERE username = ?`)
	log.Printf("Err: %+v", err)

	if err != nil {
		return acct, err
	}
	defer stmt.Close()
	log.Printf("Getting row...")
	row := stmt.QueryRow(username)
	var accountid int64
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

func (p *sqlite3) saveAccount(acct *account) error {
	log.Printf("Saving account: %+v", acct)
	stmt, err := p.db.Prepare(`INSERT INTO accounts(username, password, email, active) ` +
		`VALUES(?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(acct.Username(), acct.Password(), acct.Email(), true)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	acct.id = id
	stmt.Close()
	return err
}

func (p *sqlite3) session(id string) (session, error) {
	var sess session
	var acct account
	stmt, err := p.db.Prepare(`SELECT accountid, username, password, email, active, sessionid, expires ` +
		`FROM accounts, sessions ON accountid=account WHERE sessionid = ?`)
	if err != nil {
		return sess, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	var accountid int64
	var name, password, email, sessionID string
	var active bool
	var expires time.Time
	err = row.Scan(&accountid, &name, &password, &email, &active, &sessionID, &expires)
	if err != nil {
		return sess, err
	}
	acct = newaccount(accountid, name, password, email, active)
	sess = newsession(sessionID, acct, expires)

	log.Printf("Requested session for ID %s, retrieved: %+v", id, sess)
	return sess, err
}

func (p *sqlite3) saveSession(s session) error {
	stmt, err := p.db.Prepare(`INSERT INTO sessions(account, sessionid, expires) ` +
		`VALUES(?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(s.account.ID(), s.id, s.expires)
	stmt.Close()
	log.Printf("Saving session: %+v", s)
	return err
}

func (p *sqlite3) deleteSession(id string) error {
	stmt, err := p.db.Prepare(`DELETE FROM sessions WHERE sessionid = ?`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id)
	stmt.Close()
	log.Printf("Deleting session %s", id)
	return err
}

//Statements for creating neccesary tables
//TODO: Move statements to their own file
var createTables = []string{`create table if not exists accounts (
		accountid INTEGER not null primary key, 
		username  TEXT,
		password  TEXT,
		email     TEXT,
        active    INTEGER
	);`,
	`create table if not exists sessions (
		account    INTEGER, 
		sessionid  TEXT,
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
