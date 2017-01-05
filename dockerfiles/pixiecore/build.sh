#!/bin/sh

set -x
set -e

apk upgrade --update-cache
apk add ca-certificates git go glide musl-dev

if [ -d /tmp/stuff/.git ]; then
    echo "Building from local dev copy"
    mkdir -p /tmp/go/src/go.universe.tf
    mv -v /tmp/stuff /tmp/go/src/go.universe.tf/netboot
else
    echo "Building from git checkout"
fi

export GOPATH=/tmp/go
export GLIDE_HOME=/tmp/go
go get -v -d go.universe.tf/netboot/cmd/pixiecore
cd /tmp/go/src/go.universe.tf/netboot
glide install
go test $(glide nv)
GOBIN=/ go install ./cmd/pixiecore
cd /
apk del --purge git go glide musl-dev
rm -rf /tmp/go /tmp/stuff /var/cache/apk/*
