// Configuration is where we store the necessary information for
// this app to function properly.  Most pieces of this app
// access the same instance of a Configuration object.
//
// NOTE: If you want to add settings to the config file, you
// must add them here first and re-compile.  Otherwise the
// app will crash when loading the config.  Items missing
// from the config file are okay as they will just have their
// default (nil) values for their respective types.
package main

import (
	"code.google.com/p/gcfg"
	"log"
)

type Configuration struct {
	Main struct {
		AppName     string
		AppDesc     string
		CodeLenMin  int
		CodeLenMax  int
		Company     string
		CompanyURL  string
		Disclaimer  []string
		Logo        string
		Source      string
		Expirations []string `gcfg:"expiration"`
	}
	Daemon struct {
		PurgeFreq    int
		TmplFreq     int
		NoPurgeCheck []string
	}
	Server struct {
		Address    string
		AdminKey string
		Assets     string
		CertFile   string `gcfg:"cert"`
		KeyFile    string `gcfg:"key"`
		ListenAddr string
		MailUser   string
		MailPass   string
		Mailserver string
		NonSslPort string
		Ssl        bool
		SslPort    string
		Templates  string
	}
	Client struct {
		Description string
		DestEmail   string
		Email       string
		Expiration  string
		Field       string
		Progress    bool
		Url         string
	}
}

// This is a global variable
var Config Configuration

// Create a new instance of Configuration.
// An empty filename will result in an empty config, not an error.
func NewConfig(fpath string) (cfg Configuration) {
	err := gcfg.ReadFileInto(&cfg, fpath)
	if err != nil {
		log.Println(err.Error())
	}

	return
}

// Use CLI flag values to overwrite values in the parsed config.
func (c *Configuration) ImportFlags() {
	if UrlFlag != "" {
		c.Client.Url = UrlFlag
	}
	if FilesFieldFlag != "" {
		c.Client.Field = FilesFieldFlag
	}
	if SrcEmailFlag != "" {
		c.Client.Email = SrcEmailFlag
	}
	if DstEmailFlag != "" {
		c.Client.DestEmail = DstEmailFlag
	}
	if FileDescrFlag != "" {
		c.Client.Description = FileDescrFlag
	}
	if ProgressFlag != Config.Client.Progress {
		c.Client.Progress = ProgressFlag
	}
	if ExpirationFlag != "" {
		c.Client.Expiration = ExpirationFlag
	}
}
