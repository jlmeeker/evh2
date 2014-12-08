// The File object is used by the tracker for managing uploaded file properties.
package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// File object
type File struct {
	Name      string
	DirPath   string
	Path      string
	Size      float64
	SizeMB    float64
	Saved     bool
	Url       template.URL
	WhenSaved string
}

func (f *File) Save(r EvhRequest, tmpfh *multipart.FileHeader) error {
	// Generate full file path
	var fname = Base64Encode(f.Name)
	f.Path = filepath.Join(f.DirPath, fname)

	// Create destination file making sure the path is writeable.
	err := os.MkdirAll(f.DirPath, 0700)
	if err != nil {
		return errors.New(fmt.Sprintf("ERROR: could not create file dir (%s): %s", f.DirPath, err.Error()))
	}

	// Create destination file
	dst, dsterr := os.Create(f.Path)
	if dsterr != nil {
		return errors.New(fmt.Sprintf("ERROR: could create empty file: %s", dsterr.Error()))
	}

	// Open temp file for copying
	fhandle, tmpferr := tmpfh.Open()
	if tmpferr != nil {
		return errors.New(fmt.Sprintf("ERROR: could not open temp file for copy: %s", tmpferr.Error()))
	}

	// Copy the uploaded file to the destination file
	bytes, cpyerr := io.Copy(dst, fhandle)
	if cpyerr != nil {
		return errors.New(fmt.Sprintf("ERROR: could not write file contents: %s", cpyerr.Error()))
	}
	fhandle.Close()
	dst.Close()

	// This happened sometime, not sure why yet
	if bytes == 0 {
		return errors.New("ERROR: destination file size is 0 bytes!")
	} else {
		f.Size = float64(bytes)
		f.SizeMB = f.Size / 1024 / 1024
		f.Saved = true
		f.WhenSaved = time.Now().Local().Format(TimeLayout)
	}

	return nil
}

func (f *File) GenURL(dnldcode string) {
	f.Url = template.URL(HttpProto + "://" + Config.Server.Address + DownloadUrlPath + dnldcode + "/" + Base64Encode(f.Name))
}
