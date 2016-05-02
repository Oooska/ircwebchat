package controllers

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/oooska/ircwebchat/chat"
)

type accountsController struct {
	template *template.Template //template for registration page
}

func (ac accountsController) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/register" {
		//Bring up the registration page
		ac.register(w, req)
	} else if req.URL.Path == "/login" && req.Method == "POST" {
		//Authenticate user, set cookie
		ac.login(w, req)
	} else if req.URL.Path == "/logout" && req.Method == "POST" {
		//Sign user out
		ac.logout(w, req)
	} else {
		w.WriteHeader(404)
	}
}

func (ac accountsController) login(w http.ResponseWriter, req *http.Request) {
	username := req.FormValue("Username")
	password := req.FormValue("Password")
	if username == "" || password == "" {
		w.Write([]byte("Invalid username/password."))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	acct, err := chat.Authenticate(username, password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	//Success, create session and send off to chat
	sessID, expires, err := chat.NewSession(acct)
	if err != nil {
		log.Printf("Error starting session: %s", err.Error())
	}
	setSessionCookie(w, sessID, expires)
	http.Redirect(w, req, "/chat", http.StatusFound)
}

func (ac accountsController) logout(w http.ResponseWriter, req *http.Request) {
	_, err := validateCookie(w, req)
	if err != nil {
		//Not logged in. No signing out to be done.
		http.Redirect(w, req, "/chat", http.StatusFound)
		return
	}
	sessID := sessionIDFromCookie(req)
	chat.DeleteSession(sessID)
	deleteSessionCookie(w)
	http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
}

func (ac accountsController) register(w http.ResponseWriter, req *http.Request) {
	_, err := validateCookie(w, req)
	if err == nil {
		//Account exists - no reason to be here...
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}
	account := viewaccount{}
	account.Title = "IRC Web Chat - Register"
	account.Active = "Register"

	if req.Method == "POST" {
		account.Username = req.FormValue("Username")
		account.Email = req.FormValue("Email")
		account.Password = req.FormValue("Password")
		account.ConfirmPassword = req.FormValue("ConfirmPassword")
		errs := account.Validate()

		if len(errs) == 0 {
			mdlAcct, err := chat.Register(account.Username, account.Password, account.Email)
			if err != nil {
				errs["Model"] = err
			} else {
				//Successfully register. Get this man (or woman) an auth token
				log.Printf("Successfully registered account: %+v", mdlAcct)
				sessID, expires, err := chat.NewSession(mdlAcct)
				log.Printf("Registered. Getting account %+v a session: %s", mdlAcct, sessID)
				if err != nil {
					log.Printf("Error starting session: %s", err.Error())
				}
				setSessionCookie(w, sessID, expires)
				http.Redirect(w, req, "/settings", http.StatusFound)
				return
			}
		}

	}

	ac.template.Execute(w, account)
}

//viewaccount contains the
type viewaccount struct {
	sitedata
	Username        string
	Password        string
	ConfirmPassword string
	Email           string
	Errors          map[string]error
}

func (a *viewaccount) Validate() map[string]error {
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

func setSessionCookie(w http.ResponseWriter, hash string, expires time.Time) {
	c := http.Cookie{Name: "SessionID", Value: hash, Expires: expires}
	http.SetCookie(w, &c)

}

func deleteSessionCookie(w http.ResponseWriter) {
	c := http.Cookie{Name: "SessionID", Value: "", Expires: time.Unix(0, 0)}
	http.SetCookie(w, &c)
}

func sessionIDFromCookie(req *http.Request) string {
	cookie, err := req.Cookie("SessionID")
	if err != nil {
		return ""
	}

	return cookie.Value
}

func validateCookie(w http.ResponseWriter, req *http.Request) (chat.Account, error) {
	cookie, err := req.Cookie("SessionID")
	if err != nil {
		return nil, err
	}

	sessID := cookie.Value
	acct, err := chat.LookupSession(sessID)
	if err != nil { //No account associated with this session, delete
		deleteSessionCookie(w)
		return nil, err
	}

	return acct, err
}
