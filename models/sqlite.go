package models

import (
	"database/sql"
	"log"
	"time"

	"github.com/oooska/irc"
)

type sqlite3 struct {
	db        *sql.DB
	secretkey string
}

func (p *sqlite3) Start(filename string) error {
	var err error
	p.db, err = sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}

	for _, sqlStmt := range sqlLiteTables {
		_, err := p.db.Exec(sqlStmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *sqlite3) Stop() error {
	return p.db.Close()
}

func (p *sqlite3) account(username string) (account, error) {
	var acct account
	stmt, err := p.db.Prepare(`SELECT accountid, username, password, email, active ` +
		`FROM accounts WHERE username = ?`)

	if err != nil {
		return acct, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(username)
	var accountid int64
	var name, password, email string
	var active bool
	err = row.Scan(&accountid, &name, &password, &email, &active)
	if err != nil {
		return acct, err
	}
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

//activeAccounts obtains all active accounts from the database
func (p *sqlite3) activeAccounts() ([]account, error) {
	var accts []account
	stmt, err := p.db.Prepare(`SELECT accountid, username, password, email, active ` +
		`FROM accounts WHERE active = 1`)
	if err != nil {
		return accts, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return accts, err
	}
	defer rows.Close()
	for rows.Next() {
		var accountid int64
		var name, password, email string
		var active bool
		err = rows.Scan(&accountid, &name, &password, &email, &active)
		if err != nil {
			return accts, err
		}
		accts = append(accts, newaccount(accountid, name, password, email, active))
	}
	return accts, nil
}

func (p *sqlite3) session(id string) (session, error) {
	var sess session
	var acct account
	stmt, err := p.db.Prepare(`SELECT accounts.accountid, username, password, email, active, sessionid, expires ` +
		`FROM accounts, sessions ON accounts.accountid=sessions.accountid WHERE sessionid = ?`)
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
	stmt, err := p.db.Prepare(`INSERT INTO sessions(accountid, sessionid, expires) ` +
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

func (p *sqlite3) settings(acct Account) (Settings, error) {
	var settings Settings
	stmt, err := p.db.Prepare(`SELECT accountid, active, name, server, port, ssl, nick, pass, altnick, altpass ` +
		`FROM ircsettings WHERE accountid=?`)
	if err != nil {
		return settings, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(acct.ID())
	var accountid int64
	var active, ssl bool
	var name, server, nick, password, altnick, altpassword string
	var port int
	err = row.Scan(&accountid, &active, &name, &server, &port, &ssl, &nick, &password, &altnick, &altpassword)
	if err != nil {
		return settings, err
	}
	settings = newsettings(accountid, active, name, server, port, ssl, newirclogin(nick, password), newirclogin(altnick, altpassword))

	//Need to decrypt password
	login, err := decryptPassword(p.secretkey, settings.Login)
	if err != nil {
		log.Printf("Error decrypting IRC password: %s", err.Error())
	}
	settings.Login = login
	altlogin, err := decryptPassword(p.secretkey, settings.AltLogin)
	if err != nil {
		log.Printf("Error decrypting alt IRC password: %s", err.Error())
	}
	settings.AltLogin = altlogin
	return settings, nil
}

func (p *sqlite3) saveSettings(s Settings) error {
	//TODO: Need to encrypt irc passwords

	stmt, err := p.db.Prepare(`REPLACE INTO ` +
		`ircsettings (accountid, active, name, server, port, ssl, nick, pass, altnick, altpass) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(s.accountid, s.Enabled,
		s.Name, s.Address, s.Port, s.SSL, s.Login.Nick,
		s.Login.Password, s.AltLogin.Nick, s.AltLogin.Password)
	return err
}

func (p *sqlite3) messages(acct Account, channel string, timestamp time.Time, count int) ([]irc.Message, error) {
	var messages []irc.Message
	stmt, err := p.db.Prepare(`SELECT message, timestamp FROM messages WHERE accountid = ? AND channel = ? AND timestamp < ? ORDER BY timestamp LIMIT ?`)
	if err != nil {
		return messages, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(acct.ID(), channel, timestamp, count)
	if err != nil {
		return messages, err
	}
	defer rows.Close()

	for rows.Next() {
		var message string
		var timestamp time.Time
		err = rows.Scan(&message, &timestamp)
		if err != nil {
			return messages, err
		}
		messages = append(messages, irc.MessageWithTimestamp(message, timestamp))
	}
	return messages, nil
}

func (p *sqlite3) saveMessage(acct Account, msg irc.Message) error {
	stmt, err := p.db.Prepare(`INSERT INTO messages(accountid, command, channel, message, timestamp) ` +
		`VALUES(?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	channel := ""
	if msg.Command() == "PRIVMSG" || msg.Command() == "ACTION" && len(msg.Params()) > 1 {
		channel = msg.Params()[0]
	}
	_, err = stmt.Exec(acct.ID(), msg.Command(), channel, msg.Message(), msg.Timestamp())
	return err
}

//Statements for creating neccesary tables
var sqlLiteTables = []string{`create table if not exists accounts (
		accountid INTEGER not null primary key, 
		username  TEXT UNIQUE,
		password  TEXT,
		email     TEXT,
        active    INTEGER
	);`,
	`create table if not exists sessions (
		accountid  INTEGER, 
		sessionid  TEXT,
		expires    TIMESTAMP,
		FOREIGN KEY (accountid) REFERENCES accounts(accountid)
	);`,
	`create table if not exists ircsettings (
		accountid        INTEGER not null primary key,
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
		FOREIGN KEY (accountid) REFERENCES accounts(accountid)
	);`,
	`create table if not exists channels (
		accountid  INTEGER,
		channel     TEXT
	);`,
	`create table if not exists messages (
		messageid    INTEGER not null primary key,
		accountid    INTEGER,
		command      TEXT,
		channel      TEXT,
		message      TEXT,
		timestamp    TIMESTAMP,
		FOREIGN KEY (accountid) REFERENCES ircaccounts(accountid)
	);`,
}
