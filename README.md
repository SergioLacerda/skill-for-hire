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

O binário canônico é `skillhire`. Superfície mínima disponível hoje:

```text
skillhire validate <skill-dir>...       # package-static validation
skillhire pack     <skill-dir>...       # produz <name>-<version>.tar.gz + .sha256 + .release.yaml
skillhire lock     <release.yaml>...    # gera/verifica skillhire.lock
skillhire verify   <archive.tar.gz>...  # valida um asset baixado
skillhire version
```

A roadmap completa da CLI está em [`docs/cli.md`](docs/cli.md).

## Consumindo um release

Cada release publica, para cada skill, três arquivos:

- `<name>-<version>.tar.gz` — pacote ORKA (raiz do tar = `<name>/`)
- `<name>-<version>.tar.gz.sha256` — digest SHA-256
- `<name>-<version>.release.yaml` — bundle manifest (schema, digest, size,
  generator, `source_commit`, `contents[]`)

Além disso: binários `skillhire-<os>-<arch>`, `SHA256SUMS`, SBOM
CycloneDX e attestations de build provenance.

### Download → verify → extract (sem depender do CLI)

```bash
VERSION=v0.0.1
BASE="https://github.com/SergioLacerda/skill-for-hire/releases/download/${VERSION}"
SKILL="atlas-0.1.0"

curl -sSLO "${BASE}/${SKILL}.tar.gz"
curl -sSLO "${BASE}/${SKILL}.tar.gz.sha256"

sha256sum -c "${SKILL}.tar.gz.sha256"
tar -xzf "${SKILL}.tar.gz"        # extrai para ./atlas/
```

### Download → verify → extract (com o CLI)

```bash
curl -sSLO "${BASE}/${SKILL}.tar.gz"
curl -sSLO "${BASE}/${SKILL}.tar.gz.sha256"
curl -sSLO "${BASE}/${SKILL}.release.yaml"

skillhire verify "${SKILL}.tar.gz"    # cross-check archive x .sha256 x manifest
tar -xzf "${SKILL}.tar.gz"
```

`skillhire verify` valida três coisas: o digest do tar.gz bate com o
`.sha256`, o `.sha256` referencia o basename correto (não outro asset),
e o `.release.yaml` (quando presente) concorda com digest e size.

### Attestations (opcional)

```bash
gh attestation verify "${SKILL}.tar.gz" \
  --repo SergioLacerda/skill-for-hire
```

### Lockfile

Consumidores fixam versões em `skillhire.lock` (uma cópia por projeto).
Este repositório mantém o próprio [`skillhire.lock`](skillhire.lock)
como referência do formato e como gate de CI (`make lock-verify`).

## Documentação

- [`docs/architecture/`](docs/architecture/) — filosofia e arquitetura do
  ecossistema, padrão de skills, orientações IA-first.

## Licença

Ver [`LICENSE`](LICENSE).
