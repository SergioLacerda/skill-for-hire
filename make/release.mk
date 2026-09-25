.PHONY: pack-skills snapshot release-check release-dry-run release install-goreleaser clean

# pack-skills produces dist/<skill>-<version>.tar.gz + .sha256 + .release.yaml
# for each discovered skill. This is what the release workflow uploads
# next to the skillhire binary produced by goreleaser.
pack-skills: build
	@if [ -z "$(SKILL_DIRS)" ]; then \
		echo "pack-skills: no skills found under $(SKILLS_DIR)/" >&2; \
		exit 1; \
	fi
	./$(SKILLHIRE_BIN) pack $(SKILL_DIRS) --out "$(DIST_DIR)"

# lock-skills builds skillhire.lock from every release.yaml sidecar in
# DIST_DIR/. Depends on pack-skills so the sidecars exist first.
LOCK_FILE ?= skillhire.lock
lock-skills: pack-skills
	./$(SKILLHIRE_BIN) lock $(DIST_DIR)/*.release.yaml --out "$(LOCK_FILE)"

# lock-verify fails when skillhire.lock differs from what the current
# release.yaml sidecars would produce — the CI gate that catches an
# unmerged pack without a matching lockfile bump.
lock-verify: pack-skills
	./$(SKILLHIRE_BIN) lock $(DIST_DIR)/*.release.yaml --out "$(LOCK_FILE)" --verify

# install-goreleaser pins the version. Module-proxy hiccups occasionally
# drop large downloads mid-stream; a bounded retry keeps CI quiet.
install-goreleaser:
	@for attempt in 1 2 3; do \
	  GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION) && exit 0; \
	  echo "install-goreleaser: attempt $$attempt failed" >&2; \
	  [ "$$attempt" -lt 3 ] && sleep $$((attempt * 10)); \
	done; \
	echo "install-goreleaser: failed after 3 attempts" >&2; exit 1

release-check:
	"$(GORELEASER)" check

# snapshot builds release artifacts locally without publishing (no token).
snapshot:
	"$(GORELEASER)" release --snapshot --clean --skip=publish

release-dry-run: install-goreleaser release-check snapshot pack-skills

# release publishes to GitHub — CI provides GITHUB_TOKEN via the workflow.
release:
	"$(GORELEASER)" release --clean

clean:
	rm -rf bin/ dist/ packs/
