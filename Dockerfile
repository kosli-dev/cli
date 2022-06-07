ARG GO_VERSION="1.17.11"
ARG ALPINE_VERSION="3.16"


### Go Builder & Tester ###
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk add --update --no-cache git bash make

WORKDIR /go/src/merkely

COPY . .

RUN make build

### Final Image ###
FROM alpine:${ALPINE_VERSION} as base

RUN apk add --update --no-cache git openssh bash

COPY --from=builder /go/src/merkely/merkely /bin/merkely