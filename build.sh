#!/usr/bin/env sh

CGO_ENABLED=0 go build -installsuffix cgo -ldflags '-s -w' -o /go/bin/app
app
