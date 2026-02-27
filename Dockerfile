# syntax=docker/dockerfile:1

ARG GO_VERSION="1.25"
ARG ALPINE_VERSION="3.21"


### Go Builder ###
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk add --update --no-cache git bash make ca-certificates

WORKDIR /go/src/kosli

COPY . .

RUN make build

### Final Image ###
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/kosli/kosli /bin/kosli
ENTRYPOINT ["/bin/kosli"]
