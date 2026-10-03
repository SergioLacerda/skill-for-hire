# Runbook — nova skill com runtime

Cria uma skill do zero seguindo o padrão "**estilo B**" (contrato ORKA +
runtime Go + CLI standalone + adapter Knowledge API). É o mesmo shape
que `treasure-chest@0.2.0` tem hoje. Use quando a skill precisar rodar
código próprio (indexação, ranking, curadoria, recuperação).

Se você só quer contrato-only (SKILL.md + skill.yaml, sem código),
use o template do `atlas` como modelo direto — este runbook é para
quem vai além disso.

## Quando usar

- Nova skill que expõe capabilities executáveis (Prepare/Search/…).
- Portar uma skill de outro repo e escrever o runtime do zero (para
  importar código existente, use `migrate-external-skill.md`).

## Pré-requisitos

- `git status` limpo em `develop` (ou branch derivada dela).
- Toolchain: Go 1.23+, `make`, `golangci-lint` v2.x, opcionalmente
  `goreleaser` v2.12+ para testar cross-compile local.
- Nome da skill escolhido — kebab-case, único no `roster/skills.yaml`.

## Passos

### 1. Criar a branch

```bash
git checkout develop
git pull --ff-only
git checkout -b feat/<name>-skill
```

### 2. Layout mínimo

```bash
NAME="my-skill"          # <-- ajuste
VERSION="0.1.0"

mkdir -p "skills/${NAME}"/{references,runtime/domain}
mkdir -p "cmd/${NAME}"
```

Estrutura final que este runbook produz:

```text
skills/${NAME}/
├── SKILL.md
├── skill.yaml
├── README.md
├── CHANGELOG.md
├── references/
│   ├── capabilities.md
│   └── lifecycle.md
└── runtime/
    ├── doc.go
    ├── adapter.go
    └── adapter_test.go

cmd/${NAME}/
├── main.go
├── root.go
├── commands.go
└── commands_test.go
```

### 3. `SKILL.md`

Contrato instrucional portátil (frontmatter YAML + Markdown):

```markdown
---
name: my-skill
description: Use esta skill para <descrição operacional curta>. Use quando <trigger>. Não use quando <exclusão>.
metadata:
  version: 0.1.0
  author: skill-for-hire
---

# My Skill

## Objetivo

<uma frase>.

## Quando usar

- <trigger 1>
- <trigger 2>

## Quando não usar

- <exclusão 1>

## Entradas necessárias

- Raiz do projeto (workspace root).

## Procedimento

1. **Prepare** — descobre fontes, calcula digests, indexa.
2. **Search** — devolve candidatos ranqueados com razão.
3. **Refresh** — recalcula seletivamente.
4. **Status** — reporta saúde, capabilities e freshness.
5. **Explain** — explica por que um item foi selecionado ou descartado.

## Critérios de conclusão

- Toda resposta cita fontes com path + digest.
- Freshness sempre reportada.

## Saídas

- Envelope Knowledge API com `provider`, `version`, `capabilities_used`,
  `freshness`, `trust`, `sources[]`, `limitations`, `fallback_state`.

## Limites e escalonamento

- Filesystem: leitura na raiz autorizada.
- Rede: negada por padrão.

## Referências sob demanda

- `references/capabilities.md` — contrato das operações.
- `references/lifecycle.md` — estados de índice + política de atualização.
```

### 4. `skill.yaml`

Manifesto operacional (schema `skillsforhire.dev/v1`):

```yaml
apiVersion: skillsforhire.dev/v1
kind: Skill

metadata:
  name: my-skill
  version: "0.1.0"
  displayName: My Skill
  description: <descrição curta para catálogo>.
  owner: skill-for-hire
  lifecycle: experimental
  tags:
    - <tag1>
    - <tag2>

spec:
  entrypoint: SKILL.md

  compatibility:
    clients:
      - name: generic-agent-skills
        version: ">=1"

  capabilities:
    network:
      mode: denied
    filesystem:
      mode: workspace-read
    subprocess:
      mode: denied
    secrets:
      mode: denied

  execution:
    timeoutSeconds: 120
    maxOutputBytes: 10485760
    installDependenciesAtRuntime: false

  dependencies:
    required: []
    optional: []

  composition:
    provides:
      - my-skill.example-capability
    conflictsWith: []
    priority: 100

  quality:
    triggerEvals: tests/trigger/eval-queries.json
    behaviorTests: tests/behavior/cases.yaml

  distribution:
    licenseFile: ../../LICENSE
    changelog: CHANGELOG.md
    sourceRepository: https://github.com/SergioLacerda/skill-for-hire
    integrity:
      algorithm: sha256
```

Regra crítica: os capabilities em `spec.composition.provides` têm que
estar **em sincronia** com a lista `capabilities` no `runtime/adapter.go`
(passo 8). Um teste de compile-time flag isso.

### 5. `README.md`, `CHANGELOG.md`

`README.md`:

