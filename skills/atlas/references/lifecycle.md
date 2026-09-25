# atlas — lifecycle e freshness

O Atlas é **derivado e descartável**. Sua identidade principal é o par
`(repository, commit)`.

## Estados

| Estado | Significado |
| --- | --- |
| `FRESH` | Atlas representa o `HEAD` atual. |
| `PARTIALLY_STALE` | Repositório mudou; regiões afetadas identificáveis. |
| `STALE` | Mudanças impedem atualização incremental segura. |
| `INCOMPATIBLE` | Schema ou versão do gerador não compatível com o runtime. |
| `MISSING` | Atlas ainda não foi construído. |

## Pipeline própria

```text
status → build | update | rebuild → validate → Atlas@HEAD
```

Independente da pipeline de missão do consumidor.

## Política de atualização

- `on-demand` — reindexa quando o consumidor solicitar e HEAD divergir.
- `on-commit` (CI) — reindexa a cada commit da default branch.
- `manual` — via `skillhire ...` ou CLI equivalente.

## Invariantes

1. Atlas é derivado e descartável.
2. Todo Atlas conhece o commit que representa.
3. Mudança de HEAD não implica rebuild total automaticamente.
4. Atualização incremental precisa ser validável.
5. Incerteza deve provocar expansão de leitura ou rebuild, não falsa
   confiança.
6. Fonte original deve ser reaberta quando necessária para decisão.

## Fingerprints obrigatórios

```yaml
fingerprint:
  repository: sha256:...
  manifest: sha256:...
schema_version: "1.0"
generator_version: <string>
repository:
  commit: <sha>
  branch: <string>
```
