FROM alpine:edge

MAINTAINER David Anderson <dave@natulte.net>

COPY . /tmp/sandbox/context

RUN set -x;                                                                     \
    set -e;                                                                     \
                                                                                \
    apk upgrade --update-cache;                                                 \
    apk add ca-certificates;                                                    \
    apk add --virtual .build-deps                                               \
        git                                                                     \
        go                                                                      \
        glide                                                                   \
        musl-dev;                                                               \
                                                                                \
    if [ -d /tmp/sandbox/context/cmd/pixiecore ]; then                          \
        echo "Building from local dev copy";                                    \
        mkdir -p /tmp/sandbox/go/src/go.universe.tf;                            \
        mv -v /tmp/sandbox/context /tmp/sandbox/go/src/go.universe.tf/netboot;  \
    else                                                                        \
        echo "Building from git checkout";                                      \
    fi;                                                                         \
                                                                                \
    export GOPATH=/tmp/sandbox/go;                                              \
    export GLIDE_HOME=/tmp/sandbox;                                             \
    go get -v -d go.universe.tf/netboot/cmd/pixiecore;                          \
    cd /tmp/sandbox/go/src/go.universe.tf/netboot;                              \
    glide install;                                                              \
    go test $(glide nv);                                                        \
    GOBIN=/usr/local/bin go install ./cmd/pixiecore;                            \
    apk del --purge .build-deps;                                                \
    rm -rf /tmp/sandbox /var/cache/apk/*;

ENTRYPOINT ["/usr/local/bin/pixiecore"]
