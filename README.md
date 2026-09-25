# Skills for Hire

> Recruit, compose, and deploy reusable AI skills.

**Skills for Hire** é um catálogo e um toolkit de distribuição para skills de IA
autônomas, versionadas e combináveis. Cada skill é uma especialista
autocontida — com `SKILL.md`, manifesto, versão e ciclo de vida próprios —
que pode ser instalada isoladamente ou composta em *companies* (bundles
declarativos) para missões maiores.

## Princípios

- **Autocontenção** — cada skill vive em um diretório próprio, com tudo o
  que precisa para ser compreendida, validada, testada e empacotada.
- **Versionamento independente** — monorepo para autoria, releases
  individuais por skill (SemVer).
- **Composição declarativa** — *companies* referenciam skills por
  capability e faixa de versão; nunca copiam conteúdo.
- **Catálogo aberto** — o *roster* aceita skills do core, do monorepo,
  de repositórios próprios e de terceiros.
- **Instalação explícita** — a orquestradora não baixa código
  silenciosamente; instalação, atualização e remoção são operações
  auditáveis via `skillhire`.
- **Deny-by-default** — capacidades não declaradas são consideradas
  negadas; o runtime impõe sandbox, allowlists e confirmação.

## Vocabulário

| Técnico | Temático | Função |
| --- | --- | --- |
| Skill | Specialist | Unidade autônoma de capacidade |
| Catálogo | Roster | Skills conhecidas e disponíveis |
| Bundle | Company | Composição versionada de skills |
| Instalação | Recruit | Torna uma skill disponível localmente |
| Manifesto | Dossier | Metadados, versão e contratos da skill |

## Estrutura do repositório

```text
skill-for-hire/
├── skills/                # skills autocontidas, uma por diretório kebab-case
│   ├── atlas/             # cartografia estrutural e documental do projeto
│   └── treasure-chest/    # curadoria, lifecycle e recuperação de conhecimento
├── roster/                # catálogo federado (skills.yaml, companies.yaml)
├── doctrine/              # schemas, convenções e políticas compartilhadas
├── platform/              # contratos compartilhados (Knowledge API, source registry, packaging)
├── adapters/              # adaptadores por cliente/orquestrador (Strategist, standalone)
├── distributions/         # artefatos e bundles publicáveis
├── cmd/skillhire/         # CLI (Go) — em construção
└── docs/                  # documentação de arquitetura e padrões
```

## Skills atuais

| Skill | Status | Descrição |
| --- | --- | --- |
| [`atlas`](skills/atlas/) | experimental | Mapeamento e pré-indexação arquitetural e documental de projetos. |
| [`treasure-chest`](skills/treasure-chest/) | experimental | Mineração, curadoria, lifecycle e recuperação de conhecimento reutilizável (a ser exportada de outro projeto). |

## CLI

O binário canônico será `skillhire`. Ainda não implementado — a superfície
mínima planejada segue em `docs/cli.md`.

## Documentação

- [`docs/architecture/`](docs/architecture/) — filosofia e arquitetura do
  ecossistema, padrão de skills, orientações IA-first.

## Licença

Ver [`LICENSE`](LICENSE).
