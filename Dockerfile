# syntax=docker/dockerfile:1

# The version the binary reports; empty builds dev+<sha>. Passed in rather than
# read from git tags, so the image cannot depend on the checkout's tags (#1133).
ARG VERSION=""


### Go Builder ###
# Pinned by digest so a rebuild gets the same toolchain; Dependabot moves it.
FROM golang:1.27.3-alpine3.24@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS builder

RUN apk add --update --no-cache git bash make ca-certificates

ENV GOTOOLCHAIN=auto

WORKDIR /go/src/kosli

COPY . .

# Re-declare inside the stage — a global ARG is not in scope in a build stage.
ARG VERSION
RUN make build VERSION="${VERSION}"

RUN mkdir -p /image-tmp

### Final Image ###
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/kosli/kosli /bin/kosli
COPY --from=builder --chmod=1777 /image-tmp /tmp
ENTRYPOINT ["/bin/kosli"]
