# Set V to 1 for verbose output from the Makefile
Q=$(if $V,,@)
SRC=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

all: lint test

ci: test

.PHONY: all ci

#########################################
# Build
#########################################

build: ;

#########################################
# Bootstrapping
#########################################

bootstra%:
	$Q curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.49
	$Q go install golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: bootstrap

#########################################
# Test
#########################################

test:
	$Q $(GOFLAGS) go test -coverprofile=coverage.out ./...

race:
	$Q $(GOFLAGS) go test -race ./...

.PHONY: test race

#########################################
# Linting
#########################################

fmt:
	$Q goimports -local github.com/golangci/golangci-lint -l -w $(SRC)

lint: SHELL:=/bin/bash
lint:
	$Q LOG_LEVEL=error golangci-lint run --config <(curl -s https://raw.githubusercontent.com/smallstep/workflows/master/.golangci.yml) --timeout=30m
	$Q govulncheck ./...

.PHONY: fmt lint
