THIS_MAKE_FILE := $(lastword $(MAKEFILE_LIST))
PRE_RELEASE_SUFFIX?=alpha
IMAGE_TAG?=latest
SHELL := bash
BUILD_ENV := .envrc
CURRENT_DIR = $(shell pwd)

.PHONY:	all
all:	help	clean build	test	vendor	tidy	run	lint	codecov	pre-release	release	git-tag

# Ensure what image build we are doing pre-release or release
IF_PRE_RELEASE_BUILD := $(filter pre-release,$(MAKECMDGOALS))

ifeq "$(IF_PRE_RELEASE_BUILD)" "pre-release"
$(eval IMAGE_TAG := $(shell svu next --suffix=$(PRE_RELEASE_SUFFIX)))
else
$(eval IMAGE_TAG := $(shell svu next))
endif
	
build:	## Build the app
	goreleaser build --snapshot --rm-dist --single-target --debug

test:	## Run Tests
	./hack/test.sh

vendor:		tidy	## Run go mod vendor
	go mod vendor

tidy:	## Run go mod tidy
	go mod tidy

run:	## Runs the binary locally
	go run -mod=vendor cmd/main.go

lint:	## Lint the code
	golangci-lint run

codecov:	test	## Code Coverage
	bash <(curl -s https://codecov.io/bash)

clean:	## Cleans the build and release directory
	go clean
	rm -rf dist

build-image: ## builds the container image
	./hack/generate_yamls.sh $(CURRENT_DIR) $(CURRENT_DIR)/manifests $(IMAGE_TAG)

pre-release:	build-image git-tag	## builds the pre-release container image with $PRE_RELEASE_SUFFIX

release:	build-image	git-tag	## does a GA release

git-tag:	
	git tag "$(IMAGE_TAG)"
	git push --tags

help: ## Show this help
	@echo Please specify a build target. The choices are:
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(INFO_COLOR)%-30s$(NO_COLOR) %s\n", $$1, $$2}'