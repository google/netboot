FROM alpine:3.5

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
    # Pixiecore assets                                                          \
    NAMESPACE=go.universe.tf;                                                   \
    REPO=netboot;                                                               \
    PKG=cmd/pixiecore;                                                          \
                                                                                \
    export GOPATH="$PIXIECORE_SANDBOX"/go;                                      \
    export GLIDE_HOME="$PIXIECORE_SANDBOX";                                     \
                                                                                \
    NAMESPACE_PATH="$GOPATH/src/$NAMESPACE";                                    \
    REPO_PATH="$NAMESPACE_PATH/$REPO";                                          \
    PKG_PATH="$REPO_PATH/$PKG";                                                 \
                                                                                \
    if [ -d "$PIXIECORE_CONTEXT"/"$PKG" ]; then                                 \
        echo "Building from local dev copy";                                    \
        mkdir -p "$NAMESPACE_PATH";                                             \
        mv -v "$PIXIECORE_CONTEXT" "$REPO_PATH";                                \
    else                                                                        \
        echo "Building from upstream git repo";                                 \
    fi;                                                                         \
                                                                                \
    go get -v -d "$NAMESPACE/$REPO/$PKG";                                       \
    cd "$REPO_PATH";                                                            \
    glide install;                                                              \
    go test $(glide nv);                                                        \
    GOBIN=/usr/local/bin go install -ldflags -s ./"$PKG";                       \
                                                                                \
    apk del --purge .build-deps;                                                \
    rm -rf "$PIXIECORE_SANDBOX" /var/cache/apk/*;

ENTRYPOINT ["/usr/local/bin/pixiecore"]
