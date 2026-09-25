# treasure-chest — lifecycle

## Estados de documento (herdados da plataforma)

```text
discovered → indexed → stale → refreshed
                    ↘ invalid
```

## Estados curatoriais (específicos da Treasure Chest)

```text
candidate → reviewed → accepted → deprecated
```

## Regras

1. **Candidatos não são conhecimento aceito.** Promoção exige política e,
   por padrão, revisão humana.
2. **Provenance obrigatória.** Todo item registra origem, digest,
   evidência e responsável.
3. **Trust, freshness e escopo** são propriedades curatoriais
   independentes.
4. **TTL** pode disparar reverificação; nunca altera automaticamente o
   status curatorial de um item aceito.
5. **Conflitos** entre itens são explicitados na resposta, não
   silenciados.

## Classificação mínima de runbooks

```yaml
scope: global | product | workspace
owner: atlas | treasure-chest | strategist | providence | user
portability: portable | parameterized | product_specific
distribution: embedded | optional_pack | workspace_only
```

Generalização não deve ocorrer apenas por semelhança textual. Um
runbook global precisa declarar parâmetros, dependências,
compatibilidade, safety e critérios de verificação.
