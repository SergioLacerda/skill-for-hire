# Runbook — migrar skill de repositório externo

Importa código de domínio Go existente de outro repositório para dentro
de `skills/<name>/runtime/`, adaptando ao contrato Skills for Hire e
severando acoplamentos internos do projeto-fonte.

Baseado na experiência real de importar `treasure-chest` do
`strategist-skill` (Waves 4a → 4d). O mesmo procedimento serve para
qualquer skill que já tenha runtime maduro em outro lugar.

## Quando usar

- Você tem código em outro repo Go que já implementa lógica de
  domínio equivalente à skill que quer criar aqui.
- O código já está minimamente isolado no repo-fonte (idealmente uma
  pasta própria; um `architecture_test.go` de boundary é bônus).

Se o código estiver em outra linguagem, ou muito acoplado ao runtime
do projeto-fonte, considere reescrever com base em
`new-skill-with-runtime.md` em vez de migrar.

## Pré-requisitos

- `git status` limpo em `develop`.
- Repo-fonte clonado e acessível localmente (ou attachavel via
  `add_repo` se via Claude Code).
- Toolchain: Go 1.23+, `make`, `golangci-lint` v2.x.

## Passos

### 1. Branch de dry-run

```bash
git checkout develop
git pull --ff-only
git checkout -b dry-run/<name>-probe
```

### 2. Dry-run: identificar acoplamento

Objetivo: descobrir **exatamente** quais símbolos do projeto-fonte a
pasta que vamos migrar puxa. Sem isso o custo vira chute.

```bash
SRC=/path/to/source-repo/<source-folder>       # ex. treasure-chest/
PROBE=skills/<name>/runtime                    # ex. skills/treasure-chest/runtime
mkdir -p "$PROBE" "$PROBE/domain"

# Copie só o domínio + subpackage domain, deixe cli/ e testes de
# boundary fora
cp "$SRC"/*.go            "$PROBE/"
cp "$SRC"/domain/*.go     "$PROBE/domain/"
rm -f "$PROBE"/architecture_test.go   # se depender de internal/conformance
```

Reescreva os imports para o novo module path:

```bash
OLD_MOD="github.com/OWNER/source-repo/<source-folder>"
NEW_MOD="github.com/SergioLacerda/skill-for-hire/skills/<name>/runtime"

find "$PROBE" -name '*.go' -exec sed -i \
  -e "s|${OLD_MOD}/domain|${NEW_MOD}/domain|g" \
  -e "s|${OLD_MOD}\\b|${NEW_MOD}|g" \
  -e 's|^package treasure_test\b|package runtime_test|g' \
  -e 's|^package treasure\b|package runtime|g' \
  {} +
```

Tente compilar:

```bash
go build ./skills/<name>/... 2>&1 | tee /tmp/probe-build.log
```

Cada linha `no required module provides package github.com/OWNER/...`
é um ponto de acoplamento. Se você **não** vê nenhuma dessas mensagens,
sortudo — a migração é essencialmente `go mod tidy`.

### 3. Mapear os símbolos usados

Para cada import não resolvido, liste os símbolos referenciados:

```bash
grep -h 'domain\.' skills/<name>/runtime/*.go | sort -u
```

Duas categorias comuns:

**(a) Constantes / enums pequenos** — copie o subset relevante para
`skills/<name>/runtime/domain/vocab.go`. Foi o que fizemos com
`treasure-chest` (7 constantes de string ⇒ 22 linhas).

**(b) Tipos ou funções grandes** — decida caso a caso: copia inline,
extrai para um pacote próprio dentro da skill, ou reescreve com
interface mínima.

Regra: **não** trazer o pacote fonte inteiro como sub-módulo. Isso
gera acoplamento transitivo com todo o repo-fonte.

Exemplo (treasure-chest): `runtime/domain/vocab.go`

