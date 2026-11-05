# Use the official Go image as base
FROM golang:1.25-bookworm

# Install system dependencies
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        git \
        gnupg \
        make \
        nano \
        software-properties-common \
        vim \
        wget \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v1.62.2

# Install goreleaser
RUN echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | tee /etc/apt/sources.list.d/goreleaser.list \
    && apt-get update \
    && apt-get install -y goreleaser \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install Hugo (for documentation)
RUN wget https://github.com/gohugoio/hugo/releases/download/v0.139.4/hugo_extended_0.139.4_linux-amd64.deb \
    && dpkg -i hugo_extended_0.139.4_linux-amd64.deb \
    && rm hugo_extended_0.139.4_linux-amd64.deb

# Create a non-root user
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME -s /bin/bash \
    && apt-get update \
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# Set up Go workspace
ENV GOPATH=/go
RUN mkdir -p /go/pkg/mod && chown -R ${USERNAME}:${USERNAME} /go

# Switch to non-root user
USER $USERNAME

# Install Go tools that are commonly used
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest

# Set working directory
WORKDIR /workspace

# Persist bash history across container restarts
VOLUME ["/home/vscode/.bash_history"]
