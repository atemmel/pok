#!/usr/bin/sh

go build -ldflags -H=windowsgui ./cmd/pok 
go build -ldflags -H=windowsgui ./cmd/poked
go build -ldflags -H=windowsgui ./cmd/dialogue-helper