```go
package domain

// Constants below are the ENTIRE surface of internal/domain the
// migrated code ever reached for. Copying just the vocabulary here
// severs the dependency without dragging in ~10k LOC.
const (
	ConfidenceLow    = "low"
	ConfidenceMedium = "medium"
	ConfidenceHigh   = "high"
)

const (
	EvidenceClassExplicit              = "explicit"
	EvidenceClassCorroboratedInference = "corroborated_inference"
	EvidenceClassWeakInference         = "weak_inference"
	EvidenceClassUnknown               = "unknown"
)
```

Reescreva os imports nos arquivos migrados para usar o vocab local:

```bash
sed -i \
  -e 's|"github.com/OWNER/source-repo/internal/domain"|chestdomain "github.com/SergioLacerda/skill-for-hire/skills/<name>/runtime/domain"|' \
  -e 's|\bdomain\.ConfidenceLow\b|chestdomain.ConfidenceLow|g' \
  # ... uma linha por símbolo
  skills/<name>/runtime/<file>.go
```

### 4. Validar dry-run

```bash
go mod tidy                              # traz dependências públicas (testify etc)
go build ./skills/<name>/... 2>&1 | tail
go test  ./skills/<name>/... 2>&1 | tail
```

Se testes passam, a migração pura está concluída. Commite **em branch
`dry-run/`** (não `develop`):

```bash
git add -A
git commit -m "dry-run/<name>: import probe from source-repo"
git push -u origin dry-run/<name>-probe
```

Esse commit fica como fallback caso a integração seguinte quebre e
você queira voltar ao domínio nu.

### 5. Adapter Knowledge API

Renomeie a branch (o commit dry-run vira base):

```bash
git branch -m dry-run/<name>-probe feat/<name>-runtime
```

Escreva `skills/<name>/runtime/adapter.go` sobre os loaders migrados,
seguindo o padrão de `skills/treasure-chest/runtime/adapter.go` ou o
esqueleto em `new-skill-with-runtime.md` § 8.

Coloque a lógica de ranking (se houver) em `runtime/ranking.go`
separado — mantém `adapter.go` legível.

Teste o adapter:

```go
// runtime/adapter_test.go — no mínimo:
func TestAdapterSatisfiesKnowledgeProvider(_ *testing.T) {
    var _ ka.KnowledgeProvider = (*Adapter)(nil)
}
func TestPrepareRequiresRoot(t *testing.T) { ... }
func TestSearchWithoutPrepareReturnsSentinel(t *testing.T) { ... }
```

### 6. CLI standalone

Crie `cmd/<name>/` copiando `cmd/treasure-chest/` como base (main.go,
root.go, commands.go, commands_test.go). Ajuste:

- `runtime.NewAdapter(Version)` na `root.go`
- Nome do comando raiz em `Use:` da `commands.go`
- Testes em `commands_test.go` — cobrir status human+JSON, prepare
  vazio, search em workspace vazio, explain de ID inexistente

### 7. Adicionar depguard rule

`.golangci.yaml` — adicione um bloco `deny` novo garantindo que o
projeto não volte a importar do repo-fonte:

```yaml
settings:
  depguard:
    rules:
      no-source-repo-imports:
        list-mode: lax
        deny:
          - pkg: "github.com/OWNER/source-repo"
            desc: "skill-for-hire must not depend on source-repo — the runtime was migrated in-tree"
```

O rule existente para `strategist-skill` serve como modelo.

### 8. Bump da skill + metadados

- `skills/<name>/skill.yaml` — subir `metadata.version` (ex. `0.1.0`
  → `0.2.0` se estava com runtime hello-world antes; `0.1.0`
  direto se a skill é nova).
- `skills/<name>/README.md` — trocar a nota "a ser exportada" por
  algo como "runtime importado de OWNER/source-repo, ADR-NNN".
