# ----------------------------------------------------------------------------
# global

APP = gcp-iam-lister

SHELL = /usr/bin/env bash

ifneq ($(shell command -v go),)
GO_PATH ?= $(shell go env GOPATH)
GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)

PKG := $(subst $(GO_PATH)/src/,,$(CURDIR))
GO_PACKAGES ?= $(shell go list ./...)
GO_TEST_PKGS := $(shell go list -f='{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)
endif

CGO_ENABLED ?= 0
GO_GCFLAGS=
GO_LDFLAGS=-s -w
GO_LDFLAGS_STATIC=-s -w '-extldflags=-static'

GO_BUILDTAGS=osusergo netgo
GO_BUILDTAGS_STATIC=static static_build
GO_INSTALLSUFFIX_STATIC=-installsuffix 'netgo'
GO_FLAGS = -tags='$(GO_BUILDTAGS)'
ifneq ($(GO_GCFLAGS),)
	GO_FLAGS+=-gcflags="${GO_GCFLAGS}"
endif
ifneq ($(GO_LDFLAGS),)
	GO_FLAGS+=-ldflags="${GO_LDFLAGS}"
endif

GO_TEST ?= go test
GO_TEST_FUNC ?= .
GO_BENCH_FUNC ?= .
GO_BENCH_FLAGS ?= -benchmem
ifneq ($(wildcard go.mod),)  # exist go.mod
ifneq ($(GO111MODULE),off)
	GO_TEST_FLAGS+=${GO_MOD_FLAGS}
	GO_BENCH_FLAGS+=${GO_MOD_FLAGS}
endif
endif
GO_TEST_COVERAGE_OUT := coverage.out
ifneq ($(CIRCLECI),)
	GO_TEST_COVERAGE_OUT=/tmp/ci/artifacts/coverage.out
endif

# ----------------------------------------------------------------------------
# defines

GOPHER = ""
define target
@printf "$(GOPHER)  \\x1b[1;32m$(patsubst ,$@,$(1))\\x1b[0m\\n"
endef

# ----------------------------------------------------------------------------
# targets

.DEFAULT_GOAL = static

## build and install

.PHONY: $(APP)
$(APP):
	$(call target)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go build -v $(strip $(GO_FLAGS)) -o $(APP) $(PKG)/cmd/${APP}

.PHONY: build
build: $(APP)  ## Builds a dynamic executable or package.

.PHONY: static
static: GO_LDFLAGS=${GO_LDFLAGS_STATIC}
static: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
static: GO_FLAGS+=${GO_INSTALLSUFFIX_STATIC}
static: $(APP)  ## Builds a static executable or package.

.PHONY: install
install: GO_FLAGS+=-mod=vendor
install: GO_LDFLAGS=${GO_LDFLAGS_STATIC}
install: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
install: GO_FLAGS+=${GO_INSTALLSUFFIX_STATIC}
install:  ## Installs the executable or package.
	$(call target)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go install -v $(strip $(GO_FLAGS)) $(CMD)

## test and coverage

.PHONY: test
test: CGO_ENABLED=1  # needs race test
test: GO_FLAGS+=-mod=vendor
test:  ## Runs package test including race condition.
	$(call target)
	@CGO_ENABLED=$(CGO_ENABLED) $(GO_TEST) -v -race $(strip $(GO_FLAGS)) -run=$(GO_TEST_FUNC) $(GO_TEST_PKGS)

.PHONY: coverage
coverage: GO_FLAGS+=-mod=vendor
coverage:  ## Takes packages test coverage.
	$(call target)
	CGO_ENABLED=$(CGO_ENABLED) $(GO_TEST) -v $(strip $(GO_FLAGS)) -covermode=atomic -coverpkg=./... -coverprofile=${GO_TEST_COVERAGE_OUT} $(GO_PACKAGES)

.PHONY: tools/go-junit-report
tools/go-junit-report:  # go get 'go-junit-report' binary
ifeq (, $(shell command -v go-junit-report))
	@cd $(mktemp -d); \
		go mod init tmp > /dev/null 2>&1; \
		go get -u github.com/jstemmer/go-junit-report@master
endif

.PHONY: coverage/ci
coverage/ci: GO_FLAGS+=-mod=vendor
coverage/ci: tools/go-junit-report
coverage/ci:  ## Takes packages test coverage, and output coverage results to CI artifacts.
	$(call target)
	@mkdir -p /tmp/ci/artifacts /tmp/ci/test-results
	CGO_ENABLED=$(CGO_ENABLED) $(GO_TEST) -a -v $(strip $(GO_FLAGS)) -covermode=atomic -coverpkg=$(PKG)/... -coverprofile=${GO_TEST_COVERAGE_OUT} $(GO_PACKAGES) 2>&1 | tee /dev/stderr | go-junit-report -set-exit-code > /tmp/ci/test-results/junit.xml
	@if [[ -f "${GO_TEST_COVERAGE_OUT}" ]]; then go tool cover -html=${GO_TEST_COVERAGE_OUT} -o $(dir GO_TEST_COVERAGE_OUT)/coverage.html; fi

## lint

.PHONY: lint
lint: lint/golangci-lint  ## Run all linters.

.PHONY: tools/golangci-lint
tools/golangci-lint:  # go get 'golangci-lint' binary
ifeq (, $(shell command -v golangci-lint))
	@GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
endif

.PHONY: lint/golangci-lint
lint/golangci-lint: tools/golangci-lint .golangci.yml  ## Run golangci-lint.
	$(call target)
	@golangci-lint run ./...

## clean

.PHONY: clean
clean:  ## Cleanups binaries and extra files in the package.
	$(call target)
	@$(RM) $(APP) *.out *.test *.prof trace.log

## miscellaneous

.PHONY: help
help:  ## Show make target help.
	@perl -nle 'BEGIN {printf "Usage:\n  make \033[33m<target>\033[0m\n\nTargets:\n"} printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 if /^([a-zA-Z\/_-].+)+:.*?\s+## (.*)/' ${MAKEFILE_LIST}
