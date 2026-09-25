# Guia do consumidor — Skills for Hire

Este documento descreve **como um projeto consumidor** (Strategist,
outra CLI, um agente autônomo) importa, verifica, fixa e integra as
skills publicadas por este repositório.

Alvo desta versão: **`v0.0.1`**.

---

## 1. O que uma release publica

Cada tag `v*.*.*` dispara o workflow de release e publica em
`https://github.com/SergioLacerda/skill-for-hire/releases/tag/<version>`
o conjunto de assets abaixo.

### 1.1 Binários da CLI `skillhire`

Uma matriz de 6 binários estáticos (CGO desabilitado, `-trimpath`,
`-s -w`, `mod_timestamp` reprodutível):

```text
skillhire-linux-amd64
skillhire-linux-arm64
skillhire-darwin-amd64
skillhire-darwin-arm64
skillhire-windows-amd64.exe
skillhire-windows-arm64.exe
```

Acompanhados de `SHA256SUMS` (formato coreutils) com os digests de todos
os binários da matriz.

### 1.2 Skill packs

Para **cada** skill publicada, três arquivos ficam alinhados por nome:

```text
<name>-<version>.tar.gz          # o pacote (layout ORKA)
<name>-<version>.tar.gz.sha256   # digest coreutils "digest  filename"
<name>-<version>.release.yaml    # bundle manifest (sidecar)
```

Skills disponíveis a partir de `v0.0.2`:

| Skill | Versão | Runtime | Capabilities publicadas |
| --- | --- | --- | --- |
| `atlas` | `0.1.0` | *hello-world* (só contrato) | `architecture.map`, `architecture.search`, `architecture.impact`, `adr.discover`, `adr.relate` |
| `treasure-chest` | `0.2.0` | ✅ runtime Go + binário standalone | `knowledge.mine`, `knowledge.search`, `knowledge.curate`, `runbook.select`, `learning.reuse` |

`treasure-chest` já implementa a Knowledge API v1 end-to-end (Prepare /
Search / Refresh / Status / Explain). O binário `treasure-chest-<os>-<arch>`
é publicado como asset da release. `atlas` segue como contrato
sem implementação por enquanto — consumidores devem tratá-lo como a
*specialist* que descreve o contrato que uma implementação futura
satisfaz.

### 1.3 Binários das skills (novo em `v0.0.2`)

Skills que shipam runtime executável ganham binários standalone
próprios, mesma matriz do `skillhire`:

```text
treasure-chest-linux-amd64
treasure-chest-linux-arm64
treasure-chest-darwin-amd64
treasure-chest-darwin-arm64
treasure-chest-windows-amd64.exe
treasure-chest-windows-arm64.exe
```

Uso:

```bash
treasure-chest prepare --root /path/to/workspace
treasure-chest search  --root /path/to/workspace --intent "dependency upgrade"
treasure-chest status  --root /path/to/workspace --json
treasure-chest explain <jewel-id> --root /path/to/workspace
```

Todas as respostas seguem o envelope da Knowledge API v1
(§5 abaixo).

### 1.4 Supply chain

- **CycloneDX SBOM** — `skillhire-<version>-sbom.cdx.json` cobrindo o
  grafo de dependências Go do CLI.
- **Build provenance attestations** (armazenadas no attestation store
  do GitHub, não como asset):
  - CLI binaries — atestados via `SHA256SUMS`
  - Skill packs — atestados por path (`*.tar.gz` + `*.release.yaml`)

---

## 2. Layout do pacote ORKA

O tar.gz de cada skill tem como **raiz o próprio nome da skill** (não
o path fonte no monorepo). Ao extrair:

```text
<extract-dir>/
└── atlas/
    ├── SKILL.md
    ├── skill.yaml
    ├── README.md
    ├── CHANGELOG.md
    └── references/
        ├── capabilities.md
        └── lifecycle.md
```

### 2.1 `SKILL.md` — contrato instrucional portátil

Frontmatter YAML mínimo (compatível com Agent Skills abertos):

```yaml
---
name: atlas
description: Use esta skill para mapear e pré-indexar a arquitetura...
metadata:
  version: 0.1.0
  author: skill-for-hire
---
```

Corpo em Markdown com seções semânticas: objetivo, quando usar / não
usar, entradas, procedimento, critérios de conclusão, saídas, limites,
referências sob demanda.

### 2.2 `skill.yaml` — manifesto operacional

Adaptador do projeto (schema `skillsforhire.dev/v1`):

```yaml
apiVersion: skillsforhire.dev/v1
kind: Skill
metadata:
  name: atlas
  version: "0.1.0"
  owner: skill-for-hire
  lifecycle: experimental
spec:
  entrypoint: SKILL.md
  capabilities:
    network:    { mode: denied }
    filesystem: { mode: workspace-read }
    subprocess: { mode: denied }
    secrets:    { mode: denied }
  composition:
    provides:
      - architecture.map
      - architecture.search
      # ...
```

**Deny-by-default:** capacidades ausentes são consideradas negadas. O
consumidor impõe sandbox, allowlists e confirmações independentemente
do frontmatter.

