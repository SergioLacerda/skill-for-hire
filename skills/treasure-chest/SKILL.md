---
name: treasure-chest
description: Use esta skill para transformar experiências, documentos e evidências em conhecimento selecionável, rastreável e reutilizável — Jewels, Potions, runbooks e playbooks — com trust, freshness, escopo e provenance. Use para minerar candidatos, curar, deduplicar e recuperar conhecimento aplicável a uma missão. Não use para mapear arquitetura (Atlas), interpretar história (Pathfinder) ou executar a pipeline do orquestrador.
metadata:
  version: 0.1.0
  author: skill-for-hire
---

# Treasure Chest

## Objetivo

Manter um baú curado de conhecimento reutilizável, com lifecycle próprio
e recuperação seletiva. Provê aos consumidores precedentes, runbooks e
lições aplicáveis a uma missão, sempre com proveniência.

## Quando usar

- Recuperar precedentes e lições reutilizáveis para uma missão.
- Selecionar runbooks/playbooks aplicáveis a um cenário.
- Minerar candidatos a partir de fontes (ADRs, PRs, missões passadas).
- Deduplicar, curar e aplicar TTL a itens conhecidos.

## Quando não usar

- Para mapear arquitetura, dependências ou ADRs (papel do Atlas).
- Para interpretar causalidade histórica (papel do Pathfinder).
- Como fonte canônica de código, testes ou documentação do projeto.
- Como mecanismo de execução da pipeline do orquestrador.

## Entradas necessárias

- Raiz do projeto e fontes autorizadas para mineração (read-only).
- Sinais de missão para consultas: intent, componentes, conceitos.
- Política de trust, TTL e escopo (global / product / workspace).

## Procedimento

1. **Prepare** — inventaria fontes autorizadas, snapshot e digest.
2. **Mine** — extrai itens candidatos com provenance completa.
3. **Curate** — deduplica, classifica, aplica trust e escopo.
4. **Search / Select** — devolve itens ranqueados por aplicabilidade,
   compatibilidade, trust e freshness.
5. **Refresh** — reavalia TTL, detecta drift, propõe reindexação.

## Critérios de conclusão

- Todo item devolvido carrega `source`, `digest`, `trust`, `freshness`,
  `scope` e `owner`.
- Candidatos **nunca** são promovidos automaticamente a itens aceitos.
- Conflitos são explicitados, não silenciados.

## Saídas

- Envelope padrão do `KnowledgeProvider`.
- Itens tipados: `jewel`, `potion`, `runbook`, `playbook`, `candidate`.
- Estados curatoriais: `candidate → reviewed → accepted → deprecated`.

## Limites e escalonamento

- Filesystem: leitura somente das fontes autorizadas.
- Rede: negada por padrão.
- Escrita em fonte canônica requer permissão explícita e gate.
- Promoção de candidato exige política e revisão humana por padrão.

## Referências sob demanda

- Leia `references/capabilities.md` para o contrato de operações.
- Leia `references/lifecycle.md` para estados curatoriais e TTL.

## Nota de origem

Esta skill será **exportada de outro projeto** e incorporada aqui como
package independente. Este arquivo é o esqueleto do contrato portátil —
o conteúdo instrucional completo (mineração, curadoria, ranking) será
migrado preservando o padrão ORKA + Skills for Hire.
