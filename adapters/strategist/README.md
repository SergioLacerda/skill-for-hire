# adapters/strategist

Adapter que traduz a Knowledge API (`platform/knowledge-api/`) para os
pilotos do Strategist:

| Piloto (Strategist) | Skill (Skills for Hire) | Capabilities consumidas |
| --- | --- | --- |
| Cartógrafo | `atlas` | `architecture.map`, `architecture.search`, `architecture.impact`, `adr.discover`, `adr.relate` |
| Jewelcrafter | `treasure-chest` | `knowledge.mine`, `knowledge.search`, `knowledge.curate`, `runbook.select`, `learning.reuse` |

## Regras

- Skills não sabem que estão sendo consumidas pelo Strategist. Toda a
  tradução vive aqui.
- O adapter adiciona policy, telemetria e fallback embarcado.
- Pilotos são consumidores privilegiados, **não** o runtime obrigatório
  das skills.
- Adapters **não** vazam detalhes internos das skills para o Strategist.

## Namespaces neutros expostos pelo Strategist

```bash
strategist knowledge prepare
strategist knowledge status
strategist knowledge refresh
```

Internamente, o roteamento escolhe Atlas, Treasure Chest ou ambos.

## Status

Placeholder — implementação será feita quando a Knowledge API estiver
congelada e as duas skills tiverem release `0.1.0` estável.
