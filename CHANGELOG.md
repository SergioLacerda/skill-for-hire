# Changelog

Todas as mudanças relevantes deste projeto são registradas aqui, seguindo
[Keep a Changelog](https://keepachangelog.com/) e [SemVer](https://semver.org/).

## [Unreleased]

### Added — Waves 1-4 (pre-`v0.0.2` polish)

- **Slim skill packs** — `internal/pack` filtra `*_test.go` por padrão
  (opção `IncludeTests`), treasure-chest cai de 68 KB → 44 KB.
- **Cross-compile `treasure-chest`** no `.goreleaser.yaml` — binários
  para linux/darwin/windows × amd64/arm64 shippam como asset da release.
- **`docs/consumer-guide.md`** atualizado para `v0.0.2`: seção nova
  documentando o binário standalone.
- **Testes do CLI `cmd/treasure-chest/`** cobrindo version, status
  human+JSON, prepare-vazio, search em workspace vazio, explain de ID
  desconhecido e prepare contra layout mínimo de chest.
- **Boundary architectural via depguard** (`.golangci.yaml`): deny de
  qualquer import `github.com/SergioLacerda/strategist-skill`;
  substitui o `architecture_test.go` do fonte que dependia de
  `internal/conformance`.
- **Security review pass** — 0 findings HIGH/MEDIUM contra o diff
  (transcript no chat da sessão).
- **Ranking real** (`runtime/ranking.go`): substring naive substituído
  por ranker explicável com pesos por campo (Statement +3, Scope +2,
  AppliesWhen +2, Kind +1, AvoidWhen -4) e tie-breaker por
  curatorial score. `TokenBudget` agora aplicado após ranking, não
  antes.
- **Refresh com delta real** — `Refresh` compara assinaturas
  `fmt.Sprintf` antes e depois do reindex e devolve
  `added/updated/removed` acurados; limitations flag deixa claro que
  `Scope.Since`/`Scope.Paths` ainda são full reindex.
- **Adapter Strategist executável** (`adapters/strategist/`):
  `Router` resolve por capability, `Piloted` wrapper aplica `Policy`
  (stale reads, degradação) + `TelemetrySink`, satisfazendo o mesmo
  `KnowledgeProvider`. Constantes `PilotCartografo`/`PilotJewelcrafter`.

### Added — Batch 4 (treasure-chest runtime import)

- **Domínio treasure-chest migrado** do repositório `strategist-skill`:
  72 arquivos Go, ~14.810 LOC (2 pacotes: `runtime` + `runtime/domain`)
  em `skills/treasure-chest/runtime/`. O acoplamento residual com
  `internal/domain` do fonte foi substituído por
  `runtime/domain/vocab.go` (22 linhas, 7 constantes). Zero outro
  acoplamento com o strategist-skill.
- **Adapter Knowledge API** (`runtime/adapter.go`) implementa
  `KnowledgeProvider` end-to-end (Prepare/Search/Refresh/Status/
  Explain) sobre os loaders existentes do domínio, com envelope
  completo em toda resposta.
- **CLI standalone** `cmd/treasure-chest/` — subcomandos `prepare`,
  `search`, `status`, `explain`, `version`; flag `--json` universal.
- **`treasure-chest` bumped para `0.2.0`**; skillhire.lock regenerado.
- **Pack estendido**: `internal/pack/pack.go` inclui `runtime/` e
  `contracts/` no default include ORKA + Skills-for-Hire.
- Nova dependência: `github.com/stretchr/testify` (via go.mod tidy).

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
