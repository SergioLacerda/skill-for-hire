# atlas — capabilities

Contrato das operações publicadas pela skill. Todas seguem o envelope
padrão do `KnowledgeProvider`.

## `architecture.map`

Devolve o mapa arquitetural do projeto vinculado ao commit atual.

**Entrada**

```yaml
scope:
  root: <path>
  commit: <sha>          # opcional, default HEAD
  depth: <int>           # opcional
```

**Saída**

```yaml
provider: atlas
version: "0.1.0"
commit: <sha>
freshness: fresh | partially_stale | stale
modules: [...]
entrypoints: [...]
languages: [...]
docs: [...]
```

## `architecture.search`

Consulta por paths, símbolos ou conceitos. Devolve candidatos ranqueados
com razão explicável.

**Entrada**

```yaml
query:
  intent: <string>
  concepts: [...]
  symbols: [...]
  paths: [...]
  token_budget: <int>
```

**Saída**

```yaml
items:
  - id: <string>
    kind: path | symbol | concept | adr
    score: <float>
    reason: <string>
    source:
      path: <string>
      digest: sha256:...
```

## `architecture.impact`

Análise de impacto estrutural de um diff.

**Entrada**

```yaml
diff:
  from: <sha>
  to: <sha>
```

**Saída**

```yaml
affected:
  paths: [...]
  symbols: [...]
  tests: [...]
  docs: [...]
  concepts: [...]
```

## `adr.discover` / `adr.relate`

Descobre ADRs presentes no projeto e as relaciona a componentes.

Cada projeção carrega:

```yaml
kind: adr_projection
id: <adr-id>
source:
  path: <string>
  digest: sha256:...
decision_status: proposed | accepted | rejected | superseded
summary: <string>
relations:
  affects: [...]
index_freshness: fresh | stale | unknown
implementation_alignment: aligned | drifted | unknown
```

**Invariante:** o Atlas nunca altera `decision_status` automaticamente.
TTL pode disparar reverificação, não mudança decisória.
