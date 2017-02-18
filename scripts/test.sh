#!/bin/bash

set -e

if [ "$1" = "ci" ]; then
    echo installing tools
    scripts/install_tools.sh
fi

echo building binaries
scripts/build.sh

echo running tests
non_vendor_pkgs=$(go list ./... | grep -v /vendor/)
non_vendor_dirs=$(ls -d */ | grep -v vendor/)
go test -race $non_vendor_pkgs

echo running cleaners
goimports -w $non_vendor_dirs
gofmt -s -w $non_vendor_dirs
misspell -w $non_vendor_dirs

echo running linters
go vet $non_vendor_pkgs
deadcode $non_vendor_dirs
golint $non_vendor_pkgs
aligncheck $non_vendor_pkgs
structcheck $non_vendor_pkgs
varcheck $non_vendor_pkgs
errcheck $non_vendor_pkgs
echo $non_vendor_dirs | xargs -n 1 ineffassign
interfacer $non_vendor_pkgs
unconvert -v -apply $non_vendor_pkgs
gosimple $non_vendor_pkgs
staticcheck $non_vendor_pkgs
unused $non_vendor_pkgs

echo 'All Tests Passed!'
