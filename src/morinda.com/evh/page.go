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
	AppVer            string
	Config            Configuration
	DownloadUrlPath   string
	Expirations       map[int]Expiration
	HttpProto         string
	Message           template.HTML
	RequestHost       string
	StatusCode        int
	Total             float64
	Tracker           Tracker
	UploadUrlPath     string
	Year              int
	TrackerOfTrackers TrackerOfTrackers
}

func NewPage(r *http.Request) Page {
	var p = Page{
		Config:          Config,
		Year:            time.Now().Local().Year(),
		AppVer:          VERSION,
		HttpProto:       HttpProto,
		DownloadUrlPath: DownloadUrlPath,
		UploadUrlPath:   Config.Server.UrlPrefix + UploadUrlPath,
		StatusCode:      200,
	}

	// Get accurate server name/address (detects if we are being proxied)
	if _, ok := r.Header["X-Forwarded-Host"]; ok {
		p.RequestHost = r.Header["X-Forwarded-Host"][0]
	} else {
		p.RequestHost = r.Host
	}

	return p
}

// Render the template and send it to the client (or show 404)
func DisplayPage(w http.ResponseWriter, r *http.Request, tmpl string, data Page) {
	w.WriteHeader(data.StatusCode)
	if data.StatusCode == http.StatusNotFound {
		err := Templates.ExecuteTemplate(w, "404", data)
		if err != nil {
			log.Println(err.Error())
		}
	} else if SiteDown {
		var fpath = filepath.Join(Config.Server.Templates, "maintenance.html")
		http.ServeFile(w, r, fpath)
	} else {
		err := Templates.ExecuteTemplate(w, tmpl, data)
		if err != nil {
			log.Println(err.Error())
		}
	}
}
