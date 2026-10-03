# Runbook — bump da versão de uma skill

Publica uma nova versão de uma skill existente (ex.
`treasure-chest@0.2.0` → `0.2.1` para patch, `0.3.0` para minor,
`1.0.0` para major).

## Quando usar

- Alterações em código/contrato de uma skill que já foi shipada em
  alguma tag anterior.
- SemVer estritamente: PATCH = fix compatível; MINOR = capability nova
  compatível; MAJOR = breaking change no contrato ou envelope.

## Pré-requisitos

- Mudanças na skill já testadas localmente (`make ci-test` verde).
- Você sabe qual bump aplicar (PATCH/MINOR/MAJOR).

## Passos

### 1. Branch

```bash
git checkout develop
git pull --ff-only
git checkout -b feat/<name>-<newversion>
```

### 2. Bump em cada lugar

Todos precisam bater — o `lock-verify` do CI falha se qualquer um
divergir:

- [ ] `skills/<name>/skill.yaml` → `metadata.version: "<new>"`
- [ ] `skills/<name>/SKILL.md` → `metadata.version: <new>` (no
      frontmatter)
- [ ] `skills/<name>/CHANGELOG.md` → nova seção `## [<new>] — YYYY-MM-DD`
- [ ] `roster/skills.yaml` → entrada da skill, `version: "<new>"`
- [ ] `docs/consumer-guide.md` → tabela de skills, coluna versão
- [ ] `CHANGELOG.md` (raiz) → entrada Unreleased descrevendo o batch

### 3. Regenerar lockfile

```bash
make ci-test          # falha em lock-verify — esperado
make lock-skills      # sobrescreve com os novos digests
make lock-verify      # agora passa
```

Inspecione `git diff skillhire.lock` — só as linhas `version:`,
`digest:`, `size:`, `resolved-from:` da skill afetada devem mudar. Se
outra skill mudou digest sem razão, investigue antes de commitar (ver
seção "Rollback" abaixo).

### 4. Testes finais

```bash
make ci-lint
make lint             # 0 issues
```

### 5. Commit

```bash
git add -A
git commit -m "feat(<name>): bump to <new>"
git push -u origin feat/<name>-<newversion>
```

Abrir PR contra `develop`.

## Verificação

- [ ] Todos os 6 lugares acima bumped
- [ ] `skillhire.lock` mostra digest novo **só** para a skill afetada
- [ ] `make ci-lint && make ci-test && make lint` verdes

## Rollback

Se o pack digest de outra skill mudou sem justificativa (bug de
reprodutibilidade), NÃO commite o lockfile. Investigue:

```bash
git stash                                    # tira as mudanças de lado
git checkout main -- skillhire.lock          # lockfile antigo
rm -rf dist packs
make pack-skills
diff <(cat skillhire.lock) <(./bin/skillhire lock dist/*.release.yaml --out - 2>/dev/null || cat skillhire.lock)
```

Se a diferença persistir mesmo sem tocar em nada, o bug está no
`internal/pack/` — abra issue e não faça o bump.