### 2.3 `<name>-<version>.release.yaml` — bundle manifest

Sidecar reprodutível (byte-identical entre runs) carregando:

```yaml
schema_version: 1
name: atlas
version: "0.1.0"
archive: atlas-0.1.0.tar.gz
digest: sha256:f8671ffdec326ca0000998321e949775ce6afc7f4fe6a0e881c21ab410329f02
size: 4514
generator: skillhire@v0.0.1
source_commit: <sha>
contents:
  - atlas/CHANGELOG.md
  - atlas/README.md
  - atlas/SKILL.md
  - atlas/references/capabilities.md
  - atlas/references/lifecycle.md
  - atlas/skill.yaml
```

---

## 3. Fluxo de consumo

### 3.1 Fluxo mínimo (sem CLI, só shell)

```bash
VERSION=v0.0.1
BASE="https://github.com/SergioLacerda/skill-for-hire/releases/download/${VERSION}"
SKILL="atlas-0.1.0"

curl -sSLO "${BASE}/${SKILL}.tar.gz"
curl -sSLO "${BASE}/${SKILL}.tar.gz.sha256"

sha256sum -c "${SKILL}.tar.gz.sha256"
tar -xzf "${SKILL}.tar.gz"          # extrai para ./atlas/
```

### 3.2 Fluxo com `skillhire verify` (cross-check completo)

```bash
curl -sSLO "${BASE}/${SKILL}.tar.gz"
curl -sSLO "${BASE}/${SKILL}.tar.gz.sha256"
curl -sSLO "${BASE}/${SKILL}.release.yaml"

skillhire verify "${SKILL}.tar.gz"    # checa archive x .sha256 x manifest
tar -xzf "${SKILL}.tar.gz"
```

`skillhire verify` recusa três situações que `sha256sum -c` ignora:
- `.sha256` que referencia um basename diferente do archive
- Manifesto com digest divergente do archive
- Manifesto com size divergente

### 3.3 Fluxo com attestation (opcional)

```bash
gh attestation verify "${SKILL}.tar.gz" \
  --repo SergioLacerda/skill-for-hire
```

Valida que o pacote foi produzido pelo workflow oficial no commit
esperado.

---

## 4. Lockfile (`skillhire.lock`)

Projetos consumidores fixam suas skills em um `skillhire.lock`
committado ao repositório. Formato (schema
[`doctrine/lockfile.schema.json`](../doctrine/lockfile.schema.json)):

```yaml
schema-version: 1
generator: "skillhire@v0.0.1"
installed:
  atlas:
    version: "0.1.0"
    digest: sha256:f8671ffdec326ca0000998321e949775ce6afc7f4fe6a0e881c21ab410329f02
    size: 4514
    source: monorepo:skills/atlas
    resolved-from: atlas-0.1.0.release.yaml
  treasure-chest:
    version: "0.1.0"
    digest: sha256:03eebd0e91e81c9ee757287913c75971f7b0504975ca08e544d30d962e15aad3
    size: 4306
    source: monorepo:skills/treasure-chest
    resolved-from: treasure-chest-0.1.0.release.yaml
```

### 4.1 Geração e verificação

```bash
# Gera o lockfile a partir dos manifestos baixados:
skillhire lock atlas-0.1.0.release.yaml treasure-chest-0.1.0.release.yaml \
  --out skillhire.lock

# Verifica que o lockfile continua batendo com os manifestos:
skillhire lock atlas-0.1.0.release.yaml treasure-chest-0.1.0.release.yaml \
  --out skillhire.lock --verify
```

`--verify` compara **apenas o mapa `installed`**; o campo `generator`
naturalmente varia entre local / CI / release tagueada.

### 4.2 Integração em CI

Adicione um gate:

```bash
skillhire lock <manifests>... --out skillhire.lock --verify
```

Falha loud quando o lockfile committado não reflete os manifestos das
skills atualmente listadas.

---

## 5. Contrato de integração (Knowledge API v1)

Cada skill de conhecimento futura satisfaz a interface Go em
[`platform/knowledge-api/`](../platform/knowledge-api/):

```go
type KnowledgeProvider interface {
    Prepare(ctx, PrepareRequest) (PrepareResult, error)
    Search (ctx, Query)          (SearchResult, error)
    Refresh(ctx, Scope)          (RefreshResult, error)
    Status (ctx)                 (Status, error)
    Explain(ctx, itemID string)  (Explanation, error)
}
```

Toda resposta carrega um `Envelope`:

```go
type Envelope struct {
    Provider         string
    Version          string
    SchemaVersion    int
    CapabilitiesUsed []string
    Sources          []Source          // path + digest + commit
    Freshness        Freshness         // fresh | partially_stale | stale | unknown
    Trust            float64
    Compatibility    string
    Cost             Cost
    Limitations      []string
    FallbackState    FallbackState     // none | embedded | degraded
}
```

Erros normalizados (para roteamento no consumidor):

```text
ErrNotPrepared
ErrIncompatibleVersion
ErrPermissionDenied
ErrSourceMissing
ErrStale
ErrFallbackTriggered
ErrUnavailable
```

