.PHONY: help build clean clean-cache deps fmt lint vet docker
.DEFAULT_GOAL := help
.DELETE_ON_ERROR:
.DEFAULT: help
.DEFAULT_GOAL := help
.DELETE_ON_ERROR:

export CGO_ENABLED=0
export GO111MODULE=on

LDFLAGS := -w -s
BIN	:= kosli

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

GOTESTSUM  = $(shell which gotestsum || echo "~/go/bin/gotestsum")

# Fake CI env vars so that tests don't emit "Repo information will not be reported" warnings.
# These are only used when running tests locally (real CI already sets them).
FAKE_CI_ENV = GITHUB_RUN_NUMBER=1 GITHUB_SERVER_URL=https://github.com GITHUB_REPOSITORY=kosli-dev/cli GITHUB_REPOSITORY_ID=123456

ifdef VERSION
	BINARY_VERSION = $(VERSION)
endif
BINARY_VERSION ?= ${GIT_TAG}

# Only set Version if building a tag or VERSION is set
ifneq ($(BINARY_VERSION),)
	LDFLAGS += -X github.com/kosli-dev/cli/internal/version.version=${BINARY_VERSION}
endif

VERSION_METADATA = unreleased
# Clear the "unreleased" string in BuildMetadata
ifneq ($(GIT_TAG),)
	VERSION_METADATA =
endif

LDFLAGS += -X github.com/kosli-dev/cli/internal/version.metadata=${VERSION_METADATA}
LDFLAGS += -X github.com/kosli-dev/cli/internal/version.gitCommit=${GIT_COMMIT}
LDFLAGS += -X github.com/kosli-dev/cli/internal/version.gitTreeState=${GIT_DIRTY}
LDFLAGS += -extldflags "-static"

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
	  /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
	  /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

ldflags: ## Print ldflags
	@echo $(LDFLAGS)

fmt: ## Reformat package sources
	@go fmt ./...

ensure_golangci-lint:
	@$HOMEBREW_NO_AUTO_UPDATE=1 brew upgrade golangci-lint

lint: deps vet ensure_golangci-lint ## Run linting
	@golangci-lint run --timeout=5m --color always  -v ./...

vet: fmt ## Run Go vet
	@go vet ./...

deps: ## Install depdendencies. Runs `go get` internally.
	@GOFLAGS="" go mod download
	@GOFLAGS="" go mod tidy

build: deps vet ## Build the binary
	@go build -o $(BIN) -ldflags '$(LDFLAGS)' ./cmd/kosli/

check_dirty:
	@git diff-index --quiet HEAD --  || echo "Cannot test release with dirty git repo"
	@git diff-index --quiet HEAD --

add_test_tag:
	@git tag -d v0.0.99 2> /dev/null || true
	@git tag v0.0.99

build_release: check_dirty add_test_tag
	rm -rf dist/
	goreleaser release --skip-publish --debug
	@git tag -d v0.0.99 2> /dev/null || true

clean: ## Clean build artefacts
	rm -rf $(BIN) dist/

clean-cache: ## Clean Golang mod cache
	go clean --modcache

ensure_network:
	docker network inspect cli_net > /dev/null || docker network create --driver bridge cli_net

ensure_gotestsum:
	@go install gotest.tools/gotestsum@latest

clear_local_image_ref:
	rm /tmp/server-image.txt

test_setup: ensure_gotestsum
# cat and exit if error
	@test -f /tmp/server-image.txt || ./hack/get-server-image.sh /tmp/server-image.txt
	export KOSLI_SERVER_IMAGE=$$(cat /tmp/server-image.txt) && ./bin/reset-or-start-server.sh

test_setup_restart_server: ensure_gotestsum
# cat and exit if error
	@test -f /tmp/server-image.txt || ./hack/get-server-image.sh /tmp/server-image.txt
	export KOSLI_SERVER_IMAGE=$$(cat /tmp/server-image.txt) && ./bin/reset-or-start-server.sh force

setup_test_to_use_local_image:
	@echo merkely > /tmp/server-image.txt
	@docker ps -aq | xargs -r docker rm -fv
	@echo "Run make build in the server repo you want to use"
	@echo "Then run make test_integration"
	@echo "To look at the logs from local kosli server run: make follow_integration_test_server"

setup_test_to_use_staging_server_image:
	@rm /tmp/server-image.txt
	@docker ps -aq | xargs -r docker rm -fv
	@echo "Now run make test_integration"
	@echo "To look at the logs from kosli server run: make follow_integration_test_server"

test_integration: deps vet ensure_network test_setup ## Run tests except the too slow ones
	@[ -e ~/.kosli.yml ] && mv ~/.kosli.yml ~/.kosli-renamed.yml || true
	@export KOSLI_TESTS=true $(FAKE_CI_ENV) && $(GOTESTSUM) -- --short -p=8 -coverprofile=cover.out ./...
	@go tool cover -func=cover.out | grep total:
	@go tool cover -html=cover.out
	@[ -e ~/.kosli-renamed.yml ] && mv ~/.kosli-renamed.yml ~/.kosli.yml || true


test_integration_full: deps vet ensure_network test_setup ## Run all tests
	@[ -e ~/.kosli.yml ] && mv ~/.kosli.yml ~/.kosli-renamed.yml || true
	@mkdir -p junit-test-results
	@export KOSLI_TESTS=true $(FAKE_CI_ENV) && $(GOTESTSUM) --junitfile junit-test-results/junit.xml -- -p=8 -coverprofile=cover.out ./...
	@go tool cover -func=cover.out
	@[ -e ~/.kosli-renamed.yml ] && mv ~/.kosli-renamed.yml ~/.kosli.yml || true


