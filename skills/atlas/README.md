# atlas

Cartografia estrutural e documental de projetos — a specialist responsável
por mapear módulos, dependências, entrypoints, símbolos, testes,
documentação e ADRs de um repositório e responder consultas com
evidência.

## Status

`experimental` — apenas o esqueleto do template está definido. A
implementação será feita em fases (ver
`docs/architecture/05-atlas-cartographer.md` para o roadmap conceitual).

## Capabilities publicadas

| Capability | Propósito |
| --- | --- |
| `architecture.map` | Devolve o mapa arquitetural do projeto no commit atual. |
| `architecture.search` | Consulta por paths, símbolos ou conceitos. |
| `architecture.impact` | Análise de impacto estrutural de um diff. |
| `adr.discover` | Descobre e indexa ADRs presentes no projeto. |
| `adr.relate` | Relaciona ADRs a componentes, símbolos e mudanças. |

## Interface conceitual

Segue o contrato `KnowledgeProvider` da plataforma:

```text
Prepare(ctx, request)    → PrepareResult
Search(ctx, query)       → []KnowledgeItem
Refresh(ctx, scope)      → RefreshResult
Status(ctx)              → ProviderStatus
Explain(ctx, item_id)    → Explanation
```

Todo envelope de resposta carrega: `provider`, `version`,
`capabilities_used`, `freshness`, `trust`, `sources[]`, `limitations`,
`fallback_state`.

## Standalone

Atlas deve ser executável sem Strategist. O Cartógrafo (piloto do
Strategist) é apenas um consumidor privilegiado.

## Instalação

```bash
skillhire install atlas
```

*(CLI ainda não implementada.)*

## Estrutura

```text
atlas/
├── SKILL.md           # contrato instrucional portátil
├── skill.yaml         # manifesto operacional
├── README.md          # este arquivo
├── CHANGELOG.md
└── references/
    ├── capabilities.md
    └── lifecycle.md
```
