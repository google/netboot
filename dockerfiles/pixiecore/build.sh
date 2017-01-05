#!/bin/sh

set -x
set -e

apk upgrade --update-cache
apk add ca-certificates
apk add --virtual .build-deps  \
    git                        \
    go                         \
    glide                      \
    musl-dev

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
apk del --purge .build-deps
rm -rf /tmp/go /tmp/stuff /var/cache/apk/*
