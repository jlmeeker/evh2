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
	AppVer          string
	BaseUrl         string
	Config          Configuration
	DownloadUrlPath string
	Expirations     map[int]Expiration
	HttpProto       string
	Message         template.HTML
	Total           float64
	Tracker         Tracker
	UploadUrlPath   string
	Year            int
}

func NewPage() Page {
	// Set the proper protocol and, if necessary, the port too
	var url = HttpProto + "://" + Config.Server.Address
	if Config.Server.Ssl && Config.Server.SslPort != "443" {
		url = url + ":" + Config.Server.SslPort
	} else if Config.Server.Ssl == false && Config.Server.NonSslPort != "80" && Config.Server.UrlPrefix == "" {
		url = url + ":" + Config.Server.NonSslPort
	}

	return Page{
		Config:          Config,
		Year:            time.Now().Local().Year(),
		AppVer:          VERSION,
		HttpProto:       HttpProto,
		BaseUrl:         url,
		DownloadUrlPath: DownloadUrlPath,
		UploadUrlPath:   Config.Server.UrlPrefix + UploadUrlPath,
	}
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
