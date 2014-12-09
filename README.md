evh2
====

Event Horizon (evh) a single-use file transfer system. Version 2.

EVH is designed to be a single-use file transfer system.  Its purpose is to replace aging methods of sharing files such as FTP.  With the advent of services like DropBox, Box, Google Drive and the like, this type of service is becoming more commonplace EVH has some differentiating features that make it an especially good tool for corporations and/or home use.  These features include:

* Easy to run on your own server (and port)
* Just a single binary and a tmpl directory to get started
* VERY low memory usage (~10-20mb)
* File transfers are never loaded into memory, always streamed directly to/from disk
* All uploads have an expiration.  As long as the server is running, cleanup happens automatically.
* EMail notifications to uploader and to any number of recipients (with download links)
* Web interface for viewing all files at once (requires emailed URL for access and is nearly impossible to guess)
* Multiple file uploads (use Ctrl or Shift or Command when selecting multiple files)
* Unlimited file size
* Can be run as a client (to upload files) via CLI in a terminal (with progress bar)
* Runs on Linux, Windows, OSX (same OS and architecture lists as Go)
* Secure against XSS, SQL injection etc.
* even more....

EVH runs in two modes: server and client.  Server hosts a web server interface for uploading and downloading files.  The Client is for uploading only and runs in a terminal.  This app is designed to run on all platforms that Go supports.

Bulding:
```
./build.sh
```

Executing:
```Bash
# Server (listen)
evh -server -c <path to config>

# Client (could use -c <configfile> instead of options)
evh -client [options] <file1... file2...>
evh -client -c <config file> <file1... file2...>
```

Other options:
```Bash
user@hostname $ evh -h
Usage of ./bin/evh:
  -c="": Location of the Configuration file
  -description="": File desription (use quotes) (client only)
  -expires="": Example 1:d for 1 day (client only)
  -field="": Field name of the form (client only)
  -from="": Email address of uploader (client only)
  -progress=true: Show progress bar during upload (client only)
  -server=false: Listen for incoming file uploads
  -to="": Comma separated set of email address(es) of file recipient(s) (client only)
  -url="": Remote server URL to send files to (client only)
```

Sample output:
```Bash
# Server output
user@hostname $ evh -server -c local-config.gcfg
2014/12/08 16:24:35 321b207a1d76a30f4d2598239c2c5 New incoming transfer starting for 127.0.0.1:44400
2014/12/08 16:24:35 321b207a1d76a30f4d2598239c2c5 Processing files field: file
2014/12/08 16:24:35 321b207a1d76a30f4d2598239c2c5 Saved file (evh, 7.39 MB): assets/321b207a1d76a30f4d2598239c2c5/ZXZo
2014/12/08 16:24:35 321b207a1d76a30f4d2598239c2c5 Invalid expiration specified, using default of 1 day

# Client command and output for above server request
user@hostname $ evh -to sample@myemail.com,sample2@theiremail.com -from myself@me.org -description "Test upload" -url http://127.0.0.1:8080/upload/ evh
2014/12/08 16:24:35 open : no such file or directory
2014/12/08 16:24:35 WARNING: no -field value set, using "file" instead
2014/12/08 16:24:35 Starting transfer of evh
7.39 MB / 7.39 MB [=============================================================================] 100.00 % 554.02 MB/s 0
Finished transfer of evh


Successfully uploaded 1 of 1 files.  Your files are available here:
http://127.0.0.1:8080/download/321b207a1d76a30f4d2598239c2c5?vercode=2a71c29a76cb8d3a63c873
```
