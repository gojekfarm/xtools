ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)
# extract go minor version from go version output
GO_MINOR_VERSION := $(shell go version | cut -d' ' -f3 | cut -d'.' -f2)

EXCLUDE_DIRS := ./examples
EXCLUDE_GO_MOD_DIRS := $(shell find $(EXCLUDE_DIRS) -type f -name 'go.mod' -exec dirname {} \; | sort)

# set build directory based on go minor version
GO_BUILD_DIRS := $(foreach dir,$(ALL_GO_MOD_DIRS),$(shell GO_MOD_VERSION=$$(grep "go 1.[0-9]*" $(dir)/go.mod | cut -d' ' -f2 | cut -d'.' -f2) && [ -n "$$GO_MOD_VERSION" ] && [ $(GO_MINOR_VERSION) -ge $$GO_MOD_VERSION ] && echo $(dir)))

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
BIN_DIR := $(PROJECT_DIR)/.bin

fmt:
	@$(call run-go-mod-dir,go fmt ./...,"go fmt")

gofmt:
	@$(call run-go-mod-dir,go fmt ./...,"gofmt -s -w")

vet:
	@$(call run-go-mod-dir,go vet ./...,"go vet")

lint: golangci-lint
	@$(call run-go-mod-dir,$(GOLANGCI_LINT) run --timeout=10m -v,".bin/golangci-lint")

.PHONY: ci
ci: test test-cov test-xml

imports: gci
	@$(call run-go-mod-dir,$(GCI) -w -local github.com/gojekfarm ./ | { grep -v -e 'skip file .*' || true; },".bin/gci")

.PHONY: gomod.tidy
gomod.tidy:
	@$(call run-go-mod-dir,go mod tidy,"go mod tidy")

.PHONY: go.generate
go.generate: mockery protoc
	@$(call run-go-mod-dir,go generate ./...,"go generate")

## test: Run all tests
.PHONY: test test-run test-cov test-xml

test: check test-run

test-run:
	@$(call run-go-mod-dir,go test -race -covermode=atomic -coverprofile=coverage.out ./...,"go test")

test-cov: gocov
	@$(call run-go-mod-dir-exclude,$(GOCOV) convert coverage.out > coverage.json,$(EXCLUDE_GO_MOD_DIRS),"gocov convert")
	@$(call run-go-mod-dir-exclude,$(GOCOV) convert coverage.out | $(GOCOV) report,$(EXCLUDE_GO_MOD_DIRS),"gocov report")

test-xml: test-cov gocov-xml
	@jq -n '{ Packages: [ inputs.Packages ] | add }' $(shell find . -type f -name 'coverage.json' | sort) | $(GOCOVXML) > coverage.xml

.PHONY: check
check: fmt vet imports lint
	@git diff --quiet || test $$(git diff --name-only | grep -v -e 'go.mod$$' -e 'go.sum$$' | wc -l) -eq 0 || ( echo "The following changes (result of code generators and code checks) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty

# ========= Helpers ===========

## Determine the golangci-lint version based on $(GO_MINOR_VERSION)
###################
GOLANGCI_LINT_V18 := v1.50.1
GOLANGCI_LINT_DEFAULT := v1.53.3

get-golangci-lint-version = $(or $(value GOLANGCI_LINT_V$(1)), $(GOLANGCI_LINT_DEFAULT))
GOLANGCI_LINT_VERSION := $(call get-golangci-lint-version,$(GO_MINOR_VERSION))

###################

GOLANGCI_LINT = $(BIN_DIR)/golangci-lint
golangci-lint:
	$(call go-get-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))

GCI = $(BIN_DIR)/gci
gci:
	$(call go-get-tool,$(GCI),github.com/daixiang0/gci@v0.2.9)

GOCOV = $(BIN_DIR)/gocov
gocov:
	$(call go-get-tool,$(GOCOV),github.com/axw/gocov/gocov@v1.0.0)

GOCOVXML = $(BIN_DIR)/gocov-xml
gocov-xml:
	$(call go-get-tool,$(GOCOVXML),github.com/AlekSi/gocov-xml@v1.0.0)

MOCKERY = $(BIN_DIR)/mockery
mockery:
	$(call go-get-tool,$(MOCKERY),github.com/vektra/mockery/v2@v2.20.0)

PROTOC = $(BIN_DIR)/protoc
protoc:
	$(call go-get-tool,$(PROTOC),google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1)

# go-get-tool will 'go get' any package $2 and install it to $1.
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(BIN_DIR) go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# run-go-mod-dir runs the given $1 command in all the directories with
# a go.mod file
define run-go-mod-dir
set -e; \
for dir in $(GO_BUILD_DIRS); do \
	[ -z $(2) ] || echo "$(2) $${dir}/..."; \
	cd "$(PROJECT_DIR)/$${dir}" && PATH=$(BIN_DIR):$$PATH $(1); \
done;
endef

# run-go-mod-dir-exclude runs the given $1 command in all the directories with
# a go.mod file except the directories in $2
define run-go-mod-dir-exclude
set -e; \
for dir in $(filter-out $(2),$(GO_BUILD_DIRS)); do \
	[ -z $(3) ] || echo "$(3) $${dir}/..."; \
	cd "$(PROJECT_DIR)/$${dir}" && PATH=$(BIN_DIR):$$PATH $(1); \
done;
endef
