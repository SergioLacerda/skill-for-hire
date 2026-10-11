.PHONY: vuln secrets security install-govulncheck install-gitleaks

# Security gates. Each one is a Makefile target so a developer and CI run the
# exact same command; the workflows only call `make <target>`. Tool versions
# are pinned here and installed on demand, never taken from whatever happens
# to be on PATH.

GOVULNCHECK_VERSION ?= v1.8.0
GITLEAKS_VERSION    ?= v8.30.1

GOVULNCHECK         := $(shell which govulncheck 2>/dev/null || echo $(GOPATH_BIN)/govulncheck)
GITLEAKS            := $(shell which gitleaks 2>/dev/null || echo $(GOPATH_BIN)/gitleaks)

# Bounded retry, same shape as install-goreleaser: module-proxy hiccups
# occasionally drop large downloads mid-stream.
install-govulncheck:
	@for attempt in 1 2 3; do \
	  GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) && exit 0; \
	  echo "install-govulncheck: attempt $$attempt failed" >&2; \
	  [ "$$attempt" -lt 3 ] && sleep $$((attempt * 10)); \
	done; \
	echo "install-govulncheck: failed after 3 attempts" >&2; exit 1

# The module path declared by gitleaks' go.mod is zricethezav/gitleaks, not
# gitleaks/gitleaks, even for current releases.
install-gitleaks:
	@for attempt in 1 2 3; do \
	  GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go install github.com/zricethezav/gitleaks/v8@$(GITLEAKS_VERSION) && exit 0; \
	  echo "install-gitleaks: attempt $$attempt failed" >&2; \
	  [ "$$attempt" -lt 3 ] && sleep $$((attempt * 10)); \
	done; \
	echo "install-gitleaks: failed after 3 attempts" >&2; exit 1

# vuln reports known vulnerabilities that the module's code actually reaches.
# It does not replace dependency review: it only sees what the build graph
# calls, and only what the Go vulnerability database already lists.
vuln: install-govulncheck
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" "$(GOVULNCHECK)" ./...

# secrets scans the git history for committed credentials. --redact keeps
# matched values out of CI logs. Findings that are accepted must be recorded
# in .gitleaksignore with a dated comment naming who accepted them and why.
secrets: install-gitleaks
	"$(GITLEAKS)" detect --source . --redact --no-banner --exit-code 1

security: vuln secrets
