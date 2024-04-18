##
# Gojira Makefile
#

SHELL=bash

VERSION=0.11.5
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
