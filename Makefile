SHELL := /usr/bin/env bash
.DEFAULT_GOAL := all

MAKEFLAGS += --no-print-directory

PROJECT_ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Build metadata injected into main via -ldflags. GoReleaser sets the same
# at release time (see .goreleaser.yml).
VERSION ?= 0.0.0
GO_LDFLAGS := -s -w -X main.version=$(VERSION)

GOLANGCI-LINT := $(PROJECT_ROOT_DIR)/tools/golangci-lint.sh
SVU := $(PROJECT_ROOT_DIR)/tools/svu.sh

.PHONY: help # Print this help message.
help:
	@grep -E '^\.PHONY: [a-zA-Z0-9_-]+ .*?# .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = "(: |#)"}; {printf "%-30s %s\n", $$2, $$3}'

.PHONY: all # Build the provider.
all: build

.PHONY: build # Build the provider binary.
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="$(GO_LDFLAGS)" -o bin/terraform-provider-admiral .

.PHONY: install # Install the provider binary to GOPATH/bin.
install:
	CGO_ENABLED=0 go install -trimpath -ldflags="$(GO_LDFLAGS)" .

.PHONY: test # Run unit tests.
test:
	go test -race -covermode=atomic ./...

.PHONY: test-verbose # Run unit tests with verbose output.
test-verbose:
	go test -v -race -covermode=atomic ./...

# Acceptance tests run against a real Admiral API and need ADMIRAL_API_KEY.
# They create and delete applications in that tenant.
.PHONY: testacc # Run acceptance tests (needs ADMIRAL_API_KEY).
testacc:
	TF_ACC=1 go test -race -covermode=atomic -timeout 120m -v ./...

.PHONY: lint # Lint the code.
lint:
	$(GOLANGCI-LINT) run --timeout 2m30s

.PHONY: lint-fix # Lint and fix the code.
lint-fix:
	$(GOLANGCI-LINT) run --fix
	go mod tidy

.PHONY: fmt # Format the code.
fmt:
	go fmt ./...

# tfplugindocs renders docs/ from the provider schema, examples/ and
# templates/. It needs a terraform binary on PATH.
.PHONY: generate # Regenerate docs/ from the provider schema and examples.
generate:
	go generate ./...

# A schema or example change alters docs/ without anyone touching docs/,
# which is easy to miss in review -- so it is checked rather than remembered.
.PHONY: generate-verify # Fail if docs/ is stale.
generate-verify: generate
	@$(PROJECT_ROOT_DIR)/tools/ensure-no-diff.sh docs

.PHONY: verify # Verify go modules are tidy.
verify: generate-verify
	go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod or go.sum is not tidy" && exit 1)

.PHONY: deps # Download dependencies.
deps:
	go mod download
	go mod tidy

.PHONY: release # Tag and push the next version (auto-detected from commits).
release:
	@VERSION=$$($(SVU) next) && \
	echo "Current version: $$($(SVU) current)" && \
	echo "Next version:    $$VERSION" && \
	echo "" && \
	read -p "Proceed? [y/N] " confirm && [ "$$confirm" = "y" ] && \
	git tag -a $$VERSION -m "Release $$VERSION" && \
	git push origin $$VERSION

.PHONY: release-patch # Tag and push a patch release.
release-patch:
	@VERSION=$$($(SVU) patch) && \
	echo "Current version: $$($(SVU) current)" && \
	echo "Next version:    $$VERSION" && \
	git tag -a $$VERSION -m "Release $$VERSION" && \
	git push origin $$VERSION

.PHONY: release-minor # Tag and push a minor release.
release-minor:
	@VERSION=$$($(SVU) minor) && \
	echo "Current version: $$($(SVU) current)" && \
	echo "Next version:    $$VERSION" && \
	git tag -a $$VERSION -m "Release $$VERSION" && \
	git push origin $$VERSION

.PHONY: release-major # Tag and push a major release.
release-major:
	@VERSION=$$($(SVU) major) && \
	echo "Current version: $$($(SVU) current)" && \
	echo "Next version:    $$VERSION" && \
	git tag -a $$VERSION -m "Release $$VERSION" && \
	git push origin $$VERSION

.PHONY: version # Show current and next version.
version:
	@echo "Current: $$($(SVU) current)"
	@echo "Next:    $$($(SVU) next)"

.PHONY: clean # Remove build artifacts.
clean:
	rm -rf bin/ dist/
