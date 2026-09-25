---
name: atlas
description: Use esta skill para mapear e pré-indexar a arquitetura e a estrutura documental de um projeto — módulos, dependências, entrypoints, símbolos, testes, documentação e ADRs — reduzindo o custo de descoberta em missões subsequentes. Use quando o consumidor precisar saber onde estão as coisas e como se relacionam. Não use para decidir a solução de uma missão, para interpretar causalidade histórica (Pathfinder) ou para curar conhecimento reutilizável (Treasure Chest).
metadata:
  version: 0.1.0
  author: skill-for-hire
---

# Atlas

## Objetivo

Fornecer uma cartografia estrutural e documental do projeto para que
consumidores (missões, orquestradores, agentes) descubram rapidamente
onde ler antes de ler.

O Atlas é um **cache de navegação e relações**, não uma fonte canônica.
Código, documentação e ADRs permanecem a fonte de verdade.

## Quando usar

- Bootstrap de uma nova missão que precisa entender módulos, entrypoints
  ou relações de dependência.
- Consulta pontual: "onde está X?", "quem depende de Y?", "quais testes
  cobrem Z?".
- Descoberta e indexação de ADRs de um projeto cliente.
- Análise de impacto estrutural de uma mudança candidata.

## Quando não usar

- Para escolher a solução de uma missão (papel do orquestrador).
- Para interpretar história e causalidade (papel do Pathfinder).
- Para curar, promover ou selecionar conhecimento reutilizável (papel da
  Treasure Chest).
- Como substituto de Git, ADRs, testes ou documentação canônica.

## Entradas necessárias

- Raiz do projeto (workspace root) com permissão de leitura.
- Opcional: intent, componentes, conceitos ou faixa de commits para
  operações de search e impact.

## Procedimento

1. **Prepare** — descobre fontes, calcula digests, constrói o índice
   arquitetural e documental vinculado ao commit atual.
2. **Search** — responde consultas retornando candidatos ranqueados
   (paths, símbolos, conceitos, ADRs) com evidência e razão de seleção.
3. **Refresh** — recalcula seletivamente o índice a partir de um diff
   contra o commit representado.
4. **Status** — reporta commit indexado, freshness, saúde e capabilities
   ativas.

Em qualquer operação, o Atlas deve devolver primeiro **índices, resumos
e razões**; conteúdo completo é carregado sob demanda pelo consumidor.

## Critérios de conclusão

- Toda resposta cita fontes com path, digest e commit.
- Freshness (`fresh` | `partially_stale` | `stale`) é sempre reportada.
- Candidatos vêm ranqueados com razão explicável.

## Saídas

- Envelope de resposta com `provider`, `version`, `capabilities_used`,
  `freshness`, `trust`, `sources[]`, `limitations`, `fallback_state`.
- Projeções de ADR contendo path, digest, `decision_status`,
  `index_freshness`, `implementation_alignment`.
- Impact analysis: paths, símbolos, testes e docs afetados por um diff.

## Limites e escalonamento

- Filesystem: leitura somente na raiz autorizada.
- Rede: negada por padrão.
- Se o índice estiver `stale` ou `incompatible`, o Atlas deve **explicar**
  e não simular resposta indexada.
- ADRs pertencem ao repositório cliente; o Atlas indexa, não copia.

## Referências sob demanda

- Leia `references/capabilities.md` para o contrato das operações
  publicadas.
- Leia `references/lifecycle.md` para estados do índice (`FRESH`,
  `PARTIALLY_STALE`, `STALE`, `INCOMPATIBLE`, `MISSING`) e política de
  atualização incremental.