```markdown
# my-skill

<uma frase>.

## Status

`experimental` — versão `0.1.0`.

## Capabilities

| Capability | Propósito |
| --- | --- |
| `my-skill.example-capability` | <descrição>. |

## Uso standalone

\`\`\`bash
my-skill prepare --root /path/to/workspace
my-skill search  --root /path/to/workspace --intent "..."
my-skill status  --root /path/to/workspace [--json]
my-skill explain <item-id> --root /path/to/workspace
\`\`\`

## Estrutura

\`\`\`text
my-skill/
├── SKILL.md
├── skill.yaml
├── README.md
├── CHANGELOG.md
├── references/
└── runtime/
\`\`\`
```

`CHANGELOG.md`:

```markdown
# Changelog — my-skill

## [Unreleased]

## [0.1.0] — YYYY-MM-DD

- Bootstrap com runtime + adapter Knowledge API + CLI standalone.
```

### 6. Referências (`references/*.md`)

Um `.md` por tópico, <200 linhas cada. Modele com base em
[`skills/atlas/references/`](../../skills/atlas/references/) ou
[`skills/treasure-chest/references/`](../../skills/treasure-chest/references/).

Mínimo esperado:

- `references/capabilities.md` — contrato de cada operação (Prepare/…),
  com entrada + saída YAML.
- `references/lifecycle.md` — estados de índice, invariantes.

### 7. `runtime/doc.go`

```go
// Package runtime holds the my-skill Go implementation of the
// platform/knowledge-api KnowledgeProvider contract.
package runtime
```

### 8. `runtime/adapter.go`

Aqui é onde a skill vira executável. Template mínimo que compila e
passa lint:

```go
package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
)

const providerName = "my-skill"

// capabilities MUST stay in sync with skills/my-skill/skill.yaml's
// spec.composition.provides.
var capabilities = []string{
	"my-skill.example-capability",
}

// Adapter implements platform/knowledge-api.KnowledgeProvider on top of
// the my-skill domain.
type Adapter struct {
	version string

	mu         sync.RWMutex
	root       string
	prepared   bool
	preparedAt time.Time
	// Add domain state fields here (index, byID, etc.)
}

// NewAdapter returns a fresh adapter. Pass the skill's SemVer.
func NewAdapter(version string) *Adapter {
	return &Adapter{version: version}
}

func (a *Adapter) Prepare(_ context.Context, req ka.PrepareRequest) (ka.PrepareResult, error) {
	if req.Root == "" {
		return ka.PrepareResult{}, fmt.Errorf("my-skill: PrepareRequest.Root is required")
	}

	// TODO: load domain state from req.Root here.

	a.mu.Lock()
	a.root = req.Root
	a.prepared = true
	a.preparedAt = time.Now().UTC()
	a.mu.Unlock()

	return ka.PrepareResult{Envelope: a.envelope(nil)}, nil
}

func (a *Adapter) Search(_ context.Context, _ ka.Query) (ka.SearchResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if !a.prepared {
		return ka.SearchResult{}, ka.ErrNotPrepared
	}
	// TODO: rank domain items against the query.
	return ka.SearchResult{Envelope: a.envelope(nil)}, nil
}

func (a *Adapter) Refresh(ctx context.Context, scope ka.Scope) (ka.RefreshResult, error) {
	root := scope.Root
	if root == "" {
		a.mu.RLock()
		root = a.root
		a.mu.RUnlock()
	}
	if root == "" {
		return ka.RefreshResult{}, ka.ErrNotPrepared
	}
	if _, err := a.Prepare(ctx, ka.PrepareRequest{Root: root}); err != nil {
		return ka.RefreshResult{}, err
	}
	return ka.RefreshResult{Envelope: a.envelope(nil)}, nil
}

func (a *Adapter) Status(_ context.Context) (ka.Status, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return ka.Status{
		Provider:      providerName,
		Version:       a.version,
		SchemaVersion: ka.SchemaVersion,
		Healthy:       a.prepared,
		Capabilities:  capabilities,
		LastPrepareAt: a.preparedAt.Format(time.RFC3339),
	}, nil
}

func (a *Adapter) Explain(_ context.Context, itemID string) (ka.Explanation, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if !a.prepared {
		return ka.Explanation{}, ka.ErrNotPrepared
	}
	// TODO: look up itemID in the domain state.
	return ka.Explanation{
		ItemID:   itemID,
		Decision: ka.DecisionDiscarded,
		Reason:   "id not present in current index",
	}, nil
}

func (a *Adapter) envelope(limitations []string) ka.Envelope {
	freshness := ka.FreshnessUnknown
	if a.prepared {
		freshness = ka.FreshnessFresh
	}
	return ka.Envelope{
		Provider:         providerName,
		Version:          a.version,
		SchemaVersion:    ka.SchemaVersion,
		CapabilitiesUsed: capabilities,
		Freshness:        freshness,
		Trust:            1.0,
		FallbackState:    ka.FallbackNone,
		Limitations:      limitations,
	}
}

// Compile-time proof.
var _ ka.KnowledgeProvider = (*Adapter)(nil)
```

