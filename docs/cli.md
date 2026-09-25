# CLI — `skillhire`

Superfície mínima planejada, sujeita a refinamento antes da implementação
em Go (`cmd/skillhire/`).

```text
skillhire
├── search           # busca no roster
├── info             # detalhes de uma skill
├── list             # skills instaladas ou disponíveis
├── install          # instala uma ou mais skills
├── remove           # remove uma skill instalada
├── update           # atualiza skills
├── verify           # valida checksums e lockfile
├── doctor           # diagnóstico do ambiente
├── company
│   ├── list
│   ├── info
│   └── install
├── roster
│   ├── list
│   ├── add
│   ├── remove
│   └── update
├── publish          # grupo para mantenedores
│   ├── validate
│   ├── pack
│   ├── release
│   └── index
└── completion
```

## Flags globais

```text
--config <path>       arquivo de configuração
--target <client>     cliente de destino (Claude Code, Codex, etc.)
--scope <scope>       project | user
--registry <name>     roster selecionado
--json                saída estruturada
--quiet               apenas erros e resultados essenciais
--yes                 confirma operações não interativas
--dry-run             mostra mudanças sem aplicá-las
--locked              exige correspondência com o lockfile
--offline             proíbe acesso à rede
```

## Aliases temáticos (opcionais)

- `recruit` → `install`
- `release` (contexto de skill) permanece reservado para publicação; a
  remoção continua sendo `remove`.

## Regras de design

1. Nenhuma instalação implícita durante execução de uma skill.
2. Operações reversíveis (`--dry-run`, mensagens claras).
3. Reprodutibilidade via `skillhire.lock`.
4. Adaptação por cliente feita pela CLI, sem alterar a fonte.
5. Automação first-class (`--json`, `--yes`, códigos de saída
   previsíveis).
