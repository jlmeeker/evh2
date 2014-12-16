// uploadHandler controls showing the upload form and the
// processing of POSTed data.  Files are saved automatically
// (via http package) to env(TMPDIR). Files are then moved to
// the specified assets directory.
//
// NOTE: It is strongly recommended that you set the TMPDIR
// environment variable when you launch the evh service and
// set it to a directory on the same filesystem as assets.
// Moving the temp file to the permament location will be
// much faster this way.
package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

//This is where the action happens.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	var requestAddr = r.RemoteAddr

	// Get accurate client address
	if val, ok := r.Header["X-Forwarded-For"]; ok {
		requestAddr = strings.Join(val, ",")
	}

	// Get a new Page object
	var page = NewPage(r)

	if r.URL.Path != UploadUrlPath {
		page.StatusCode = 404
	}

	// Set the appropriate protocol prefix for URLs
	if r.TLS != nil {
		page.HttpProto = "https"
	}

	// Prep our available expirations
	var expirations = ExpandExpirations()
	page.Expirations = ExpirationsToHtmlMap(expirations)

	switch r.Method {
	// Show the upload form
	case "GET":
		DisplayPage(w, r, "upload", page)

	// Process form submission
	case "POST":
		// Initialize our file count to zero
		var filecount = 0

		// New request object
		req, reqerr := NewRequest(requestAddr)
		if reqerr != nil {
			req.Log(reqerr.Error())
			return
		}

		// Setup our tracker
		page.Tracker = NewTracker(req.Dnldcode)
		page.Tracker.Files = make(map[string]File)

		// This is the download URL for the session, not used for individual files
		req.Log("New incoming transfer starting for", requestAddr)

		// Parse the multipart form in the request (set max memory in bytes)
		err := r.ParseMultipartForm(10000)
		if err != nil {
			page.Message = template.HTML("Transfer aborted (client disconnected)")
			req.Log(string(page.Message))
			DisplayPage(w, r, "upload", page)
			return
		} else {
			// Store form values
			page.Tracker.Description = r.FormValue("FileDescr")
			page.Tracker.SrcEmail = r.FormValue("SrcEmail")
			page.Tracker.DstEmail = r.FormValue("DstEmail")
			page.Tracker.Expiration = r.FormValue("Expires")

			// Validate emails against restrictions
			if len(Config.Server.MailDomains) != 0 {
				var domainokay bool
				var allEmailAddresses = strings.Split(page.Tracker.DstEmail, ",")
				allEmailAddresses = append(allEmailAddresses, page.Tracker.SrcEmail)

				for _, addr := range allEmailAddresses {
					valid := ValidateEmailAddress(addr)
					if valid {
						domainokay = true
					}

					// An address failed, is this fatal?
					if !valid && Config.Server.MailDomainStrict {
						page.Message = template.HTML("Upload not authorized")
						req.Log("MailDomainStrict enabled and non-compliant email found, upload discarded: " + addr)
						DisplayPage(w, r, "upload", page)
						return
					}
				}

				// None of the addressed passed the check, deny further processing
				if !domainokay {
					page.Message = template.HTML("Upload not authorized")
					req.Log("MailDomains enabled and no compliant emails found, upload discarded")
					DisplayPage(w, r, "upload", page)
					return
				}
			}

			if r.FormValue("client") == "1" {
				page.Tracker.CliUpload = true
			}

			// Path to save file to
			req.Path = filepath.Join(Config.Server.Assets, req.Dnldcode)

			// Get the *fileheaders and keep count of uploadedjack filesN
			//   We don't care what the form field is called, just iterate over all form fields of type file
			var filename string
			var newfile File
			for fieldname, files := range r.MultipartForm.File {
				req.Log("Processing files field:", fieldname)
				for i, _ := range files {
					filename = ScrubFilename(files[i].Header.Get("Content-Disposition"))
					filecount++

					// Create a File object
					newfile = NewFile(filename, req.Path)

					// Move the temp file to the permament location
					err := newfile.Save(files[i])
					if err != nil {
						req.Errors = append(req.Errors, err.Error())
						req.Log()
					} else {
						req.Log(fmt.Sprintf("Saved file (%s, %.2f MB): %s", newfile.Name, newfile.Size/1024/1024, newfile.Path))
					}

					// Update our tracker
					page.Tracker.Files[newfile.Name] = newfile
					page.Tracker.Size += newfile.Size
					page.Tracker.SizeMB = page.Tracker.Size / 1024 / 1024
					page.Tracker.AddLog("Added file " + newfile.Name)
				}
			}

			// Set our expiration
			if val, ok := expirations[page.Tracker.Expiration]; ok {
				page.Tracker.ExpirationDate = val
			} else {
				req.Log("Invalid expiration specified, using default of 1 day")
				page.Tracker.ExpirationDate = expirations["1:d"]
			}

			// Send notification
			req.Notify(&page)
			page.Tracker.Save()
		}

		// DisplayPage result message (using template.HTML() allows the template to show the non-garbled URL)
		var filespageurl = page.HttpProto + "://" + page.RequestHost + DownloadUrlPath + page.Tracker.Dnldcode + "?vercode=" + page.Tracker.Vercode
		if r.FormValue("client") == "1" {
			page.Message = template.HTML(fmt.Sprintf("Successfully uploaded %d of %d files.  Your files are available here:\n%s\n", page.Tracker.CountSaved(), filecount, filespageurl))
			DisplayPage(w, r, "uploadPlain", page)
		} else {
			page.Message = template.HTML(fmt.Sprintf("Successfully uploaded %d of %d files.  Your files are available <a href=\"%s\">here</a>.", page.Tracker.CountSaved(), filecount, filespageurl))
			DisplayPage(w, r, "upload", page)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
