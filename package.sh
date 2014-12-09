#!/bin/bash

VERSION=`grep VERSION src/morinda.com/evh/main.go | cut -d "=" -f 2 | tr -d " \""`
./build.sh && cd ../ && tar -zcvf evh2-$VERSION.tar.gz evh2/LICENSE evh2/README.md evh2/sample-config.gcfg evh2/tmpl evh2/bin/
