#!/bin/bash

set -e

go get -u golang.org/x/tools/cmd/goimports
go get -u github.com/tsenart/deadcode
go get -u github.com/golang/lint/golint
go get -u github.com/opennota/check/cmd/aligncheck
go get -u github.com/opennota/check/cmd/structcheck
go get -u github.com/opennota/check/cmd/varcheck
go get -u github.com/kisielk/errcheck
go get -u github.com/gordonklaus/ineffassign
go get -u github.com/mvdan/interfacer/cmd/interfacer
go get -u github.com/mdempsky/unconvert
go get -u honnef.co/go/simple/cmd/gosimple
go get -u honnef.co/go/staticcheck/cmd/staticcheck
go get -u honnef.co/go/unused/cmd/unused
go get -u github.com/client9/misspell/cmd/misspell
