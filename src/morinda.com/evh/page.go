// The Page object facilitates transfer of information into the HTML
// templating engine.
//
// NOTE: As html templates expand, this object should house the
// necessary data objects.
package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type Page struct {
	AppVer      string
	Config      Configuration
	Expirations map[string]string
	HttpProto   string
	Message     template.HTML
	Total       float64
	Tracker     Tracker
	Year        int
}

func NewPage() Page {
	return Page{Config: Config, Year: time.Now().Local().Year(), AppVer: VERSION, HttpProto: HttpProto}
}

// Render the template and send it to the client
func DisplayPage(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	if SiteDown {
		var fpath = filepath.Join(Config.Server.Templates, "maintenance.html")
		http.ServeFile(w, r, fpath)
	} else {
		err := Templates.ExecuteTemplate(w, tmpl, data)
		if err != nil {
			log.Println(err.Error())
		}
	}
}
