package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func SpitSlurp() {
	var connStr = fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4,utf8&allowOldPasswords=1&parseTime=true", Config.Evh1.DbUser, Config.Evh1.DbPass, Config.Evh1.DbHost, Config.Evh1.DbName)

	if Config.Evh1.Timezone != "" {
		connStr = connStr + "&loc=" + url.QueryEscape(Config.Evh1.Timezone)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Println("ERROR: sql.Open:", err.Error())
		return
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		log.Println("ERROR: db.Ping:", err.Error())
		return
	}

	// Prepare statement for inserting data
	rows, err := db.Query("SELECT id, sessionid, name from Files where 1 order by id")
	if err != nil {
		log.Println("ERROR: query error:", err.Error())
		return
	}

	// Loop over files in EVH
	for rows.Next() {
		var id int
		var sessionid, name string

		err := rows.Scan(&id, &sessionid, &name)
		if err != nil {
			log.Println("ERROR: row error:", err.Error())
			return
		}

		// Fetch the rest of the information for the session
		var fileid, size int
		var avail, dnldcode, modcode, srcemail, destemail, description string
		var indate, outdate time.Time
		sessinfo, sesserr := db.Prepare("select Sessions.id, dnldcode, modcode, indate, outdate, avail, srcemail, destemail, size, description, Files.id from Sessions,Files where Sessions.id = ? and Files.sessionid=Sessions.id")
		if sesserr != nil {
			log.Println("ERROR: query error:", sesserr.Error())
			return
		}
		defer sessinfo.Close()

		sessinfoerr := sessinfo.QueryRow(sessionid).Scan(&sessionid, &dnldcode, &modcode, &indate, &outdate, &avail, &srcemail, &destemail, &size, &description, &fileid)
		if sessinfoerr != nil {
			log.Println("ERROR: row error:", sessinfoerr.Error())
			return
		}

		// Does this already exist (skip if it does)
		var trackerfile = filepath.Join(Config.Server.Assets, dnldcode, TrackerFileName)
		tracker, loaderr := LoadTrackerFromFile(trackerfile)
		if loaderr != nil {
			log.Println("Found new file to migrate:", dnldcode)
		} else {
			log.Println("Skipping file, already migrated: ", dnldcode)
			continue
		}

		// New Tracker, lets get this file migrated!
		tracker = NewTracker(dnldcode)
		tracker.Description = description
		tracker.Vercode = dnldcode
		tracker.When = indate
		tracker.ExpirationDate = outdate
		tracker.Expiration = avail
		tracker.ExpirationStr = tracker.ExpirationDate.Format(TimeLayout)
		tracker.SrcEmail = srcemail
		tracker.DstEmail = destemail
		tracker.Size = float64(size)
		tracker.SizeMB = tracker.Size / 1024 / 1024

		var file = NewFile(name, tracker.BaseDir)
		file.Size = tracker.Size
		file.SizeMB = tracker.SizeMB
		file.Saved = true
		file.WhenSaved = tracker.When.Format(TimeLayout)

		tracker.Files = append(tracker.Files, file)

		// Create new asset dir
		mkdirerr := os.MkdirAll(tracker.BaseDir, 0700)
		if mkdirerr != nil {
			log.Println(dnldcode, "ERROR: Could not create tracker dir:", mkdirerr.Error())
			return
		}

		// Write tracker
		saveerr := tracker.Save()
		if saveerr != nil {
			log.Println(dnldcode, "ERROR: Could not save tracker:", saveerr.Error())
			return
		}

		// Copy (or move) file to new location
		var oldfile = filepath.Join(Config.Evh1.Assets, tracker.Dnldcode, name)
		var newfile = filepath.Join(Config.Server.Assets, tracker.Dnldcode, file.Base64Name)

		// Move
		if Config.Evh1.Move {
			moveerr := os.Rename(oldfile, newfile)
			if moveerr != nil {
				log.Println(dnldcode, "ERROR: Could not move file:", moveerr.Error())
				RemoveDir(tracker.BaseDir)
				continue
			} else {
				log.Println(dnldcode, "Move successful")
			}
		} else { // Copy
			// Create destination file
			dst, dsterr := os.Create(newfile)
			if dsterr != nil {
				log.Println(fmt.Sprintf("%s ERROR: could create empty file: %s", dnldcode, dsterr.Error()))
				RemoveDir(tracker.BaseDir)
				continue
			}

			// Open old file for copying
			fhandle, tmpferr := os.Open(oldfile)
			if tmpferr != nil {
				log.Println(fmt.Sprintf("%s ERROR: could not open old file for copy: %s", dnldcode, tmpferr.Error()))
				RemoveDir(tracker.BaseDir)
				continue
			}

			// Copy the uploaded file to the destination file
			bytes, cpyerr := io.Copy(dst, fhandle)
			if cpyerr != nil {
				log.Println(fmt.Sprintf("%s ERROR: could not write file contents: %s", dnldcode, cpyerr.Error()))
				RemoveDir(tracker.BaseDir)
				continue
			}
			fhandle.Close()
			dst.Close()

			log.Println(fmt.Sprintf("%s Copied %.2f MB", dnldcode, float64(bytes)/1024/1024))
		}
	}
}
