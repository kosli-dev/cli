# syntax=docker/dockerfile:1

ARG GO_VERSION="1.19.5"
ARG ALPINE_VERSION="3.17"


### Go Builder ###
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk add --update --no-cache git bash make

WORKDIR /go/src/kosli

COPY . .

RUN make deps && make vet

RUN make build

### Final Image ###
FROM alpine:${ALPINE_VERSION} as base

RUN apk add --update --no-cache git openssh bash

COPY --from=builder /go/src/kosli/kosli /bin/kosli