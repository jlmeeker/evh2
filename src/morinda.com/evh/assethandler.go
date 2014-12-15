// assetHandler controls all request for existing (already uploaded)
// file data.  This either a listing of files in an upload session
// or the contents of a specific file.
package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func AssetHandler(w http.ResponseWriter, r *http.Request) {
	// Get initial client address
	var requestAddr = r.RemoteAddr

	// Get accurate client address
	if val, ok := r.Header["X-Forwarded-For"]; ok {
		requestAddr = strings.Join(val, ",")
	}

	var dnldcode string
	var isfiledownload bool

	// Get a new Page object
	var page = NewPage(r)

	// Get our GET request variable(s)
	var vercode = r.URL.Query().Get("vercode")

	// All of path AFTER DownloadUrlPath
	reqpath := strings.TrimPrefix(r.URL.Path, DownloadUrlPath)

	// Do not build a file list for root downloads dir
	if reqpath == "" {
		page.Message = "File(s) not found"
		DisplayPage(w, r, "files", page)
		return
	}

	// Split up the requested path for further processing
	reqpathdir, reqpathfile := filepath.Split(reqpath)

	// If a dir is specified then one of reqpathdir or reqpathfile will be empty
	//   dnldcode ends up being set to whatever isn't empty (minus any trailing slash)
	if reqpathdir == "" || reqpathfile == "" {
		dnldcode = reqpathdir + reqpathfile
	} else {
		dnldcode = reqpathdir
		isfiledownload = true
	}
	dnldcode = strings.TrimRight(dnldcode, "/")

	// Make sure Dnldcode dir exists
	var ddpath = filepath.Join(Config.Server.Assets, dnldcode)

	// Get our tracker info after we know that the download dir exists
	tracker, trackererr := LoadTracker(Config.Server.Assets, dnldcode)
	if trackererr != nil {
		log.Println("Tracker file not found: " + dnldcode)
		page.Message = "File(s) not found"
		DisplayPage(w, r, "files", page)
		return
	}
	page.Tracker = tracker

	// Get a basic EvhRequest object
	// Keep trying until we success (race warning!!!)
	req, reqerr := NewRequest(requestAddr)
	for reqerr != nil {
		req, _ = NewRequest(requestAddr)
	}
	req.Dnldcode = dnldcode

	// Send the requested file
	if isfiledownload {
		// Set the full path to the file
		fullpath := filepath.Join(ddpath, reqpathfile)
		req.Log(fullpath)

		// Decode for the real filename (person downloading would like to have the original name)
		realfname, berr := Base64Decode(filepath.Base(fullpath))
		if berr != nil {
			page.Message = "Error, cannot serve requested file (1000)."
			req.Log("ERROR: " + berr.Error())
			DisplayPage(w, r, "files", page)
			return
		}

		// Open the file for streaming
		fh, ferr := os.Open(fullpath)
		if ferr != nil {
			page.Message = "Error, cannot serve requested file (1001)."
			req.Log("ERROR: " + ferr.Error())
			DisplayPage(w, r, "files", page)
			return
		}

		// Log start of transfer
		req.Log("Sending \"" + realfname + "\" to " + requestAddr)

		// Set HTML headers and send file stream
		w.Header().Set("Content-Disposition", "attachment; filename=\""+realfname+"\"")
		w.Header().Set("Content-Type", "application/octet-stream")
		_, wrerr := io.Copy(w, fh)
		if wrerr != nil {
			req.Log("ERROR: " + wrerr.Error())
		} else {
			page.Tracker.AddLog(realfname + " downloaded by " + requestAddr)
			page.Tracker.Save()
		}

		// Log end of transfer
		req.Log("Sending \"" + realfname + "\" to " + requestAddr + "... finished.")

		// Don't send any more data
		return
	} else {
		// Verify vercode matches, or deny access
		if vercode != tracker.Vercode || vercode == "" {
			page.Message = "File(s) not found"
			req.Log("invalid vercode (" + vercode + " != " + page.Tracker.Vercode + "), access denied for " + requestAddr)
			DisplayPage(w, r, "files", page)
			return
		}
	}

	// Render the files.html template
	DisplayPage(w, r, "files", page)
}
