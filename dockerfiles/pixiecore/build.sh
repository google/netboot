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
apk -U add ca-certificates git go gcc musl-dev
apk upgrade
go get -v github.com/Masterminds/glide
go get -v -d go.universe.tf/netboot/cmd/pixiecore
cd /tmp/go/src/go.universe.tf/netboot
/tmp/go/bin/glide install
go test $(/tmp/go/bin/glide nv)
cd cmd/pixiecore
go build .
cp ./pixiecore /pixiecore
cd /
apk del git go gcc musl-dev
rm -rf /tmp/go /tmp/stuff /root/.glide /usr/lib/go /var/cache/apk/*
