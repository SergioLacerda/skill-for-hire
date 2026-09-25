# Changelog

Todas as mudanças relevantes deste projeto são registradas aqui, seguindo
[Keep a Changelog](https://keepachangelog.com/) e [SemVer](https://semver.org/).

## [Unreleased]

### Added — Batch 3 (Consumer verify + quickstart)

- **CLI** `skillhire verify <archive.tar.gz>...` — cross-check digest
  do arquivo contra `.sha256` companion e (quando presente) contra a
  release manifest sidecar (`digest`, `size`). Aceita `.sha256` no
  formato coreutils (`digest  filename`) e bare-digest; recusa arquivos
  cujo companion referencia outro basename.
- **`internal/verify`** com cobertura para: pack válido, arquivo
  adulterado, manifest adulterado, `.sha256` spoofado com nome
  incorreto, bare digest, e ausência de manifest.
- **README raiz** ganhou seção "Consumindo um release" com fluxos
  download → verify → extract (com e sem CLI) e nota sobre attestations
  via `gh attestation verify`.

### Added — Batch 2 (Lockfile + provenance)

- **Lockfile** `skillhire.lock` (schema `skillsforhire.dev/lockfile v1`):
  consumer-side pin de skills por versão, digest sha256, size e source.
  Renderização determinística (sort estável por nome, sem timestamps).
- **CLI** `skillhire lock <release.yaml>... --out skillhire.lock` — lê
  os sidecars produzidos por `skillhire pack` e emite/verifica o
  lockfile.
- **`--verify`** compara apenas o mapa `installed` (ignora `generator`,
  que naturalmente varia entre local/CI/tag), com diff human-readable.
- **`doctrine/lockfile.schema.json`** define o formato canônico.
- **Make targets**: `lock-skills`, `lock-verify`; `ci-test` agora inclui
  `lock-verify` como gate.
- **Release workflow** ganha SBOM CycloneDX (`anchore/sbom-action`) e
  duas etapas de `attest-build-provenance`: uma para os binários
  (via `dist/SHA256SUMS`) e outra para os skill packs em `packs/`.
  Permissões `id-token: write` e `attestations: write` habilitadas.

### Added

- CLI `skillhire` com comandos `validate`, `pack` e `version`.
- Package-static validation contra o schema `skillsforhire.dev/v1`:
  kebab-case, SemVer, entrypoint existente, deny-by-default de
  capabilities.
- Empacotamento reprodutível: `<name>-<version>.tar.gz` com layout ORKA
  e `.sha256` companion, digest determinístico entre runs.
- **Bundle manifest sidecar** `<name>-<version>.release.yaml` com
  `schema_version`, `name`, `version`, `archive`, `digest`, `size`,
  `generator`, `source_commit` e `contents[]` (§5.5 do doc de
  arquitetura).
- **Knowledge API v1** provider-neutral em `platform/knowledge-api/`
  (Go, stdlib only): interface `KnowledgeProvider` (Prepare / Search /
  Refresh / Status / Explain), `Envelope` com provenance/freshness/
  trust/fallback, erros normalizados (`ErrNotPrepared`,
  `ErrIncompatibleVersion`, `ErrPermissionDenied`, `ErrSourceMissing`,
  `ErrStale`, `ErrFallbackTriggered`, `ErrUnavailable`), enums para
  Freshness/FallbackState/Decision/ScopeFilter.
- Makefile e submakes (`go.mk`, `quality.mk`, `release.mk`).
- Configuração `.goreleaser.yaml` v2 para binários `skillhire`
  multi-plataforma (linux/darwin/windows × amd64/arm64) + upload dos
  skill packs como assets do release.
- GitHub Actions: CI (`lint`, `test`, `release-dry-run`) e Release
  (tag `v*.*.*` publica via goreleaser).
- Configuração `.golangci.yaml` mínima (errcheck, gosec, staticcheck,
  govet, revive, misspell, ineffassign, unconvert).
