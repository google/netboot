#!/bin/sh

set -x
set -e

if [ -d /tmp/stuff/.git ]; then
    echo "Building from local dev copy"
    mkdir -p /tmp/go/src/go.universe.tf
    cp -R /tmp/stuff /tmp/go/src/go.universe.tf/netboot
else
    echo "Building from git checkout"
fi

export GOPATH=/tmp/go
export GLIDE_HOME=/tmp/go
apk -U add ca-certificates git go glide musl-dev
apk upgrade
go get -v -d go.universe.tf/netboot/cmd/pixiecore
cd /tmp/go/src/go.universe.tf/netboot
glide install
go test $(glide nv)
cd cmd/pixiecore
go build .
cp ./pixiecore /pixiecore
cd /
apk del --purge git go glide musl-dev
rm -rf /tmp/go /tmp/stuff /var/cache/apk/*