### 9. `runtime/adapter_test.go`

```go
package runtime_test

import (
	"context"
	"errors"
	"testing"

	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
	"github.com/SergioLacerda/skill-for-hire/skills/my-skill/runtime"
)

func TestAdapterSatisfiesKnowledgeProvider(_ *testing.T) {
	var _ ka.KnowledgeProvider = (*runtime.Adapter)(nil)
}

func TestPrepareRequiresRoot(t *testing.T) {
	a := runtime.NewAdapter("0.1.0")
	if _, err := a.Prepare(context.Background(), ka.PrepareRequest{}); err == nil {
		t.Fatal("expected error when Root is empty")
	}
}

func TestSearchWithoutPrepareReturnsSentinel(t *testing.T) {
	a := runtime.NewAdapter("0.1.0")
	_, err := a.Search(context.Background(), ka.Query{})
	if !errors.Is(err, ka.ErrNotPrepared) {
		t.Fatalf("expected ErrNotPrepared, got %v", err)
	}
}
```

### 10. CLI standalone — `cmd/<name>/`

Copie o esqueleto de `cmd/treasure-chest/` como referência:

- [`main.go`](../../cmd/treasure-chest/main.go)
- [`root.go`](../../cmd/treasure-chest/root.go) — só troque `NewAdapter`
  e as constantes de piloto pelas suas
- [`commands.go`](../../cmd/treasure-chest/commands.go) — `prepare`,
  `search`, `status`, `explain`, `version` com `--root` e `--json`
- [`commands_test.go`](../../cmd/treasure-chest/commands_test.go)

O `make/go.mk` compila automaticamente qualquer `cmd/<name>/` novo — não
precisa mexer no Makefile.

### 11. Registrar no roster

`roster/skills.yaml` — adicionar a entrada:

```yaml
skills:
  # ... skills existentes ...
  my-skill:
    displayName: My Skill
    description: <descrição para catálogo>.
    version: "0.1.0"
    lifecycle: experimental
    source:
      type: monorepo
      path: skills/my-skill
    provides:
      - my-skill.example-capability
    integrity:
      algorithm: sha256
```

### 12. Cross-compile no release

`.goreleaser.yaml` — adicione um novo bloco `builds:` copiando o de
`treasure-chest`:

```yaml
  - id: my-skill
    main: ./cmd/my-skill
    binary: my-skill
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    env:
      - CGO_ENABLED=0
    flags:
      - -trimpath
    ldflags:
      - -s -w -X main.Version={{.Version}}
    mod_timestamp: "{{ .CommitTimestamp }}"
```

E um novo `archives:`:

```yaml
  - id: my-skill-binary
    ids: [my-skill]
    formats: [binary]
    name_template: "my-skill-{{ .Os }}-{{ .Arch }}"
```

### 13. Validar

```bash
make fmt-check
make ci-lint       # build + vet + fmt-check + mod-check
make lint          # golangci-lint (0 issues obrigatório)
make ci-test       # build + validate-skills + pack-skills + lock-verify
```

Se `lock-verify` falhar (esperado — o novo pack ainda não está no
lockfile), regenere:

```bash
make lock-skills
make lock-verify   # agora passa
```

### 14. Documentação de consumidor

`docs/consumer-guide.md` § 1.2 — adicionar linha na tabela de skills:

```markdown
| `my-skill` | `0.1.0` | ✅ runtime Go + binário standalone | `my-skill.example-capability` |
```

### 15. Commit + PR

```bash
git add -A
git commit -m "feat(my-skill): bootstrap runtime + CLI at 0.1.0"
git push -u origin feat/my-skill-skill
```

Abrir PR contra `develop`.

## Verificação

Marque cada item antes de mergear:

- [ ] `make ci-lint` verde
- [ ] `make ci-test` verde
- [ ] `make lint` reporta 0 issues
- [ ] `skills/<name>/skill.yaml`'s `spec.composition.provides` bate com
      `runtime/adapter.go`'s `capabilities`
- [ ] `roster/skills.yaml` inclui a nova skill
- [ ] `docs/consumer-guide.md` menciona a nova skill
- [ ] `skillhire.lock` regenerado e commitado
- [ ] `.goreleaser.yaml` produz binário no snapshot (`goreleaser release
      --snapshot --clean --skip=publish` mostra 6 binários novos)

## Rollback

Antes de merge:

```bash
git checkout develop
git branch -D feat/my-skill-skill
git push origin --delete feat/my-skill-skill
```

Depois de merge, antes de tag: reverter o commit no develop com
`git revert <sha>` e abrir PR do revert.

Depois de tag `v*.*.*` publicada: a release já saiu; para "unshipar" a
skill, cortar tag nova (`v0.0.<N+1>`) removendo os assets em um commit
antes.
