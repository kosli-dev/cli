export CGO_ENABLED=0
export GO111MODULE=on

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

LDFLAGS += -X github.com/merkely-development/watcher/internal/version.metadata=${VERSION_METADATA}
LDFLAGS += -X github.com/merkely-development/watcher/internal/version.gitCommit=${GIT_COMMIT}
LDFLAGS += -X github.com/merkely-development/watcher/internal/version.gitTreeState=${GIT_DIRTY}
LDFLAGS += -extldflags "-static"

fmt: ## Reformat package sources
	@go fmt ./...
.PHONY: fmt

lint:
	@docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.39-alpine golangci-lint run -v ./...
.PHONY: lint

vet: fmt
	@go vet ./...
.PHONY: vet

deps: ## Install depdendencies. Runs `go get` internally.
	@GOFLAGS="" go mod download
	@GOFLAGS="" go mod tidy
.PHONY: deps

build: deps vet ## Build the package
	@go build -o watcher -ldflags '$(LDFLAGS)' cmd/watcher/main.go
.PHONY: build

test: deps vet ## Run unit tests
	@go test -v -cover -p=1 ./...
.PHONY: test

docker: deps vet lint
	@docker build -t watcher .
.PHONY: docker	

