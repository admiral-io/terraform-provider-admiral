# Use bash as the shell, with environment lookup
SHELL := /usr/bin/env bash

.DEFAULT_GOAL := build

MAKEFLAGS += --no-print-directory

VERSION ?= 0.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PROJECT_ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Tool binaries
GOLANGCI-LINT := ./tools/golangci-lint.sh

.PHONY: help # Print this help message.
help:
	@grep -E '^\.PHONY: [a-zA-Z_-]+ .*?# .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = "(: |#)"}; {printf "%-30s %s\n", $$2, $$3}'

.PHONY: build # Build the provider binary.
build:
	go build -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)" \
		-o ./build/terraform-provider-admiral .

.PHONY: install # Install the provider locally.
install: build
	go install -v ./...

.PHONY: generate # Generate documentation and code.
generate:
	go generate ./...

.PHONY: fmt # Format all Go source files.
fmt:
	gofmt -s -w -e .

.PHONY: lint # Lint the Go source code.
lint:
	$(GOLANGCI-LINT) run --timeout 2m30s

.PHONY: lint-fix # Lint and fix the Go source code.
lint-fix:
	$(GOLANGCI-LINT) run --fix
	go mod tidy

.PHONY: test # Run unit tests.
test:
	go test -race -covermode=atomic -timeout=120s ./...

.PHONY: testacc # Run acceptance tests.
testacc:
	TF_ACC=1 go test -race -covermode=atomic -timeout 120m -v ./...

.PHONY: verify # Verify go modules' requirements files are clean.
verify:
	go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "error: go.mod/go.sum are not tidy" && exit 1)

.PHONY: release # Tag and push the next version (auto-detected from commits).
release:
	@VERSION=$$(./tools/svu.sh next) && \
	echo "Current version: $$(./tools/svu.sh current)" && \
	echo "Next version:    $$VERSION" && \
	echo "" && \
	read -p "Proceed? [y/N] " confirm && [ "$$confirm" = "y" ] && \
	git tag -a $$VERSION -m "Release $$VERSION" && \
	git push origin $$VERSION

.PHONY: release-patch # Tag and push a patch release.
release-patch:
	@VERSION=$$(./tools/svu.sh patch) && \
	echo "Current version: $$(./tools/svu.sh current)" && \
	echo "Next version:    $$VERSION" && \
	git tag -a $$VERSION -m "Release $$VERSION" && \
	git push origin $$VERSION

.PHONY: release-minor # Tag and push a minor release.
release-minor:
	@VERSION=$$(./tools/svu.sh minor) && \
	echo "Current version: $$(./tools/svu.sh current)" && \
	echo "Next version:    $$VERSION" && \
	git tag -a $$VERSION -m "Release $$VERSION" && \
	git push origin $$VERSION

.PHONY: release-major # Tag and push a major release.
release-major:
	@VERSION=$$(./tools/svu.sh major) && \
	echo "Current version: $$(./tools/svu.sh current)" && \
	echo "Next version:    $$VERSION" && \
	git tag -a $$VERSION -m "Release $$VERSION" && \
	git push origin $$VERSION

.PHONY: version # Show current and next version.
version:
	@echo "Current: $$(./tools/svu.sh current)"
	@echo "Next:    $$(./tools/svu.sh next)"
	
.PHONY: clean # Remove build and cache artifacts.
clean:
	rm -rf build dist