// These functions are http handlers too small to warrant their own file.
//
// NOTE: You REALLY want SSL setup if you are using this!!!
package main

import (
	"encoding/base64"
	"net/http"
	"strings"
)

type handler func(w http.ResponseWriter, r *http.Request)

// Handles / requests
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to SSL if enabled
	if r.TLS == nil && Config.Server.Ssl {
		redirectToSsl(w, r)
		return
	}

	var page = NewPage()
	DisplayPage(w, r, "home", page)
}

// Handles /admin/ requests
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	var page = NewPage()
	if Config.Server.AllowAdmin {
		page.TrackerOfTrackers = NewTrackerOfTrackers()
	} else {
		page.Message = "Access denied."
	}

	DisplayPage(w, r, "admin", page)
}

// Wrapper for checking for SSL (if required)
func SSLCheck(pass handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// Redirect to SSL if enabled
		if r.TLS == nil && Config.Server.Ssl {
			redirectToSsl(w, r)
			return
		}

		pass(w, r)
	}
}

// Helper function for SSLCheck
func redirectToSsl(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+Config.Server.Address+":"+Config.Server.SslPort+r.RequestURI, http.StatusTemporaryRedirect)
	return
}

// Wrapper for ensuring HTML Basic Authentication
// This is a modified copy of part of the example found here:
// http://bl.ocks.org/tristanwietsma/8444cf3cb5a1ac496203
func BasicAuth(pass handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Header["Authorization"]; !ok {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
			http.Error(w, "Authorization Required", http.StatusUnauthorized)
			return
		}

		auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "bad syntax", http.StatusBadRequest)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !Validate(pair[0], pair[1]) {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		pass(w, r)
	}
}

// Helper function for BasicAuth
func Validate(username, password string) bool {
	if username == Config.Server.AdminUser && password == Config.Server.AdminPass {
		return true
	}
	return false
}
