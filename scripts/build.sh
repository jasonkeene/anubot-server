#!/bin/bash

set -e

cmd_path=github.com/jasonkeene/anubot-server/cmd
go_dir=$(echo "$GOPATH" | tr ':' ' ' | awk '{print $1}')

ls "$go_dir/src/$cmd_path" | while read line; do
    echo building $line
    go install "$cmd_path/$line"
done
