#!/bin/bash

set -e

echo installing goimports
go get -u golang.org/x/tools/cmd/goimports
echo installing deadcode
go get -u github.com/tsenart/deadcode
echo installing golint
go get -u github.com/golang/lint/golint
echo installing aligncheck
go get -u github.com/opennota/check/cmd/aligncheck
echo installing structcheck
go get -u github.com/opennota/check/cmd/structcheck
echo installing varcheck
go get -u github.com/opennota/check/cmd/varcheck
echo installing errcheck
go get -u github.com/kisielk/errcheck
echo installing ineffassign
go get -u github.com/gordonklaus/ineffassign
echo installing interfacer
go get -u github.com/mvdan/interfacer/cmd/interfacer
echo installing unconvert
go get -u github.com/mdempsky/unconvert
echo installing gosimple
go get -u honnef.co/go/tools/cmd/gosimple
echo installing staticcheck
go get -u honnef.co/go/tools/cmd/staticcheck
echo installing unused
go get -u honnef.co/go/tools/cmd/unused
echo installing misspell
go get -u github.com/client9/misspell/cmd/misspell
