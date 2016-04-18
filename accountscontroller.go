package ircwebchat

import (
	"html/template"
	"net/http"
	"time"

	"github.com/oooska/ircwebchat/models"
	"github.com/oooska/ircwebchat/viewmodels"
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
	acct, err := modelAccounts.Authenticate(username, password)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Success, create session and send off to chat
	sessID, expires := modelSessions.Start(acct)
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
	modelSessions.Delete(sessID)
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
	account := viewmodels.GetAccount()
	account.Title = "IRC Web Chat - Register"

	if req.Method == "POST" {
		account.Username = req.FormValue("Username")
		account.Email = req.FormValue("Email")
		account.Password = req.FormValue("Password")
		account.ConfirmPassword = req.FormValue("ConfirmPassword")
		errs := account.Validate()

		if len(errs) == 0 {
			mdlAcct, err := modelAccounts.Register(account.Username, account.Password, account.Email)
			if err != nil {
				errs["Model"] = err
			} else {
				//Successfully register. Get this man (or woman) an auth token
				sessID, expires := modelSessions.Start(mdlAcct)
				setSessionCookie(w, sessID, expires)
				http.Redirect(w, req, "/settings", http.StatusFound)
				return
			}
		}

	}

	ac.template.Execute(w, account)
}

//TODO: Set expiration on cookies
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

func validateCookie(w http.ResponseWriter, req *http.Request) (models.Account, error) {
	cookie, err := req.Cookie("SessionID")
	if err != nil {
		return nil, err
	}

	sessID := cookie.Value

	acct, err := modelSessions.Lookup(sessID)
	if err != nil { //No account associated with this session, delete
		deleteSessionCookie(w)
	}
	return acct, err
}
