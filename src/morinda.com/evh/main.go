// EVH is designed to be a single-use file transfer system.  Its purpose is to replace
// aging methods of sharing files such as FTP.  With the advent of services like
// DropBox, Box, Google Drive and the like, this type of service is becoming more
// commonplace EVH has some differentiating features that make it an especially
// good tool for corporations and/or home use.
//
// EVH runs in two modes: server and client.  Server hosts a web server interface for
// uploading and downloading files.  The Client is for uploading only and runs
// in a terminal.  This app is designed to run on all platforms that Go supports.
package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
)

// Flags
var ConfigFileFlag string
var DstEmailFlag string
var ExpirationFlag string
var FileDescrFlag string
var FilesFieldFlag string
var ProgressFlag bool
var ServerFlag bool
var SrcEmailFlag string
var UrlFlag string
var Evh1ImportFlag bool

// Global Variables
var UploadUrlPath = "/upload/"
var DownloadUrlPath = "/download/"
var AdminUrlPath = "/admin/"
var Files []string
var HttpProto = "http"
var SiteDown bool
var Templates *template.Template

// Constants
const VERSION = "2.5.0"
const TimeLayout = "Jan 2, 2006 at 3:04pm (MST)"

func init() {
	flag.StringVar(&ConfigFileFlag, "c", "", "Location of the Configuration file")
	flag.BoolVar(&ServerFlag, "server", false, "Listen for incoming file uploads")

	// Client flags
	flag.StringVar(&UrlFlag, "url", "", "Remote server URL to send files to (client only)")
	flag.StringVar(&FilesFieldFlag, "field", "", "Field name of the form (client only)")
	flag.StringVar(&SrcEmailFlag, "from", "", "Email address of uploader (client only)")
	flag.StringVar(&DstEmailFlag, "to", "", "Comma separated set of email address(es) of file recipient(s) (client only)")
	flag.StringVar(&FileDescrFlag, "description", "", "File desription (use quotes) (client only)")
	flag.BoolVar(&ProgressFlag, "progress", true, "Show progress bar during upload (client only)")
	flag.StringVar(&ExpirationFlag, "expires", "", "Example 1:d for 1 day (client only)")
	flag.BoolVar(&Evh1ImportFlag, "import", false, "Import data from EVH1 instance (client only)")
}

func main() {
	flag.Parse()

	// Load in our Config
	Config = NewConfig(ConfigFileFlag)
	Config.ImportFlags()

	if ServerFlag {
		// Final sanity check
		if Config.Server.Assets == "" {
			log.Fatal("ERROR: Cannot continue without specifying assets path")
		}
		if Config.Server.Templates == "" {
			log.Fatal("ERROR: Cannot continue without specifying templates path")
		}
		if Config.Server.ListenAddr == "" {
			log.Fatal("ERROR: Cannot continue without specifying listenaddr value")
		}
		if Config.Server.Mailserver == "" {
			log.Println("WARNING: cannot send emails, mailserver not set")
		}

		// Set so all generated URLs use https if enabled
		if Config.Server.Ssl {
			HttpProto = "https"
		}

		// Setup our assets dir (if it don't already exist)
		err := os.MkdirAll(Config.Server.Assets, 0700)
		if err != nil {
			log.Fatal("Cannot setup assetdir as needed: " + err.Error())
		}

		// Parse our html templates
		go RefreshTemplates()
		go ScrubDownloads()

		// Register our handler functions
		http.HandleFunc(UploadUrlPath, SSLCheck(UploadHandler))
		http.HandleFunc(DownloadUrlPath, SSLCheck(AssetHandler))
		http.HandleFunc(AdminUrlPath, BasicAuth(SSLCheck(AdminHandler)))
		http.HandleFunc("/", Evh1Intercept(SSLCheck(HomeHandler)))

		// Listen
		log.Println("Listening...")

		// Spawn HTTPS listener in another thread
		go func() {
			if Config.Server.Ssl == false || Config.Server.SslPort == "" {
				return
			}
			var addrSsl = Config.Server.ListenAddr + ":" + Config.Server.SslPort
			listenErrSsl := http.ListenAndServeTLS(addrSsl, Config.Server.CertFile, Config.Server.KeyFile, nil)
			if listenErrSsl != nil {
				log.Fatal("ERROR: ssl listen problem: " + listenErrSsl.Error())
			}
		}()

		// Start non-SSL listener
		var addrNonSsl = Config.Server.ListenAddr + ":" + Config.Server.NonSslPort
		listenErr := http.ListenAndServe(addrNonSsl, nil)
		if listenErr != nil {
			log.Fatal("ERROR: non-ssl listen problem: " + listenErr.Error())
		}
	} else {
		// Run import if requested
		if Evh1ImportFlag {
			SpitSlurp()
			return
		}

		// Final sanity check
		if Config.Client.DestEmail == "" {
			log.Println("WARNING: no -destemail value set, cannot send reciever an email")
		}
		if Config.Client.Email == "" {
			log.Println("WARNING: no -email value set, cannot send email to uploader")
		}
		if Config.Client.Field == "" {
			log.Println("WARNING: no -field value set, using \"file\" instead")
			Config.Client.Field = "file"
		}
		if Config.Client.Url == "" {
			log.Fatal("ERROR: Cannot continue without specifying -url value")
		}

		// All filenames are unflagged arguments, loop through them and uplod the file(s)
		for _, fname := range flag.Args() {
			fi, err := os.Stat(fname)
			if err != nil {
				log.Println("WARNING: Cannot read file, skipping ", fname, ": ", err.Error())
			} else {
				if fi.Mode().IsRegular() {
					Files = append(Files, fname)
				}
			}
		}

		Upload(Files)
	}
}
