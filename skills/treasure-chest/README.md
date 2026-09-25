# treasure-chest

Baú de conhecimento reutilizável — mineração, curadoria, lifecycle e
recuperação de Jewels, Potions, runbooks, playbooks e aprendizados.

## Status

`experimental` — versão `0.2.0` traz o **runtime importado do projeto
strategist-skill** (72 arquivos Go, ~14k LOC) sob `runtime/`,
implementando o contrato `platform/knowledge-api.KnowledgeProvider`.
O acoplamento com `internal/*` do projeto-fonte foi eliminado (só um
`domain/vocab.go` local de 22 linhas replicando 7 constantes de string).

## Capabilities publicadas

| Capability | Propósito |
| --- | --- |
| `knowledge.mine` | Extrai candidatos de fontes autorizadas com provenance. |
| `knowledge.search` | Recupera itens aplicáveis a uma consulta de missão. |
| `knowledge.curate` | Deduplica, classifica e aplica trust/scope. |
| `runbook.select` | Seleciona runbooks/playbooks compatíveis. |
| `learning.reuse` | Aplica lições e precedentes a missões correntes. |

## Interface conceitual

Contrato `KnowledgeProvider` da plataforma:

```text
Prepare(ctx, request)    → PrepareResult
Search(ctx, query)       → []KnowledgeItem
Refresh(ctx, scope)      → RefreshResult
Status(ctx)              → ProviderStatus
Explain(ctx, item_id)    → Explanation
```

## Integração com Atlas

Treasure Chest **pode** consumir projeções do Atlas (ADRs, mapa
arquitetural) para orientar mineração, mas não depende obrigatoriamente
dele. Ambos podem operar isolados.

## Standalone

Executável sem Strategist. O Jewelcrafter (piloto do Strategist) é
apenas um consumidor privilegiado.

## Instalação

```bash
skillhire install treasure-chest
```

*(CLI ainda não implementada.)*

## Uso standalone

O binário `treasure-chest` (em `cmd/treasure-chest/` do repositório) é
o protocolo standalone da skill:

```bash
treasure-chest prepare --root /path/to/workspace
treasure-chest search  --root /path/to/workspace --intent "dependency upgrade"
treasure-chest status  --root /path/to/workspace [--json]
treasure-chest explain <jewel-id> --root /path/to/workspace
```

Todas as respostas carregam o `Envelope` da Knowledge API v1
(`provider`, `version`, `schema_version`, `capabilities_used`,
`freshness`, `trust`, `fallback_state`, `limitations`).

## Migração (concluída em 0.2.0)

- [x] Importar domínio (Jewels, Potions, runbooks) do projeto de origem.
- [x] Desacoplar de `internal/*` do strategist-skill.
- [x] Implementar `KnowledgeProvider` (Prepare/Search/Refresh/Status/Explain).
- [x] Expor CLI standalone.
- [ ] Ranking real usando `ScoringPolicy` do domínio (batch futuro).
- [ ] Refresh incremental por diff (batch futuro).
- [ ] Adapter Strategist real em `adapters/strategist/` (batch futuro).
