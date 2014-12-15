// The EvhRequest object is for managing upload/download session data.
// This data is not stored on disk.  See Tracker for persistent data.
// All console logging is done through the EvhRequest object when possible.
package main

import (
	"bytes"
	"log"
	"net/smtp"
	"strings"
	"time"
)

// Uploads object
type EvhRequest struct {
	AppVer      string
	Config      Configuration
	Dnldcode    string
	DownloadURL string
	Errors      []string
	Path        string
	SourceIP    string
	Tracker     Tracker
	When        time.Time
}

// Create a new EvhRequest object
func NewRequest(srcip string) (EvhRequest, error) {
	var req = EvhRequest{
		AppVer:   VERSION,
		When:     time.Now(),
		SourceIP: srcip,
		Config:   Config,
	}

	// Generate dnldcode
	var gcerr error
	req.Dnldcode, gcerr = GenCode(true, 0)

	return req, gcerr
}

// Write to console log (prepends the dnldcode to every message)
func (r *EvhRequest) Log(msgs ...string) {
	var line = r.Dnldcode

	for _, msg := range msgs {
		line = line + " " + msg
	}

	log.Println(line)
}

//func (r *EvhRequest) SendEmail(subject, body, sender, recipients string) {
func (r *EvhRequest) SendEmail(p *Page, tmplname string) {
	var fromEmail string
	var toEmails []string

	// Setup our recipient email address(es)
	if tmplname == "senderemail" {
		toEmails = strings.Split(p.Tracker.SrcEmail, ",")
	} else {
		toEmails = strings.Split(p.Tracker.DstEmail, ",")
	}

	// Setup our sender email address
	if Config.Server.AppEmail != "" {
		fromEmail = Config.Server.AppEmail
	} else {
		fromEmail = p.Tracker.SrcEmail
	}

	// Generate the body of the message
	buffer := new(bytes.Buffer)
	err := Templates.ExecuteTemplate(buffer, tmplname, p)
	if err != nil {
		r.Log(err.Error())
		return
	}
	var body = buffer.Bytes()

	// Send email
	if Config.Server.MailUser == "" {
		err = smtp.SendMail(Config.Server.Mailserver+":"+Config.Server.MailPort, nil, fromEmail, toEmails, body)
	} else {
		var auth = smtp.PlainAuth("", Config.Server.MailUser, Config.Server.MailPass, Config.Server.Mailserver)
		err = smtp.SendMail(Config.Server.Mailserver+":"+Config.Server.MailPort, auth, fromEmail, toEmails, body)
	}

	if err != nil {
		r.Log("ERROR: Could not send email:", err.Error())
		p.Tracker.AddLog("Error sending email to " + strings.Join(toEmails, ",") + ": " + err.Error())
	} else {
		p.Tracker.AddLog("Successfully sent email to " + strings.Join(toEmails, ",") + "")
	}
}

// Send Emails (incomplete and untested)
func (r *EvhRequest) Notify(p *Page) {
	if len(p.Tracker.Files) == 0 {
		r.Log("Cannot notify, no files successfully uploaded.")
		return
	}

	// send email to recipients
	if p.Tracker.DstEmail == "" {
		r.Log("Cannot notify, destination email is empty")
	} else {
		r.SendEmail(p, "destemail")
	}

	// send email to sender (self)
	if p.Tracker.SrcEmail == "" {
		r.Log("Cannot notify, uploader email is empty")
	} else {
		r.SendEmail(p, "senderemail")
	}
}
