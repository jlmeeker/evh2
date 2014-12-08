#!/bin/bash

# Setup environment
#
	export GOPATH=`pwd`
	export GOBIN=$GOPATH/bin

# These are required is required
#
	# http://godoc.org/code.google.com/p/gcfg
	if [ ! -e src/code.google.com/p/gcfg ]; then
		go get code.google.com/p/gcfg
	fi

	# https://github.com/cheggaaa/pb
	if [ ! -e src/github.com/cheggaaa/pb ]; then
		go get github.com/cheggaaa/pb
	fi


# Format Go sources
#
	go fmt src/morinda.com/evh/*.go

# Build
#
	rm -rf $GOBIN/*
	rm -rf $GOPATH/pkg/*
	go install morinda.com/evh
	GOARCH=386 go install morinda.com/evh
