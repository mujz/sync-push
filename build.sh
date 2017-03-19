#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# osx binary
echo "Building Darwin's binary"
go build -o bin/sync-push-osx

# IMPORTANT! docker builds should be at the bottom since we need to run the docker images from the GOPATH

# debian binary
echo "Building Debian's binary"
cd $GOPATH
docker run --name sync -it --rm -v $(pwd):/go --workdir /go/src/github.com/mujz/sync-push golang go build -o bin/sync-push-debian sync-push.go
