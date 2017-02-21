#!/bin/bash

# Setup environment
#
	export GOPATH=`pwd`


# These are required is required
#
	# http://godoc.org/code.google.com/p/gcfg
	if [ ! -e src/gopkg.in/gcfg.v1 ]; then
		go get gopkg.in/gcfg.v1
	fi

	# https://github.com/cheggaaa/pb
	if [ ! -e src/github.com/cheggaaa/pb ]; then
		go get github.com/cheggaaa/pb
	fi

	# https://github.com/go-sql-driver/mysql
	if [ ! -e src/github.com/go-sql-driver/mysql ]; then
		go get github.com/go-sql-driver/mysql
	fi

	# (windows only) https://github.com/olekukonko/ts
	if [ ! -e src/github.com/olekukonko/ts ]; then
		go get github.com/olekukonko/ts
	fi


# Format Go sources
#
	go fmt src/morinda.com/evh/*.go


# Compiler flags (default flags omit debug information)
#
	FLAGS='-ldflags "-w"'


# Build (cleanup then build for 64- and 32-bit on the native platform)
#
	rm -rf $GOPATH/bin/*
	rm -rf $GOPATH/pkg/*
	GOOS=linux GOARCH=amd64 go install $FLAGS morinda.com/evh
	GOOS=linux GOARCH=386 go install $FLAGS morinda.com/evh
	GOOS=windows GOARCH=amd64 go install $FLAGS morinda.com/evh
	GOOS=darwin GOARCH=amd64 go install $FLAGS morinda.com/evh
