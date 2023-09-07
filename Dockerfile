# syntax=docker/dockerfile:1

ARG GO_VERSION="1.21.1"
ARG ALPINE_VERSION="3.18"


### Go Builder ###
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk add --update --no-cache git bash make ca-certificates

WORKDIR /go/src/kosli

COPY . .

RUN make deps && make vet

RUN make build

### Final Image ###
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/kosli/kosli /bin/kosli
ENTRYPOINT ["/bin/kosli"]
