# syntax=docker/dockerfile:1

ARG GO_VERSION="1.26"
ARG ALPINE_VERSION="3.23"


### Go Builder ###
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk add --update --no-cache git bash make ca-certificates

ENV GOTOOLCHAIN=auto

WORKDIR /go/src/kosli

COPY . .

RUN make build

RUN mkdir -p /image-tmp

### Final Image ###
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/kosli/kosli /bin/kosli
COPY --from=builder --chmod=1777 /image-tmp /tmp
ENTRYPOINT ["/bin/kosli"]
