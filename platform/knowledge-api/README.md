# platform/knowledge-api

Contrato compartilhado pelas skills de conhecimento (Atlas, Treasure
Chest e futuros providers). Este é o adaptador que consumidores como o
Strategist utilizam para invocar operações sem se acoplar a uma skill
específica.

## Interface conceitual

```go
type KnowledgeProvider interface {
    Prepare(ctx Context, request PrepareRequest) (PrepareResult, error)
    Search(ctx Context, query Query)             ([]KnowledgeItem, error)
    Refresh(ctx Context, scope Scope)            (RefreshResult, error)
    Status(ctx Context)                          (ProviderStatus, error)
    Explain(ctx Context, itemID string)          (Explanation, error)
}
```

`Explain` é obrigatório em um sistema IA-first: deve informar por que um
item foi selecionado, descartado ou considerado incompatível.

## Envelope de resposta

Toda resposta carrega, no mínimo:

```yaml
provider: <string>
version: <semver>
capabilities_used: [<string>]
sources:
  - path: <string>
    digest: sha256:...
freshness: fresh | partially_stale | stale | unknown
trust: <float>
compatibility: <string>
cost:
  estimated: <int>
  actual: <int>
limitations: [<string>]
fallback_state: none | embedded | degraded
```

## Roteamento por capacidade

Consumidores escolhem por capability, não por nome de provider:

```yaml
requires:
  - capability: architecture.map
    version: "^1.0"
```

O resolver escolhe qual skill instalada satisfaz a capability.

## Degradação explícita

Quando uma skill externa estiver ausente, incompatível ou indisponível,
o consumidor deve:

1. registrar o estado degradado;
2. usar bundle embarcado compatível, quando existir;
3. informar limitações da resposta;
4. **nunca simular** que a consulta externa ocorreu.

## Anti-patterns

- Escolher provider por nome em vez de capability.
- Considerar provider registrado como provider saudável.
- Carregar todo o catálogo no prompt.
- Tratar índice como fonte canônica.
- Esconder estado degradado do usuário.
