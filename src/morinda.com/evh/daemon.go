// The daemon is a set of go routines that run in endless loops
// and perform necessary back-end functions.  These are spawned
// one by one in main.go.
package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

// Parse HTML templates on a set interval.
// NOTE: Errors in templates will result in a "maintenance" notice
// on the site and the error below in the logs.
func RefreshTemplates() {
	var tmplGlob = filepath.Join(Config.Server.Templates, "*.html")
	var err error

	for {
		Templates, err = template.ParseGlob(tmplGlob)
		if err != nil {
			log.Println("Error parsing HTML templates!!! Site is down.")
			SiteDown = true
		} else {
			SiteDown = false
		}

		// Wait and refresh templates again
		time.Sleep(time.Minute * time.Duration(Config.Daemon.TmplFreq))
	}

	return
}

// Clean up expired uploads on a set interval
// NOTE: The occasional error may show up in logs if this runs and sees a
// directory that doesn't have an info.json file in it yet.
func ScrubDownloads() {
	for {
		entries, err := ioutil.ReadDir(Config.Server.Assets)
		if err != nil {
			log.Println("ERROR: Scrub cannot read assets directory:", err.Error())
		} else {
			// Loop over each file
			for _, entry := range entries {
				// Skip NoPurgeCheck directories
				var skipDir bool
				for _, npcdir := range Config.Daemon.NoPurgeCheck {
					if entry.Name() == npcdir {
						skipDir = true
						break
					}
				}

				if skipDir {
					continue
				}

				if entry.IsDir() {
					// See if the info.json file exists
					var infofile = filepath.Join(Config.Server.Assets, entry.Name(), "info.json")

					// Read in the info.json file
					tracker, trackererr := LoadTrackerFromFile(infofile)
					if trackererr != nil || tracker.When == (time.Time{}) {
						log.Println("WARNING: Scrub error parsing info file")
						continue
					}

					// Check tracker.ExpirationDate against current time
					//    Delete download dir if expired
					if tracker.ExpirationDate.Before(time.Now().Local()) {
						var dirpath = filepath.Join(Config.Server.Assets, entry.Name())
						rderr := RemoveDir(dirpath)
						if rderr != nil {
							log.Println(tracker.Dnldcode, "Could not remove expired files:", rderr.Error())
							continue
						}

						log.Println(tracker.Dnldcode, "Files expired and purged")
					}
				}
			}
		}

		// Wait before doing it all over again
		time.Sleep(time.Minute * time.Duration(Config.Daemon.PurgeFreq))
	}
}
