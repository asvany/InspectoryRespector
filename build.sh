#!/bin/bash
set -x
set -e
go generate
go build
go test ./...

