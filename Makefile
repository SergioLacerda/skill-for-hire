# Skills for Hire — top-level Makefile.
#
# Structure mirrors strategist-skill's split-by-concern layout: the root
# file wires the CI aggregates; per-topic recipes live under make/.

.PHONY: ci ci-lint ci-test help

# Fail fast with a clear message when make cannot reach a POSIX shell
# (typically PowerShell/cmd.exe on Windows). Without this the recipe lines
# below error one at a time with opaque messages.
ifneq ($(shell test -z "" && echo posix-shell-ok),posix-shell-ok)
$(error make requires a POSIX shell. On Windows, run make from Git Bash or WSL.)
endif

# Isolated caches so restricted environments (CI, sandboxes) stay writable
# and never reuse entries written by a different toolchain version.
GOCACHE              ?= /tmp/go-build-cache
GOMODCACHE           ?= /tmp/go-mod-cache
GOLANGCI_LINT_CACHE  ?= /tmp/golangci-lint-cache

# `go env GOPATH` prints a backslash path on Windows. tr normalizes it once
# here so nothing downstream ever sees a backslash to mis-parse.
GOPATH_BIN           := $(shell go env GOPATH | tr '\134' '/')/bin
LOCAL_BIN            := $(CURDIR)/bin

# Executable suffix of the build host (".exe" on Windows).
EXE                  ?= $(shell go env GOEXE)
SKILLHIRE_BIN        := bin/skillhire$(EXE)

DIST_DIR             ?= dist
SKILLS_DIR           ?= skills
# Discovered at recipe time so a new skills/<name>/ shows up without a
# manual list update.
SKILL_DIRS            = $(shell ls -d $(SKILLS_DIR)/*/ 2>/dev/null | sed 's:/$$::')

GORELEASER           := $(shell which goreleaser 2>/dev/null || echo $(GOPATH_BIN)/goreleaser)
GORELEASER_VERSION   ?= v2.4.4

GOLANGCI_LINT        := $(shell which golangci-lint 2>/dev/null || echo $(GOPATH_BIN)/golangci-lint)

include make/go.mk
include make/quality.mk
include make/release.mk

ci-lint: fmt-check mod-check vet build
ci-test: test validate-skills pack-skills

ci: ci-lint ci-test

help:
	@echo "Skills for Hire — common targets"
	@echo "  make build            Build the skillhire binary into $(SKILLHIRE_BIN)"
	@echo "  make test             Run Go unit tests"
	@echo "  make lint             golangci-lint over the module"
	@echo "  make fmt              Apply gofmt to tracked Go files"
	@echo "  make validate-skills  Run 'skillhire validate' on every skills/<name>/"
	@echo "  make pack-skills      Produce dist/<skill>-<version>.tar.gz + .sha256"
	@echo "  make snapshot         Local release via goreleaser (no publish)"
	@echo "  make release-check    Validate the goreleaser configuration"
	@echo "  make clean            Remove build and dist artifacts"
	@echo "  make ci               ci-lint + ci-test aggregate"
