// These functions handle checking for HTML Basic Authentication
// This is a copy of part of the example found here:
// http://bl.ocks.org/tristanwietsma/8444cf3cb5a1ac496203
//
// NOTE: You REALLY want SSL setup if you are using this!!!
package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"
)

type handler func(w http.ResponseWriter, r *http.Request)

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

func Validate(username, password string) bool {
	if username == Config.Server.AdminUser && password == Config.Server.AdminPass {
		return true
	}
	return false
}