- `skills/<name>/CHANGELOG.md` — entrada nova documentando o import.
- `roster/skills.yaml` — atualizar version + provides.
- `docs/consumer-guide.md` — atualizar tabela de skills.
- `CHANGELOG.md` (raiz) — entrada resumindo o batch.

### 9. Cross-compile

`.goreleaser.yaml` — adicionar `builds:` + `archives:` para o novo
`cmd/<name>/`. Mesmo shape que `treasure-chest`:

```yaml
builds:
  - id: <name>
    main: ./cmd/<name>
    binary: <name>
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    env: [CGO_ENABLED=0]
    flags: [-trimpath]
    ldflags: [-s -w -X main.Version={{.Version}}]
    mod_timestamp: "{{ .CommitTimestamp }}"

archives:
  - id: <name>-binary
    ids: [<name>]
    formats: [binary]
    name_template: "<name>-{{ .Os }}-{{ .Arch }}"
```

### 10. Regenerar lockfile

```bash
make ci-lint
make lock-skills   # sobrescreve skillhire.lock com os novos digests
make lock-verify   # deve passar agora
make ci-test       # full
make lint          # 0 issues
```

### 11. Security review

Import de código de terceiro merece uma passada explícita:

```bash
# Em Claude Code:
/security-review
```

Falhas HIGH/MEDIUM têm que ser resolvidas antes de merge. LOW pode ir
para issue de follow-up.

### 12. Snapshot local do release

Confirma que o cross-compile passa e que o binário roda:

```bash
goreleaser release --snapshot --clean --skip=publish
ls dist/ | grep <name>
./dist/<name>_linux_amd64_v1/<name> status --root /tmp --json | head
```

### 13. Commit final + PR

```bash
git add -A
git commit -m "feat(<name>): import runtime from source-repo at <version>"
git push origin feat/<name>-runtime
```

Abrir PR contra `develop`. Descrever no PR:

- Origem: `owner/source-repo@<sha>`, pasta `<folder>/`
- Volume: N arquivos, ~M LOC
- Acoplamento severado: quais símbolos foram para `runtime/domain/vocab.go`
- O que ficou de fora e por quê (tipicamente `cli/` do fonte)
- Resultado da security review

## Verificação

- [ ] Dry-run commit preservado em branch `dry-run/<name>-probe`
- [ ] Adapter implementa `KnowledgeProvider` (compile-time guard passa)
- [ ] CLI standalone com testes (mínimo: status human+JSON, prepare
      vazio, search vazio, explain unknown)
- [ ] `.golangci.yaml` proíbe imports do repo-fonte
- [ ] `roster/skills.yaml` + `docs/consumer-guide.md` refletem a nova
      versão
- [ ] `skillhire.lock` regenerado
- [ ] `goreleaser release --snapshot` produz 6 binários novos
- [ ] `make lint` e `make ci-test` verdes
- [ ] Security review sem findings HIGH/MEDIUM

## Rollback

Antes de merge — o commit dry-run em `dry-run/<name>-probe` é seu
snapshot puro. Se o adapter/CLI ficarem inviáveis, reseta a branch de
feature para o dry-run:

```bash
git checkout feat/<name>-runtime
git reset --hard dry-run/<name>-probe
```

E retoma do passo 5.

Depois de merge, antes de tag — reverter o commit de merge.

Depois de tag — cortar tag nova removendo o `builds:` do
`.goreleaser.yaml` e ajustando o `roster/skills.yaml` para
`lifecycle: retired`. Assets antigos permanecem no GitHub Release
antigo; o novo release não os inclui.

## Referências

- [ADR real do case treasure-chest] (em Providence, ADR-0033)
- [`docs/consumer-guide.md`](../consumer-guide.md) § 1.2 — o
  contrato observável do consumidor
- [`platform/knowledge-api/`](../../platform/knowledge-api/) — o
  contrato Go que o adapter satisfaz
- [`skills/treasure-chest/`](../../skills/treasure-chest/) — o
  exemplo canônico end-to-end
