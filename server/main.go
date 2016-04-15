package main

import (
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/oooska/ircwebchat"
)

//Starts a basic http server with the ircwebchat Handler registered
func main() {
	//mux := http.NewServeMux()
	t := populateTemplates()
	ircwebchat.Register(t, nil) //mux)
	go log.Fatal(http.ListenAndServe(":8080", nil))
}

func populateTemplates() *template.Template {
	result := template.New("templates")
	basePath := "templates"
	templatePaths := parseTemplateDirectory(basePath)
	result, err := result.ParseFiles(templatePaths...)
	if err != nil {
		log.Fatalf("Error parsing templates: %s", err.Error())
	}
	return result
}

func parseTemplateDirectory(basePath string) []string {
	templateFolder, err := os.Open(basePath)
	defer templateFolder.Close()
	if err != nil {
		log.Fatalf("Unable to open templates folder %s.", basePath)
	}

	templatePathsRaw, _ := templateFolder.Readdir(-1)

	templatePaths := new([]string)
	for _, pathInfo := range templatePathsRaw {
		if !pathInfo.IsDir() {
			*templatePaths = append(*templatePaths, basePath+"/"+pathInfo.Name())
			log.Printf("Adding %s to list of templates", basePath+"/"+pathInfo.Name())
		} else {
			subtemplatePaths := parseTemplateDirectory(basePath + "/" + pathInfo.Name())
			*templatePaths = append(*templatePaths, subtemplatePaths...)
		}
	}

	return *templatePaths

}
