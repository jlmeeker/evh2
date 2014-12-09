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

func assetHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to SSL if enabled
	if r.TLS == nil && Config.Server.Ssl {
		redirectToSsl(w, r)
		return
	}

	var dnldcode string
	var isfiledownload bool

	// Get a new Page object
	var page = NewPage()

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
	req, reqerr := NewRequest(r.RemoteAddr)
	for reqerr != nil {
		req, _ = NewRequest(r.RemoteAddr)
	}
	req.Dnldcode = dnldcode

	// Prepare directory listing
	if isfiledownload {
		// Send the requested file
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
		req.Log("Sending \"" + realfname + "\" to " + r.RemoteAddr)
		page.Tracker.AddLog("Download by " + r.RemoteAddr)
		page.Tracker.Save()

		// Set HTML headers and send file stream
		w.Header().Set("Content-Disposition", "attachment; filename=\""+realfname+"\"")
		w.Header().Set("Content-Type", "application/octet-stream")
		_, wrerr := io.Copy(w, fh)
		if wrerr != nil {
			req.Log("ERROR: " + wrerr.Error())
		}

		// Log end of transfer
		req.Log("Sending \"" + realfname + "\" to " + r.RemoteAddr + "... finished.")

		// Don't send any more data
		return
	} else {
		// Verify vercode matches, or deny access
		if vercode != tracker.Vercode || vercode == "" {
			page.Message = "File(s) not found"
			req.Log("invalid vercode (" + vercode + " != " + page.Tracker.Vercode + "), access denied for " + r.RemoteAddr)
			DisplayPage(w, r, "files", page)
			return
		}
	}

	// Render the files.html template
	DisplayPage(w, r, "files", page)
}
