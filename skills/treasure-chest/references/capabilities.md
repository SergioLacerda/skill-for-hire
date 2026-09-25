# treasure-chest — capabilities

Contrato das operações publicadas. Todas seguem o envelope padrão do
`KnowledgeProvider`.

## `knowledge.mine`

Extrai itens candidatos das fontes autorizadas.

**Entrada**

```yaml
scope:
  root: <path>
  sources: [<glob> | <path>]
  since: <sha>            # opcional, diff de fontes
policy:
  trust_floor: <float>
  scope: global | product | workspace
```

**Saída**

```yaml
candidates:
  - id: <string>
    kind: jewel | potion | runbook | playbook
    source:
      path: <string>
      digest: sha256:...
    evidence: [...]
    provenance: {...}
```

## `knowledge.search`

Recupera itens aplicáveis a uma consulta de missão.

**Entrada**

```yaml
query:
  intent: <string>
  concepts: [...]
  components: [...]
  token_budget: <int>
```

**Saída**

```yaml
items:
  - id: <string>
    kind: jewel | potion | runbook | playbook
    score: <float>
    trust: <float>
    freshness: fresh | stale | unknown
    applicability: <string>
    source: {...}
```

## `knowledge.curate`

Deduplica, classifica e aplica trust/scope. Devolve deltas curatoriais.

## `runbook.select`

Seleciona runbook/playbook aplicável a um cenário. Ordem de resolução:

1. item local fixado pelo usuário;
2. item específico do produto;
3. item global compatível;
4. fallback embarcado;
5. discovery convencional;
6. geração de candidato.

## `learning.reuse`

Aplica lições e precedentes a missões correntes com evidência.

## Estados curatoriais

```text
candidate → reviewed → accepted → deprecated
```

**Invariante:** promoção de `candidate` para `accepted` exige gate
(política + revisão humana por padrão).
