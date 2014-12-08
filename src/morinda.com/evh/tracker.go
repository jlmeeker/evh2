// The tracker is a json file that contains all info about a download
// session and the files it contains.  
//
// This data is persistent and stored on disk in an info.json file.
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

var TrackerFileName = "info.json"

type Tracker struct {
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
		log.Println(err.Error())
		trackererr = err
		return
	}

	// convert from json to tracker object
	err = json.Unmarshal(rawdata, &tracker)
	if err != nil {
		log.Println(err.Error())
		trackererr = err
		return
	}

	return
}

// Write tracker to file
func (t *Tracker) Save(basedir string) error {
	var fpath = filepath.Join(basedir, TrackerFileName)

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
