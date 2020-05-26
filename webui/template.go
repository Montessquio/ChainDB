package webui

import (
	"path"
	"html/template"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

// Cached templates are loaded on serve.
var templates map[string]*template.Template

func initTemplates() {
	files, err := ioutil.ReadDir("www/template")
	if err != nil {
		log.WithField("Error", err).Fatal("Error reading from template directory `www/template`.")
	}

	templates = make(map[string]*template.Template)
	for _, file := range files {
		if !file.IsDir() {
			content, err := ioutil.ReadFile(path.Join("www", "template", file.Name()))
			if err != nil {
				log.WithFields(log.Fields{
					"Path": path.Clean(path.Join("www", "template", file.Name())),
					"Error": err,
				}).Fatal("Error reading template file!")
			}
			template, err := template.New(file.Name()).Parse(string(content))
			if err != nil {
				log.WithFields(log.Fields{
					"Path": path.Clean(path.Join("www", "template", file.Name())),
					"Error": err,
				}).Fatal("Error parsing Template!")
			}
			templates[file.Name()] = template
		}
	}

	log.Infof("Registered %d templates.", len(templates))
}