### 5.1 Roteamento por capability, não por nome

O consumidor não deve acoplar-se a nomes (`atlas`, `treasure-chest`).
Rota por capability declarada em `spec.composition.provides`:

```yaml
requires:
  - capability: architecture.map
    version: "^1.0"
```

O resolver do consumidor escolhe qual skill instalada satisfaz.

### 5.2 Degradação explícita

Quando uma skill está ausente ou incompatível, o consumidor deve:

1. registrar o estado degradado;
2. usar bundle embarcado compatível, se existir;
3. informar limitações da resposta;
4. **nunca simular** que a consulta ocorreu.

---

## 6. Adapter do Strategist (referência)

Contrato de tradução em [`adapters/strategist/`](../adapters/strategist/):

| Piloto do Strategist | Skill consumida | Capabilities |
| --- | --- | --- |
| Cartógrafo | `atlas` | `architecture.*`, `adr.*` |
| Jewelcrafter | `treasure-chest` | `knowledge.*`, `runbook.select`, `learning.reuse` |

Namespaces neutros expostos pelo Strategist:

```bash
strategist knowledge prepare
strategist knowledge status
strategist knowledge refresh
```

O adapter adiciona policy, telemetria e fallback embarcado; as skills
não sabem que estão sendo consumidas pelo Strategist. Adapters **não**
vazam detalhes internos para o Strategist.

---

## 7. Roster + companies (composição)

O catálogo `roster/skills.yaml` e o bundle `roster/companies.yaml`
descrevem composições recomendadas:

```yaml
companies:
  strategist-knowledge-pack:
    version: "0.1.0"
    skills:
      atlas: "^0.1"
      treasure-chest: "^0.1"
```

Consumidores podem instalar a company inteira ou skills individualmente.

---

## 8. Exemplos práticos

### 8.1 Bash: baixar, verificar e extrair as duas skills

```bash
#!/usr/bin/env bash
set -euo pipefail

VERSION=v0.0.2
BASE="https://github.com/SergioLacerda/skill-for-hire/releases/download/${VERSION}"
DEST="./skills"

mkdir -p "$DEST" && cd "$DEST"

for skill in atlas-0.1.0 treasure-chest-0.2.0; do
  for suffix in tar.gz tar.gz.sha256 release.yaml; do
    curl -sSLO "${BASE}/${skill}.${suffix}"
  done
  sha256sum -c "${skill}.tar.gz.sha256"
  tar -xzf "${skill}.tar.gz"
done

ls -la
# atlas/
# treasure-chest/
# atlas-0.1.0.tar.gz
# atlas-0.1.0.tar.gz.sha256
# atlas-0.1.0.release.yaml
# treasure-chest-0.1.0.tar.gz
# ...
```

### 8.2 Go: carregar `skill.yaml` e rotear por capability

```go
package main

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

type SkillManifest struct {
    Metadata struct {
        Name    string `yaml:"name"`
        Version string `yaml:"version"`
    } `yaml:"metadata"`
    Spec struct {
        Composition struct {
            Provides []string `yaml:"provides"`
        } `yaml:"composition"`
    } `yaml:"spec"`
}

func main() {
    b, _ := os.ReadFile("skills/atlas/skill.yaml")
    var m SkillManifest
    _ = yaml.Unmarshal(b, &m)

    fmt.Printf("%s@%s provides: %v\n",
        m.Metadata.Name, m.Metadata.Version, m.Spec.Composition.Provides)
    // atlas@0.1.0 provides: [architecture.map architecture.search ...]
}
```

### 8.3 CI gate: só permite build se lockfile estiver batendo

```yaml
- name: Verify skill pins
  run: |
    skillhire lock \
      vendor/skills/*.release.yaml \
      --out skillhire.lock --verify
```

---

## 9. Atualização (bumping)

Para adotar uma versão nova de uma skill:

1. Baixe o `<name>-<version>.release.yaml` da nova release.
2. Rode `skillhire lock ... --out skillhire.lock` (sem `--verify`).
3. Diff o lockfile — a linha `digest:` deve mudar.
4. Commite `skillhire.lock` junto com qualquer código que dependa da
   nova capability.

O consumidor **nunca** deve resolver silenciosamente uma tag mutável.
Atualizações passam pelo lockfile revisado.

---

## 10. Referências cruzadas

| Documento | Escopo |
| --- | --- |
| [`README.md`](../README.md) | Visão geral do projeto |
| [`docs/architecture/`](architecture/) | Filosofia e padrões |
| [`docs/cli.md`](cli.md) | Superfície da CLI `skillhire` |
| [`doctrine/skill.schema.json`](../doctrine/skill.schema.json) | Schema do `skill.yaml` |
| [`doctrine/lockfile.schema.json`](../doctrine/lockfile.schema.json) | Schema do `skillhire.lock` |
| [`platform/knowledge-api/`](../platform/knowledge-api/) | Contrato Go da Knowledge API |
| [`adapters/strategist/`](../adapters/strategist/) | Mapeamento para pilotos do Strategist |
| [`skillhire.lock`](../skillhire.lock) | Lockfile próprio do repositório |
