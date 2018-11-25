FROM alpine:latest

RUN apk add --no-cache ca-certificates
ADD pixiecore /pixiecore
ENTRYPOINT ["/pixiecore"]
