ARG GO_VERSION="1.15.2"
ARG ALPINE_VERSION="3.12"


### Go Builder & Tester ###
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk add --update --no-cache git bash make

WORKDIR /go/src/watcher

COPY . .

RUN make build

### Final Image ###
FROM alpine:${ALPINE_VERSION} as base

RUN apk add --update --no-cache git openssh bash

COPY --from=builder /go/src/watcher/watcher /bin/watcher