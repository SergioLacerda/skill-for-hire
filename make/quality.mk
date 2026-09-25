.PHONY: lint lint-fix

# lint is diagnostic-only; it never rewrites source.
lint:
	@if [ ! -x "$(GOLANGCI_LINT)" ] && ! command -v "$(GOLANGCI_LINT)" >/dev/null 2>&1; then \
		echo "lint: golangci-lint not installed; install v2.x or set GOLANGCI_LINT=<path>" >&2; \
		exit 1; \
	fi
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOLANGCI_LINT_CACHE="$(GOLANGCI_LINT_CACHE)" \
		"$(GOLANGCI_LINT)" run ./...

lint-fix:
	git ls-files -co --exclude-standard -z '*.go' | xargs -0r gofmt -w
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOLANGCI_LINT_CACHE="$(GOLANGCI_LINT_CACHE)" \
		"$(GOLANGCI_LINT)" run --fix ./...
