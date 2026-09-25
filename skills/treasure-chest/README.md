# treasure-chest

Baú de conhecimento reutilizável — mineração, curadoria, lifecycle e
recuperação de Jewels, Potions, runbooks, playbooks e aprendizados.

## Status

`experimental` — o esqueleto está aqui como template. O conteúdo
funcional será **exportado de outro projeto** existente e migrado
preservando o padrão ORKA + Skills for Hire.

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

## Migração (a fazer)

- [ ] Importar domínio (Jewels, Potions, runbooks) do projeto de origem.
- [ ] Congelar comportamento atual em fixtures de paridade.
- [ ] Adicionar import idempotente de dados existentes.
- [ ] Publicar adapter Strategist em `adapters/strategist/`.
- [ ] Validar fallback e rollback.
