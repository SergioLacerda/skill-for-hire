.PHONY: fmt fmt-check mod-tidy mod-check vet build test bench validate-skills

fmt:
	git ls-files -co --exclude-standard -z '*.go' | xargs -0r gofmt -w

fmt-check:
	@files="$$(git ls-files -co --exclude-standard -z '*.go' | xargs -0r gofmt -l)"; \
	if [ -n "$$files" ]; then \
		echo "fmt-check: unformatted Go files detected:" >&2; \
		printf '%s\n' "$$files" | sed 's/^/  - /' >&2; \
		echo "fmt-check: run 'make fmt' to apply formatting" >&2; \
		exit 1; \
	else \
		echo "fmt-check: all Go files are formatted"; \
	fi

mod-tidy:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go mod tidy

mod-check:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go mod tidy -diff
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go mod verify

vet:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go vet ./...

build:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go build \
		-ldflags="-s -w -X main.Version=$$(git describe --tags --dirty --always 2>/dev/null || echo dev)" \
		-o "$(SKILLHIRE_BIN)" ./cmd/skillhire
	@# Also compile every other cmd/ binary (treasure-chest, future
	@# skill runtimes). Ensures a broken skill CLI fails CI, not just
	@# skillhire drift.
	@for cmddir in $$(ls -d cmd/*/ 2>/dev/null | sed 's:/$$::' | grep -v '^cmd/skillhire$$'); do \
		bin="bin/$$(basename $$cmddir)$(EXE)"; \
		echo "building $$bin"; \
		GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go build \
			-ldflags="-s -w -X main.Version=$$(git describe --tags --dirty --always 2>/dev/null || echo dev)" \
			-o "$$bin" "./$$cmddir" || exit 1; \
	done

test:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go test -race ./...

bench:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go test -bench=. -benchmem ./...

# validate-skills runs the CLI on every discovered skill dir. A missing
# SKILL_DIRS list is not a silent pass — the target fails loudly so an
# empty repo does not ship as green CI.
validate-skills: build
	@if [ -z "$(SKILL_DIRS)" ]; then \
		echo "validate-skills: no skills found under $(SKILLS_DIR)/" >&2; \
		exit 1; \
	fi
	./$(SKILLHIRE_BIN) validate $(SKILL_DIRS)
