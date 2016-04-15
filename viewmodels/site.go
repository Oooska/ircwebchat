package viewmodels

//Site contains information pertinent to all pages on the site
//Username will be "" if user is not signed in
type Site struct {
	Title    string
	Username string
}

func GetSite() Site {
	return Site{Username: "Jorge"}
}
