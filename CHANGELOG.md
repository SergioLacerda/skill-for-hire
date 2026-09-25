# Changelog

Todas as mudanças relevantes deste projeto são registradas aqui, seguindo
[Keep a Changelog](https://keepachangelog.com/) e [SemVer](https://semver.org/).

## [Unreleased]

### Added

- CLI `skillhire` com comandos `validate`, `pack` e `version`.
- Package-static validation contra o schema `skillsforhire.dev/v1`:
  kebab-case, SemVer, entrypoint existente, deny-by-default de
  capabilities.
- Empacotamento reprodutível: `<name>-<version>.tar.gz` com layout ORKA
  e `.sha256` companion, digest determinístico entre runs.
- Makefile e submakes (`go.mk`, `quality.mk`, `release.mk`).
- Configuração `.goreleaser.yaml` v2 para binários `skillhire`
  multi-plataforma (linux/darwin/windows × amd64/arm64) + upload dos
  skill packs como assets do release.
- GitHub Actions: CI (`lint`, `test`, `release-dry-run`) e Release
  (tag `v*.*.*` publica via goreleaser).
- Configuração `.golangci.yaml` mínima (errcheck, gosec, staticcheck,
  govet, revive, misspell, ineffassign, unconvert).
