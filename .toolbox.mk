## toolbox - start
## Generated with https://github.com/bakito/toolbox

## Current working directory
TB_LOCALDIR ?= $(shell which cygpath > /dev/null 2>&1 && cygpath -m $$(pwd) || pwd)
## Location to install dependencies to
TB_LOCALBIN ?= $(TB_LOCALDIR)/bin
$(TB_LOCALBIN):
	if [ ! -e $(TB_LOCALBIN) ]; then mkdir -p $(TB_LOCALBIN); fi

# Helper functions
STRIP_V = $(patsubst v%,%,$(1))

## Tool Binaries
TB_DEEPCOPY_GEN ?= $(TB_LOCALBIN)/deepcopy-gen
TB_GINKGO ?= $(TB_LOCALBIN)/ginkgo
TB_GOLANGCI_LINT ?= $(TB_LOCALBIN)/golangci-lint
TB_GORELEASER ?= $(TB_LOCALBIN)/goreleaser
TB_SEMVER ?= $(TB_LOCALBIN)/semver

## Tool Versions
# renovate: packageName=github.com/kubernetes/code-generator
TB_DEEPCOPY_GEN_VERSION ?= v0.34.0
# renovate: packageName=github.com/onsi/ginkgo/v2
TB_GINKGO_VERSION ?= v2.25.2
# renovate: packageName=github.com/golangci/golangci-lint/v2
TB_GOLANGCI_LINT_VERSION ?= v2.4.0
TB_GOLANGCI_LINT_VERSION_NUM ?= $(call STRIP_V,$(TB_GOLANGCI_LINT_VERSION))
# renovate: packageName=github.com/goreleaser/goreleaser/v2
TB_GORELEASER_VERSION ?= v2.11.2
# renovate: packageName=github.com/bakito/semver
TB_SEMVER_VERSION ?= v1.1.5

## Tool Installer
.PHONY: tb.deepcopy-gen
tb.deepcopy-gen: ## Download deepcopy-gen locally if necessary.
	@test -s $(TB_DEEPCOPY_GEN) || \
		GOBIN=$(TB_LOCALBIN) go install k8s.io/code-generator/cmd/deepcopy-gen@$(TB_DEEPCOPY_GEN_VERSION)
.PHONY: tb.ginkgo
tb.ginkgo: ## Download ginkgo locally if necessary.
	@test -s $(TB_GINKGO) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo@$(TB_GINKGO_VERSION)
.PHONY: tb.golangci-lint
tb.golangci-lint: ## Download golangci-lint locally if necessary.
	@test -s $(TB_GOLANGCI_LINT) && $(TB_GOLANGCI_LINT) --version | grep -q $(TB_GOLANGCI_LINT_VERSION_NUM) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(TB_GOLANGCI_LINT_VERSION)
.PHONY: tb.goreleaser
tb.goreleaser: ## Download goreleaser locally if necessary.
	@test -s $(TB_GORELEASER) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/goreleaser/goreleaser/v2@$(TB_GORELEASER_VERSION)
.PHONY: tb.semver
tb.semver: ## Download semver locally if necessary.
	@test -s $(TB_SEMVER) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/bakito/semver@$(TB_SEMVER_VERSION)

## Reset Tools
.PHONY: tb.reset
tb.reset:
	@rm -f \
		$(TB_DEEPCOPY_GEN) \
		$(TB_GINKGO) \
		$(TB_GOLANGCI_LINT) \
		$(TB_GORELEASER) \
		$(TB_SEMVER)

## Update Tools
.PHONY: tb.update
tb.update: tb.reset
	toolbox makefile --renovate -f $(TB_LOCALDIR)/Makefile \
		k8s.io/code-generator/cmd/deepcopy-gen@github.com/kubernetes/code-generator \
		github.com/onsi/ginkgo/v2/ginkgo \
		github.com/golangci/golangci-lint/v2/cmd/golangci-lint?--version \
		github.com/goreleaser/goreleaser/v2 \
		github.com/bakito/semver
## toolbox - end
