#!/usr/bin/env bash

set -e
echo "building binary"
cd /go/app
CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-s -w" -o /go/bin/app .
echo "starting app"
/go/bin/app