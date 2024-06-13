##
# Gojira Makefile
#

SHELL=bash

VERSION=0.12.3

REPOSITORY="https://github.com/mhersson/gojira.git"

# make will interpret non-option arguments in the command line as targets.
# This turns them into do-nothing targets, so make won't complain:
# If the first argument is "run"...
ifeq (run,$(firstword $(MAKECMDGOALS)))
# use the rest as arguments for "run"
	RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
# ...and turn them into do-nothing targets
	$(eval $(RUN_ARGS):;@:)
endif

# Get the tag on the latest commit, if it exists, and strip the leading 'v'
LATEST_TAG = $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null | sed 's/^v//')

# Check if LATEST_TAG is set, if so, override the VERSION variable
# This ensures that the VERSION variable is set to the latest tag
# if I forget to update the VERSION variable manually
ifneq ($(LATEST_TAG),)
	VERSION := $(LATEST_TAG)
endif

# Example target that prints the version
.PHONY: show-version
show-version:
	@echo $(VERSION)

GIT_COMMIT=$(shell git rev-parse --short HEAD)

LDFLAGS="-X github.com/mhersson/gojira/cmd.GojiraGitRevision=$(GIT_COMMIT) \
-X github.com/mhersson/gojira/cmd.GojiraVersion=$(VERSION) \
-X github.com/mhersson/gojira/cmd.GojiraRepository=$(REPOSITORY)"

all: build

fmt:
	go fmt ./...

vet:
	go vet ./...

test: fmt vet
	go test ./... -coverprofile cover.out
lint:
	@golangci-lint run ./... --enable-all \
--disable gochecknoinits,gochecknoglobals,gomnd,gofumpt,\
gci,exhaustivestruct,forbidigo,funlen,nestif,cyclop,scopelint,\
maligned,interfacer,tagliatelle,golint --print-issued-lines=false


build: fmt vet
	@go build -ldflags $(LDFLAGS)

install: fmt vet
	@go install -ldflags $(LDFLAGS)

run:
	@go run -ldflags $(LDFLAGS) ./main.go $(RUN_ARGS)

# end
