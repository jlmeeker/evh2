// These are a collection of smaller functions that don't really
// fit anywhere else.
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

// Wrapper for testing error
func checkerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Base64Encode returns a string representation of the Base64
// encoded value of a string.
func Base64Encode(text string) string {
	return base64.URLEncoding.EncodeToString([]byte(text))
}

// Base64Decode reurns the string representation of the
// decoded Base64 string.
func Base64Decode(text string) (string, error) {
	bytes, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		log.Printf("ERROR: cannot decode (%s): %s", text, err.Error())
		return "", err
	}

	return string(bytes), nil
}

// GenCode creates a random string for use as dnldcode or a vercode
func GenCode(dnld bool, tries int) (string, error) {
	// Minimum code length
	var mincodelen = int64(20)

	// Abort if we're having a hard time generating a unique code
	if tries > 10 {
		var msg = fmt.Sprintf("Too many code collissions: %d", tries)
		return "", errors.New(msg)
	}

	// Have a variation in the length of the generated code
	var ceil = *big.NewInt(int64(Config.Main.CodeLenMax - Config.Main.CodeLenMin))
	randlen, err := rand.Int(rand.Reader, &ceil)
	if err != nil {
		log.Printf("ERROR: could not generate randlen: %s", err.Error())
	}

	var codelen = mincodelen + randlen.Int64()
	var bytes = make([]byte, codelen)

	// Read some random data
	_, err = rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Sha the random data and choose our slice as the code
	var sha = fmt.Sprintf("%x", sha256.Sum256(bytes))
	var code = sha[:codelen]

	// Test and generate a new one if we need to (only for dnld codes)
	if dnld {
		var dirpath = filepath.Join(Config.Server.Assets, code)
		_, ferr := os.Stat(dirpath)
		if ferr == nil {
			tries++
			code, err = GenCode(dnld, tries)

			if err != nil {
				return "", err
			}
		}

		if tries > 10 {
			log.Printf("NOTICE: %d tries to get a unique dnldcode.", tries+1)
		}
	}

	return code, nil
}

// Take a full file name/path and returns just the filename
func ScrubFilename(input string) (result string) {
	var filename = ExtractFilename(input)

	// Replace all slashes with os.PathSeparator
	filename = strings.Replace(filename, "/", string(os.PathSeparator), -1)
	filename = strings.Replace(filename, "\\", string(os.PathSeparator), -1)

	// remove Native slashes
	result = filepath.Base(filename)

	return
}

// Extract filename from content-disposition
func ExtractFilename(cd string) string {
	var cdParts = strings.Split(cd, "; ")
	//log.Printf("cdParts: %#v\n", cdParts)
	for _, part := range cdParts {
		//log.Printf("part: %#v\n", part)
		if strings.HasPrefix(part, "filename=") {
			var fname = strings.TrimPrefix(part, "filename=\"")
			fname = strings.TrimSuffix(fname, "\"")
			return fname
		}
	}

	return ""
}

// Delete directory, recursive
func RemoveDir(dirpath string) error {
	return os.RemoveAll(dirpath)
}

// Return a list of email domains from addresses
func EmailDomain(addr string) (domain string) {
	var addrParts = strings.Split(addr, "@")
	if len(addrParts) == 2 {
		return strings.ToLower(addrParts[1])
	}

	return ""
}

// Validates email(s) against config
func ValidateEmailAddress(addr string) bool {
	var emdomain string
	emdomain = EmailDomain(addr)

	for _, domain := range Config.Server.MailDomains {
		if strings.HasSuffix(emdomain, strings.ToLower(domain)) {
			return true
		}
	}

	return false
}
