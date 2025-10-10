ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)

EXCLUDE_DIRS := ./examples
EXCLUDE_GO_MOD_DIRS := $(shell find $(EXCLUDE_DIRS) -type f -name 'go.mod' -exec dirname {} \; | sort)

GO_BUILD_DIRS := $(ALL_GO_MOD_DIRS)

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
BIN_DIR := $(PROJECT_DIR)/.bin

fmt: golangci-lint
	@$(call run-go-mod-dir,$(GOLANGCI_LINT) fmt,".bin/golangci-lint fmt")

vet:
	@$(call run-go-mod-dir,go vet ./...,"go vet")

lint: golangci-lint
	$(GOLANGCI_LINT) run --timeout=10m -v --fix

.PHONY: tidy
tidy:
	@$(call run-go-mod-dir,go mod tidy,"go mod tidy")

.PHONY: generate
generate: mockery protoc
	@$(call run-go-mod-dir,go generate ./...,"go generate")

## test: Run all tests
.PHONY: test test-run test-cov test-xml

test:
	@$(call run-go-mod-dir,go test -race -covermode=atomic -coverprofile=coverage.out ./...,"go test")

test-cov:
	@$(call run-go-mod-dir-exclude,go tool cover -func=coverage.out,$(EXCLUDE_GO_MOD_DIRS),"go tool cover")

.PHONY: check
check: tidy fmt vet lint
	@git diff --quiet || test $$(git diff --name-only | grep -v -e 'go.mod$$' -e 'go.sum$$' | wc -l) -eq 0 || ( echo "The following changes (result of code generators and code checks) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty

# ========= Helpers ===========

GOLANGCI_LINT = $(BIN_DIR)/golangci-lint
golangci-lint:
	$(call go-get-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest)

MOCKERY = $(BIN_DIR)/mockery
mockery:
	$(call go-get-tool,$(MOCKERY),github.com/vektra/mockery/v2@v2.43.0)

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
