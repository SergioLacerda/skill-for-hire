#!/usr/bin/env bash
# Pack every skill under skills/*/ into dist/ using the freshly built
# skillhire binary. Called from goreleaser's before hook and from
# `make pack-skills`.
set -euo pipefail

DIST_DIR="${DIST_DIR:-dist}"
SKILLS_DIR="${SKILLS_DIR:-skills}"
SKILLHIRE_BIN="${SKILLHIRE_BIN:-bin/skillhire}"

if [ ! -x "$SKILLHIRE_BIN" ]; then
  echo "pack-skills: $SKILLHIRE_BIN not found or not executable; run 'go build -o $SKILLHIRE_BIN ./cmd/skillhire' first" >&2
  exit 1
fi

shopt -s nullglob
skill_dirs=("$SKILLS_DIR"/*/)
if [ ${#skill_dirs[@]} -eq 0 ]; then
  echo "pack-skills: no skills found under $SKILLS_DIR/" >&2
  exit 1
fi

# Strip trailing slashes for cleaner CLI output.
targets=()
for d in "${skill_dirs[@]}"; do
  targets+=("${d%/}")
done

mkdir -p "$DIST_DIR"
"$SKILLHIRE_BIN" pack "${targets[@]}" --out "$DIST_DIR"
