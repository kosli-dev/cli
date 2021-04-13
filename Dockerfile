ARG GO_VERSION="1.15.2"
ARG ALPINE_VERSION="3.12"


### Go Builder & Tester ###
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk add --update --no-cache git bash

WORKDIR /go/src/watcher

COPY . .
RUN go mod download && go mod tidy  

RUN CGO_ENABLED=0 GO111MODULE=on go build -o watcher -ldflags '-extldflags "-static"' main.go

### Final Image ###
FROM alpine:${ALPINE_VERSION} as base

RUN apk add --update --no-cache git openssh bash

COPY --from=builder /go/src/watcher/watcher /bin/watcher