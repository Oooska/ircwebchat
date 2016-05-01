package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"github.com/oooska/irc"
)

var persistenceInstance Persistence

func NewPersistenceInstance(driver string, key string) (Persistence, error) {
	var p Persistence
	if driver == "sqlite3" {
		persistenceInstance = &sqlite3{secretkey: key}
		return persistenceInstance, nil
	}

	//TODO: Support an in-memory persistence object
	return p, errors.New("SQL Driver not supported")
}

type Persistence interface {
	Start(filename string) error //Opens db and connects to it
	Stop() error                 //Closes db
	Init() error                 //Creates tables

	PersistentAccounts
	PersistentSession
	PersistentSettings
	PersistentMessages
}

type PersistentAccounts interface {
	account(username string) (account, error)
	saveAccount(acct *account) error
	activeAccounts() ([]account, error)
}

type PersistentSession interface {
	session(id string) (session, error)
	saveSession(s session) error
	deleteSession(id string) error
}

type PersistentSettings interface {
	settings(account Account) (Settings, error)
	saveSettings(s Settings) error
}

type PersistentMessages interface {
	messages(acct Account, channel string, timestamp time.Time, count int) ([]irc.Message, error)
	saveMessage(acct Account, msg irc.Message) error
}

//encryptPassword encrypts the password for the supplied IRCLogin using the secret key.
//If password is blank, no encryption occurs.
//An error is only returned if encryption is attempted and fails.
func encryptPassword(key string, login IRCLogin) (IRCLogin, error) {
	if login.Password == "" {
		return login, nil
	}
	encPass, err := encrypt([]byte(key), []byte(login.Password))
	if err != nil {
		login.Password = string(encPass)
	}
	return login, err
}

//decryptPassword decrypts the password for the supplied IRC login using the secret key.
//If password is blank, no encryption occurs.
//An error is only returned if decryption is attempted and fails
func decryptPassword(key string, login IRCLogin) (IRCLogin, error) {
	if login.Password == "" {
		return login, nil
	}
	decPass, err := decrypt([]byte(key), []byte(login.Password))
	if err != nil {
		login.Password = string(decPass)
	}
	return login, err
}

//Encrypt irc user passwords
//Source: http://stackoverflow.com/questions/18817336/golang-encrypting-a-string-with-aes-and-base64
func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

//Decrypt irc user passwords
//Source: http://stackoverflow.com/questions/18817336/golang-encrypting-a-string-with-aes-and-base64
func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}
