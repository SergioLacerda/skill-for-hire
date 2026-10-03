# Runbook — cortar tag `v*.*.*`

Publica uma release oficial disparando o workflow `Release` do
GitHub Actions, que produz binários, SHA256SUMS, skill packs, SBOM e
build-provenance attestations.

## Quando usar

- Terminou uma sequência de PRs mergeados em `main` que juntos formam
  uma release coerente.
- Todos os bumps de skill relevantes já estão em `main` (ver
  `bump-skill-version.md`).

## Pré-requisitos

- Você tem permissão de push para tags em
  `SergioLacerda/skill-for-hire` no GitHub (o CI-web da sessão
  Claude Code **não** consegue pushar tags — bloqueio de policy).
- `main` está em CI verde no commit alvo.

## Passos

### 1. Alinhar local com `main`

```bash
git fetch origin
git checkout main
git pull --ff-only origin main
git log --oneline -5              # confirmar que o HEAD é o esperado
```

### 2. Sanity local do release

```bash
export GOCACHE=/tmp/go-build-cache GOMODCACHE=/tmp/go-mod-cache
make ci-lint
make ci-test
make release-dry-run              # instala goreleaser + snapshot completo
```

Se qualquer um falhar: **não** corte a tag. Volte, abra PR de fix.

### 3. Escolher versão

- **`v0.0.<N+1>`** — patch / bump de skill / correção pequena
- **`v0.<N+1>.0`** — nova capability ou nova skill
- **`v<N+1>.0.0`** — breaking change no contrato Knowledge API ou
  no schema `skill.yaml`

### 4. Criar tag anotada

Mensagem de tag deve resumir o que a release contém:

```bash
NEW_TAG=v0.0.2
git tag -a "$NEW_TAG" -m "$(cat <<EOF
$NEW_TAG — <resumo em uma linha>

Skills nesta release:
- atlas@0.1.0  (contrato-only, sem runtime)
- treasure-chest@0.2.0  (runtime importado do strategist-skill; KA v1
  end-to-end; binário standalone cross-compiled)

Skills for Hire changes:
- <highlights, ex. Wave 1-4, rankeamento real, adapter Strategist>

Consumidores: ver docs/consumer-guide.md.
EOF
)"
```

Confirme o alvo antes de pushar:

```bash
git show --stat "$NEW_TAG" | head
```

### 5. Push da tag

```bash
git push origin "$NEW_TAG"
```

Isso dispara o workflow `Release`.

### 6. Monitorar o workflow

- GitHub Actions: https://github.com/SergioLacerda/skill-for-hire/actions
- Aguarde o job `verify` passar (roda `make ci` no commit tagueado).
- Job `release` roda goreleaser + Anchore SBOM + duas
  `attest-build-provenance` (uma por `SHA256SUMS` dos binários, outra
  por path dos skill packs).

### 7. Validar assets publicados

```bash
gh release view "$NEW_TAG" --repo SergioLacerda/skill-for-hire
```

Deve listar (contagens para uma release com 2 skills e a matriz atual):

- 6 × `skillhire-<os>-<arch>[.exe]`
- 6 × `<name>-<os>-<arch>[.exe]` por skill com runtime
- 1 × `SHA256SUMS`
- 3 × `<skill>-<version>.{tar.gz,tar.gz.sha256,release.yaml}` por skill
- 1 × `skillhire-<tag>-sbom.cdx.json`

Baixe um asset e verifique digest:

```bash
BASE="https://github.com/SergioLacerda/skill-for-hire/releases/download/$NEW_TAG"
curl -sSLO "${BASE}/treasure-chest-0.2.0.tar.gz"
curl -sSLO "${BASE}/treasure-chest-0.2.0.tar.gz.sha256"
sha256sum -c treasure-chest-0.2.0.tar.gz.sha256
```

E que o attestation resolve:

```bash
gh attestation verify treasure-chest-0.2.0.tar.gz \
  --repo SergioLacerda/skill-for-hire
```

### 8. Verificar reprodutibilidade

Confirmar que os digests que vieram do CI batem com os do
`skillhire.lock` no commit tagueado:

```bash
git show "$NEW_TAG":skillhire.lock | grep digest
```

Compare com as linhas do `SHA256SUMS` publicado — os digests dos packs
têm que bater bit-a-bit.

## Verificação

- [ ] Workflow `Release` completou com `success` em `verify` e `release`
- [ ] Todos os assets esperados estão no GitHub Release
- [ ] `sha256sum -c` de pelo menos um pack passa
- [ ] `gh attestation verify` de um asset resolve para o commit tagueado
- [ ] Digests dos packs no release batem com `skillhire.lock` do commit

## Rollback

**Antes de anunciar publicamente** — se algum asset estiver quebrado
ou faltando:

```bash
gh release delete "$NEW_TAG" --repo SergioLacerda/skill-for-hire --yes
git push origin --delete "$NEW_TAG"
git tag -d "$NEW_TAG"
```

Depois corrija no `main` e volte ao passo 4 com a **mesma** versão (ou
suba pra `<tag>+1` se preferir).

**Depois de anunciar** — não delete. Corte tag nova (`v0.0.<N+1>`)
com o fix. Documente no CHANGELOG raiz que a `<tag>` original foi
substituída.
