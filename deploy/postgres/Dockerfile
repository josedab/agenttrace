FROM golang:1.25.12-alpine@sha256:56961d79ea8129efddcc0b8643fd8a5416b4e6228cfd477e3fd61deb2672c587 AS gosu-builder

RUN apk add --no-cache git \
    && git clone https://github.com/tianon/gosu.git /src/gosu \
    && cd /src/gosu \
    && git checkout 6456aaa0f3c854d199d0f037f068eb97515b7513 \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /go/bin/gosu .

FROM postgres:16-alpine@sha256:57c72fd2a128e416c7fcc499958864df5301e940bca0a56f58fddf30ffc07777

USER root
COPY --from=gosu-builder /go/bin/gosu /usr/local/bin/gosu
RUN apk upgrade --no-cache
