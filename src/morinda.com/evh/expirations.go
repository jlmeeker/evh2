// Expirations are how the daemon knows when to clean up submitted files.
// These are set via the configuration file ONLY.  Having muliple is
// recommended.  There is no current way to set an expiration to
// unlimited (never delete).
//
// Expirations are set in the format of <integer>:<suffix>.  Current
// available suffixes are as follows:
//   m = minutes
//   h = hours
//   d = days
//   w = weeks
// A default expiration is hard-coded for 1 day (1:d) and is used as
// the value when an unsupported expiration is set by the client.
//
// NOTE: If you wand to add additional suffixes, please update the
// inline documentation in the sample-config.gcfg file.
package main

import (
	"log"
	"strconv"
	"strings"
	"time"
)

type Expiration struct {
	Key   string
	Value string
}

// Parse our expirations and create a usable map
func ExpandExpirations() map[string]time.Time {
	var expirations = make(map[string]time.Time)

	// Hard code one expiration so it is always available
	expirations["1:d"] = time.Now().Local().Add(time.Hour * 24 * 1)

	for _, abbrev := range Config.Main.Expirations {
		var parts = strings.Split(abbrev, ":")

		if len(parts) != 2 {
			log.Println("Error parsing expiration:", abbrev)
			continue
		}

		// Extract our integer
		durInt, durErr := strconv.Atoi(parts[0])
		if durErr != nil {
			log.Println("Error converting", parts[0], "to int:", durErr.Error())
		}

		// Calculate our epiration date
		var expireDate time.Time

		if parts[1] == "m" {
			expireDate = time.Now().Local().Add(time.Minute * time.Duration(durInt))
		} else if parts[1] == "h" {
			expireDate = time.Now().Local().Add(time.Hour * time.Duration(durInt))
		} else if parts[1] == "d" {
			expireDate = time.Now().Local().Add(time.Hour * 24 * time.Duration(durInt))
		} else if parts[1] == "w" {
			expireDate = time.Now().Local().Add(time.Hour * 24 * 7 * time.Duration(durInt))
		} else {
			log.Println("Error creating expire date, suffix unknown:", parts[1])
			continue
		}

		// Save our newly created time to the result
		expirations[abbrev] = expireDate
	}

	return expirations
}

// Expirations are guarantee to have both parts present in key
func ExpirationsToHtmlMap(expirations map[string]time.Time) map[int]Expiration {
	var result = make(map[int]Expiration)
	var counter = 0

	for key, _ := range expirations {
		var parts = strings.Split(key, ":")
		var suffix string
		var plural bool

		durInt, durErr := strconv.Atoi(parts[0])
		if durErr != nil {
			log.Println("Error converting", parts[0], "to int:", durErr.Error())
			continue
		}
		if durInt > 1 {
			plural = true
		}

		if parts[1] == "m" {
			suffix = "minute"
		} else if parts[1] == "h" {
			suffix = "hour"
		} else if parts[1] == "d" {
			suffix = "day"
		} else if parts[1] == "w" {
			suffix = "week"
		}

		var dispVal = parts[0] + " " + suffix
		if plural {
			dispVal = dispVal + "s"
		}

		result[counter] = Expiration{Key: key, Value: dispVal}
		counter++
	}

	return result
}
