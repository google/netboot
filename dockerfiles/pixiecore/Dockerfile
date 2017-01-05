FROM alpine:edge

MAINTAINER David Anderson <dave@natulte.net>

ENV PIXIECORE_SANDBOX /tmp/sandbox
ENV PIXIECORE_CONTEXT "$PIXIECORE_SANDBOX"/context

COPY . "$PIXIECORE_CONTEXT"

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
    if [ -d "$PIXIECORE_CONTEXT"/cmd/pixiecore ]; then                          \
        echo "Building from local dev copy";                                    \
        mkdir -p /tmp/sandbox/go/src/go.universe.tf;                            \
        mv -v "$PIXIECORE_CONTEXT" /tmp/sandbox/go/src/go.universe.tf/netboot;  \
    else                                                                        \
        echo "Building from git checkout";                                      \
    fi;                                                                         \
                                                                                \
    export GOPATH="$PIXIECORE_SANDBOX"/go;                                      \
    export GLIDE_HOME="$PIXIECORE_SANDBOX";                                     \
    go get -v -d go.universe.tf/netboot/cmd/pixiecore;                          \
    cd /tmp/sandbox/go/src/go.universe.tf/netboot;                              \
    glide install;                                                              \
    go test $(glide nv);                                                        \
    GOBIN=/usr/local/bin go install ./cmd/pixiecore;                            \
    apk del --purge .build-deps;                                                \
    rm -rf "$PIXIECORE_SANDBOX" /var/cache/apk/*;

ENTRYPOINT ["/usr/local/bin/pixiecore"]
