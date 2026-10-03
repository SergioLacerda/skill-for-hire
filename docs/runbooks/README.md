# Runbooks

Procedimentos passo-a-passo para tarefas recorrentes deste repositório.

| Runbook | Quando usar |
| --- | --- |
| [`new-skill-with-runtime.md`](new-skill-with-runtime.md) | Criar uma nova skill do zero com runtime Go executável, adapter Knowledge API e CLI standalone (padrão "estilo B" — atlas, futuras skills). |
| [`migrate-external-skill.md`](migrate-external-skill.md) | Importar um domínio existente de outro repositório para dentro de `skills/`, desacoplando e adaptando ao contrato Skills for Hire. Baseado na experiência real de importar `treasure-chest` do `strategist-skill`. |
| [`bump-skill-version.md`](bump-skill-version.md) | Publicar uma nova versão de uma skill existente e regenerar o lockfile do repo. |
| [`cut-release-tag.md`](cut-release-tag.md) | Cortar uma tag `v*.*.*` do `main`, disparar o release workflow, validar os assets publicados. |

## Estrutura de um runbook

Cada runbook segue o mesmo formato:

1. **Quando usar** — o gatilho que faz esse runbook aplicável.
2. **Pré-requisitos** — o que precisa estar verdadeiro antes.
3. **Passos** — comandos exatos e diffs / arquivos a criar.
4. **Verificação** — como saber que deu certo.
5. **Rollback** — como reverter se algo quebrar.

## Convenções

- Placeholders: `<name>` (kebab-case identifier da skill), `<version>` (SemVer sem `v` — ex. `0.2.0`), `<company>` (bundle name).
- Todo runbook parte de `develop` limpo (`git status` sem modified/untracked).
- Todo runbook termina com `make ci-lint && make ci-test && make lint` verdes antes do commit final.
