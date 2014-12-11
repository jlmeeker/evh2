// The tracker is a json file that contains all info about a download
// session and the files it contains.
//
// This data is persistent and stored on disk in an info.json file.
//
// The TrackerOfTrackers object is used for viewing all of the available
// uploads currently in the system.  This is currently only used by
// the /admin/ page.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"time"
)

type TrackerOfTrackers struct {
	ScanStart     time.Time
	ScanStop      time.Time
	Trackers      map[int64]Tracker
	TotalFiles    int
	TotalSessions int
	TotalSize     float64
	TotalSizeMB   float64
	TotalSizeGB   float64
}

var TrackerFileName = "info.json"

type Tracker struct {
	BaseDir        string
	CliUpload      bool
	Description    string
	Dnldcode       string
	DstEmail       string
	Expiration     string
	ExpirationDate time.Time
	ExpirationStr  string
	Files          []File
	Log            map[string]string
	SrcEmail       string
	Size           float64
	SizeMB         float64
	Vercode        string
	When           time.Time
}

// Generate new tracker
func NewTracker(dnldcode string) (tracker Tracker) {
	tracker.BaseDir = filepath.Join(Config.Server.Assets, dnldcode)
	tracker.Dnldcode = dnldcode
	tracker.Log = make(map[string]string)
	tracker.Vercode, _ = GenCode(false, 0)
	tracker.When = time.Now().Local()
	return
}

// Read tracker from file
func LoadTracker(basedir, dnldcode string) (tracker Tracker, err error) {
	var fpath = filepath.Join(basedir, dnldcode, TrackerFileName)
	tracker, err = LoadTrackerFromFile(fpath)

	return
}

// Read and parse json file
func LoadTrackerFromFile(fpath string) (tracker Tracker, trackererr error) {
	// Read in file contents
	rawdata, err := ioutil.ReadFile(fpath)
	if err != nil {
		trackererr = err
		return
	}

	// convert from json to tracker object
	err = json.Unmarshal(rawdata, &tracker)
	if err != nil {
		trackererr = err
		return
	}

	return
}

// Write tracker to file
func (t *Tracker) Save() error {
	var fpath = filepath.Join(t.BaseDir, TrackerFileName)

	filedata, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		log.Println("error:", err.Error())
		return err
	}

	err = ioutil.WriteFile(fpath, filedata, 0600)
	if err != nil {
		log.Println("error:", err.Error())
		return err
	}

	return nil
}

// Add a log entry
func (t *Tracker) AddLog(msg string) {
	var timeStr = strconv.FormatInt(time.Now().Local().UnixNano(), 10)
	t.Log[timeStr] = msg
}

// Dump tracker data to console
func (t *Tracker) Dump() {
	fmt.Printf("%#v\n", t)
}

// Save and report the number of successful file saves
func (t *Tracker) CountSaved() int {
	var numSaved int
	for _, file := range t.Files {
		if file.Saved {
			numSaved++
		}
	}

	return numSaved
}

// Scans all available downloads and compiles the info (this is for admin view)
func NewTrackerOfTrackers() TrackerOfTrackers {
	// Our result object
	var toft = TrackerOfTrackers{}
	toft.Trackers = make(map[int64]Tracker)
	toft.ScanStart = time.Now().Local()

	fileinfos, err := ioutil.ReadDir(Config.Server.Assets)
	if err != nil {
		log.Println("ERROR: TofT is unable to read asset dir: ", err.Error())
		toft.ScanStop = time.Now().Local()
		return toft
	}

	// These should all be download dirs, but we'll check just in case
	for _, info := range fileinfos {
		if info.IsDir() {
			var trackerfpath = filepath.Join(Config.Server.Assets, info.Name(), TrackerFileName)
			tracker, trerr := LoadTrackerFromFile(trackerfpath)
			if trerr == nil {
				toft.TotalSessions++
				toft.TotalSize += tracker.Size
				toft.TotalSizeMB = toft.TotalSize / 1024 / 1024
				toft.TotalSizeGB = toft.TotalSize / 1024 / 1024 / 1024
				toft.TotalFiles += len(tracker.Files)
				toft.Trackers[tracker.ExpirationDate.UnixNano()] = tracker
			} else {
				log.Println("ERROR: ", trerr.Error())
			}
		}
	}

	toft.ScanStop = time.Now().Local()
	return toft
}
