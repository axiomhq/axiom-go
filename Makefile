# TOOLCHAIN
GO		:= CGO_ENABLED=0 GOBIN=$(CURDIR)/bin go
GOFMT	:= $(GO)fmt

# ENVIRONMENT
VERBOSE		=

# GO TOOLS
GOLANGCI_LINT	:= bin/golangci-lint
GORELEASER		:= bin/goreleaser
GOTESTSUM		:= bin/gotestsum
STRINGER		:= bin/stringer

# MISC
COVERPROFILE	:= coverage.out
DIST_DIR		:= dist

# FLAGS
GO_TEST_FLAGS		:= -race -coverprofile=$(COVERPROFILE)
GORELEASER_FLAGS	:= --snapshot --rm-dist

# DEPENDENCIES
GOMODDEPS = go.mod go.sum

# Enable verbose test output if explicitly set.
GOTESTSUM_FLAGS	=
ifdef VERBOSE
	GOTESTSUM_FLAGS += --format=standard-verbose
endif

# FUNCTIONS
# func go-list-pkg-sources(package)
go-list-pkg-sources = $(GO) list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}' $(1)
# func go-pkg-sourcefiles(package)
go-pkg-sourcefiles = $(shell $(call go-list-pkg-sources,$(strip $1)))

.PHONY: all
all: dep generate fmt lint test ## Run dep, generate, fmt, lint and test

.PHONY: clean
clean: ## Remove build and test artifacts
	@echo ">> cleaning up artifacts"
	@rm -rf $(COVERPROFILE) $(DIST_DIR)

.PHONY: cover
cover: $(COVERPROFILE) ## Calculate the code coverage score
	@echo ">> calculating code coverage"
	@$(GO) tool cover -func=$(COVERPROFILE) | tail -n1

.PHONY: dep-clean
dep-clean: ## Remove obsolete dependencies
	@echo ">> cleaning dependencies"
	@$(GO) mod tidy

.PHONY: dep-upgrade
dep-upgrade: ## Upgrade all direct dependencies to their latest version
	@echo ">> upgrading dependencies"
	@$(GO) get -d $(shell $(GO) list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	@make dep

.PHONY: dep
dep: dep-clean dep.stamp ## Install and verify dependencies and remove obsolete ones

dep.stamp: $(GOMODDEPS)
	@echo ">> installing dependencies"
	@$(GO) mod download
	@$(GO) mod verify
	@touch $@

.PHONY: fmt
fmt: ## Format and simplify the source code using `gofmt`
	@echo ">> formatting code"
	@! $(GOFMT) -s -w $(shell find . -path -prune -o -name '*.go' -print) | grep '^'

.PHONY: generate
generate: $(STRINGER) axiom/datasets_string.go axiom/monitors_string.go axiom/notifiers_string.go axiom/starred_string.go axiom/users_string.go ## Generate code using `go generate`

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint the source code
	@echo ">> linting code"
	@$(GOLANGCI_LINT) run

.PHONY: test-integration
test-integration: $(GOTESTSUM) ## Run all unit and integration tests. Run with VERBOSE=1 to get verbose test output ('-v' flag). Requires AXM_ACCESS_TOKEN and AXM_DEPLOYMENT_URL to be set.
	@echo ">> running integration tests"
	@$(GOTESTSUM) $(GOTESTSUM_FLAGS) -- $(GO_TEST_FLAGS) -tags=integration ./...

.PHONY: test
test: $(GOTESTSUM) ## Run all unit tests. Run with VERBOSE=1 to get verbose test output ('-v' flag).
	@echo ">> running tests"
	@$(GOTESTSUM) $(GOTESTSUM_FLAGS) -- $(GO_TEST_FLAGS) ./...

.PHONY: tools
tools: $(GOLANGCI_LINT) $(GORELEASER) $(GOTESTSUM) $(STRINGER) ## Install all tools into the projects local $GOBIN directory

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# GO GENERATE TARGETS

axiom/%_string.go: axiom/%.go
	@echo ">> generating $@ from $<"
	@$(GO) generate $<

# MISC TARGETS

$(COVERPROFILE):
	@make test

# GO TOOLS

$(GOLANGCI_LINT): dep.stamp $(call go-pkg-sourcefiles, github.com/golangci/golangci-lint/cmd/golangci-lint)
	@echo ">> installing golangci-lint"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint

$(GORELEASER): dep.stamp $(call go-pkg-sourcefiles, github.com/goreleaser/goreleaser)
	@echo ">> installing goreleaser"
	@$(GO) install github.com/goreleaser/goreleaser

$(GOTESTSUM): dep.stamp $(call go-pkg-sourcefiles, gotest.tools/gotestsum)
	@echo ">> installing gotestsum"
	@$(GO) install gotest.tools/gotestsum

$(STRINGER): dep.stamp $(call go-pkg-sourcefiles, golang.org/x/tools/cmd/stringer)
	@echo ">> installing stringer"
	@$(GO) install golang.org/x/tools/cmd/stringer
