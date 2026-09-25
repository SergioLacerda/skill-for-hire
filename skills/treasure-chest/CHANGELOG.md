# Changelog — treasure-chest

Todas as mudanças relevantes desta skill são registradas aqui, seguindo
[Keep a Changelog](https://keepachangelog.com/) e [SemVer](https://semver.org/).

## [Unreleased]

## [0.2.0] — 2026-09-25

### Added

- **Runtime importado** do projeto strategist-skill: 72 arquivos Go
  (~14.810 LOC) sob `runtime/` — add, index, jewel_*, potion_*,
  loaders, scoring, chest_remove, compiled_index, evidence quality e
  domain types (chest_grade, jewel_grade, potion_grade,
  promotion_packet, validation_sets).
- **Adapter Knowledge API** (`runtime/adapter.go`) implementando
  `KnowledgeProvider` (Prepare / Search / Refresh / Status / Explain)
  com envelope completo. Search é substring-match nesta versão;
  ranking via `ScoringPolicy` do domínio fica para batch futuro.
- **CLI standalone** `treasure-chest` (em `cmd/treasure-chest/` do
  repo): subcomandos `prepare`, `search`, `status`, `explain`,
  `version`, com flag `--json` para automação.
- **`runtime/domain/vocab.go`** — 22 linhas replicando as 7 constantes
  (Confidence*, EvidenceClass*) que eram a única dependência residual
  com `internal/domain` do strategist-skill. Zero outro acoplamento.

### Changed

- Pack agora inclui `runtime/` e `contracts/` como diretórios canônicos
  do layout ORKA + extensões Skills-for-Hire.

## [0.1.0] — 2026-09-25

- Bootstrap do template (hello world).