test_integration_restart_server: test_setup_restart_server
	@[ -e ~/.kosli.yml ] && mv ~/.kosli.yml ~/.kosli-renamed.yml || true
	@export KOSLI_TESTS=true $(FAKE_CI_ENV) && $(GOTESTSUM) -- --short -p=8 -coverprofile=cover.out ./...
	@go tool cover -html=cover.out
	@[ -e ~/.kosli-renamed.yml ] && mv ~/.kosli-renamed.yml ~/.kosli.yml || true

test_integration_single: test_setup
	@export KOSLI_TESTS=true $(FAKE_CI_ENV) && $(GOTESTSUM) -- -p=1 ./... -run "${TARGET}"


test_docs: deps vet ensure_network test_setup ## Test docs
	./bin/test_docs_cmds.sh docs.kosli.com/content/use_cases/simulating_a_devops_system/_index.md

logs_integration_test_server:
	@docker logs cli_kosli_server ${CONTAINER} 2>&1

follow_integration_test_server:
	@docker logs cli_kosli_server -f ${CONTAINER} 2>&1

enter_integration_test_server:
	@docker exec -it --workdir / cli_kosli_server bash

docker: ## Build CLI Docker image
	@docker build -t kosli-cli .

cli-docs: build ## Generate docs
	@rm -f docs.kosli.com/content/client_reference/kosli*
	@export DOCS=true && ./kosli docs --dir docs.kosli.com/content/client_reference

legacy-ref-docs:
	@./hack/generate-old-versions-docs.sh "v2.*"

licenses: ## Update licenses
	@rm -rf licenses || true
	@go install github.com/google/go-licenses@latest
	@go-licenses save ./... --save_path="licenses/" || true
	$(eval DATA := $(shell go-licenses csv ./...))
	@echo $(DATA) | tr " " "\n" > licenses/licenses.csv

upgrade-deps: ## Update Go dependencies
	@go get -u ./...

generate-json-metadata: ## Generate docs metadata
	echo '{"currentversion": "vlocal"}' > docs.kosli.com/assets/metadata.json

hugo: cli-docs helm-docs generate-json-metadata ## Run docs locally
	cd docs.kosli.com && hugo server --minify --buildDrafts --port=1515

hugo-local: cli-docs generate-json-metadata
	cd docs.kosli.com && hugo server --minify --buildDrafts --port=1515

helm-lint: ## Link Helm chart
	@cd charts/k8s-reporter && helm lint . \
		--set reporterConfig.kosliOrg=placeholder \
		--set 'reporterConfig.environments[0].name=placeholder'

helm-docs: helm-lint ## Update Helm docs
	@cd charts/k8s-reporter &&  docker run --rm --volume "$(PWD):/helm-docs" jnorwood/helm-docs:latest --template-files README.md.gotmpl,_templates.gotmpl --output-file README.md
	@cd charts/k8s-reporter &&  docker run --rm --volume "$(PWD):/helm-docs" jnorwood/helm-docs:latest --template-files README.md.gotmpl,_templates.gotmpl --output-file ../../docs.kosli.com/content/helm/_index.md

# Suggest next semver and changelog using Claude.
# Writes changelog to dist/release_notes.md for use with goreleaser --release-notes.
# Requires: jq, curl, op (1Password CLI). API key from 1Password via op.
# Usage: make suggest-version-ai [BASE_REF=v1.2.3]
suggest-version-ai: ## Suggest next version using AI
	@command -v jq >/dev/null 2>&1 || (echo "Install jq (e.g. brew install jq)" && exit 1)
	@command -v curl >/dev/null 2>&1 || (echo "Install curl (e.g. brew install curl)" && exit 1)
	@bin/suggest-version-ai.sh $(BASE_REF) -o dist/release_notes.md

# Release: without tag → suggest version + changelog, then interactive edit & confirm, then tag and push.
# With tag → escape hatch: create annotated tag (body = dist/release_notes.md if present), push. No AI, no prompt.
# Release notes are carried in the tag message so GitHub Actions can pass them to GoReleaser.
release: ## Cut a new release and push next tag
	@current=$$(git branch --show-current 2>/dev/null || git rev-parse --abbrev-ref HEAD); \
	if [ "$$current" != "main" ]; then echo "ERROR: release must be run from main branch (current: $$current)"; exit 1; fi; \
	if [ -z "$(tag)" ]; then \
	  command -v jq >/dev/null 2>&1 || (echo "Install jq (e.g. brew install jq)" && exit 1); \
	  command -v curl >/dev/null 2>&1 || (echo "Install curl (e.g. brew install curl)" && exit 1); \
	  bin/suggest-version-ai.sh -o dist/release_notes.md; \
	  if [ ! -f dist/suggested_version ]; then \
	    echo "Suggestion failed or no previous tag. Use: make release tag=vX.Y.Z"; exit 1; \
	  fi; \
	  bin/release-interactive.sh; \
	else \
	  git remote update; \
	  git status -uno | grep --silent "Your branch is up to date" || (echo "ERROR: your branch is NOT up to date with remote" && exit 1); \
	  ([ -f dist/release_notes.md ] && git tag -a $(tag) -F dist/release_notes.md) || git tag -a $(tag) -m"$(tag)"; \
	  git push origin $(tag); \
	fi

# check-links:
# 	@docker run -v ${PWD}:/tmp:ro --rm -i --entrypoint '' ghcr.io/tcort/markdown-link-check:stable /bin/sh -c 'find /tmp/docs.kosli.com/content -name \*.md -print0 | xargs -0 -n1 markdown-link-check -q -c /tmp/link-checker-config.json'

check-links: ## Run html test for dead links
	@cd docs.kosli.com && hugo --minify
	@docker run -v ${PWD}:/test --rm wjdp/htmltest -c .htmltest.yml -l 1
