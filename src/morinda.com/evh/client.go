// This code is based on an example of streaming http client uploads I found
// here: https://github.com/gebi/go-fileupload-example/blob/master/main.go.
//
// This code is the brains of the CLI version of this app. In a nutshell it
// builds the http connection and streams the files to the server. By streaming
// the files there is no maximum upload file size limit (unless the server
// doesn't have the necessary disk space).
package main

import (
	"crypto/tls"
	"fmt"
	"github.com/cheggaaa/pb"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Streams upload directly from file -> mime/multipart -> pipe -> http-request
func streamingUploadFile(params map[string]string, paths []string, w *io.PipeWriter, writer *multipart.Writer) {
	defer w.Close()

	// Set all of the standard parameters for the form submission
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	// Iterate over the files and send them one at a time
	//   This streams the files to the server one at a time
	//   as part of the form submission.
	for _, path := range paths {
		// Get some info on the the file we will be uploading
		fi, fierr := os.Stat(path)
		checkerr(fierr)

		// Open the file for reading
		file, err := os.Open(path)
		checkerr(err)
		defer file.Close()

		// Create our form field that will contain the file
		part, err := writer.CreateFormFile(Config.Client.Field, filepath.Base(path))
		checkerr(err)

		// Log that we're starting the transfer
		log.Println("Starting transfer of", filepath.Base(path))

		// Initialize the progress bar (if requested)
		if Config.Client.Progress {
			bar := pb.New(int(fi.Size())).SetUnits(pb.U_BYTES)
			bar.ShowTimeLeft = true
			bar.ShowSpeed = true
			bar.SetMaxWidth(120)
			bar.Start()

			// Pipe the file contents to both the bar counter and the server
			doublewriter := io.MultiWriter(part, bar)
			_, err = io.Copy(doublewriter, file)
			checkerr(err)
			bar.FinishPrint("Finished transfer of " + filepath.Base(path))
		} else {
			// Pipe the file contents to the server only
			_, err = io.Copy(part, file)
			checkerr(err)
			log.Println("Finished transfer of " + filepath.Base(path))
		}

		fmt.Println("")
	}

	// Be nice and close our server connection
	err := writer.Close()
	checkerr(err)
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(params map[string]string, paths []string) (*http.Request, error) {
	r, w := io.Pipe()
	writer := multipart.NewWriter(w)
	go streamingUploadFile(params, paths, w, writer)

	req, reqerr := http.NewRequest("POST", Config.Client.Url, r)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	return req, reqerr
}

// Called by main to initiate the client uploads
func Upload(fnames []string) {
	extraParams := map[string]string{
		"SrcEmail":  Config.Client.Email,
		"DstEmail":  Config.Client.DestEmail,
		"FileDescr": Config.Client.Description,
		"Expires":   Config.Client.Expiration,
		"client":    "1",
	}

	// Create and multipart form and the file data
	request, err := newfileUploadRequest(extraParams, fnames)
	checkerr(err)

	// Setup proxy server (if set)
	var proxyFunc func(*http.Request) (*url.URL, error)
	if Config.Client.Proxy != "" {
		if Config.Client.Proxy == "env" {
			proxyFunc = http.ProxyFromEnvironment
		} else {
			proxyUrl, urlErr := url.Parse(Config.Client.Proxy)
			if urlErr == nil {
				proxyFunc = http.ProxyURL(proxyUrl)
			} else {
				log.Println("ERROR: Proxy url is unusable. Trying without a proxy.")
			}
		}
	}

	// POST our http request (this will contact the server, submit the form and start the file streams)
	var tlsconfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	var transport http.RoundTripper = &http.Transport{
		TLSClientConfig: tlsconfig,
		Proxy:           proxyFunc,
	}

	// Create HTTP client object
	var client = &http.Client{
		Transport: transport,
	}

	// Get things started
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		// Copy the server response to stdout
		_, err := io.Copy(os.Stdout, resp.Body)
		checkerr(err)
		resp.Body.Close()
	}
}
