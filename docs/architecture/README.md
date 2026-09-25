# Arquitetura — Skills for Hire

Este diretório é o ponto de entrada para a filosofia e o padrão do
ecossistema. Cada documento tem foco distinto:

| Documento | Escopo |
| --- | --- |
| `01-ecosystem-philosophy.md` | Filosofia e arquitetura geral do ecossistema (monorepo, catálogo federado, bundles, releases individuais). |
| `02-cli-identity.md` | Identidade do produto e superfície da CLI `skillhire`. |
| `03-skill-standard.md` | Padrão canônico de uma skill (`SKILL.md`, `skill.yaml`, perfis de maturidade, pipeline de validação e publicação). |
| `04-ai-first-guidelines.md` | Orientações IA-first para Atlas e Treasure Chest — plataforma compartilhada, standalone, adapters, packaging imutável. |
| `05-atlas-cartographer.md` | Refinamento do Cartógrafo e do Project Atlas: separação de responsabilidades, versionamento pelo commit, pipeline própria, MVPs. |

Os arquivos serão importados a partir das notas de design. Enquanto isso,
os princípios operacionais consolidados estão no `README.md` da raiz e no
padrão de skill vigente (ORKA + Skills for Hire), materializado nos
templates `skills/atlas/` e `skills/treasure-chest/`.

## Camadas do padrão

1. **Contrato ORKA (portátil)** — o pacote publicável: raiz em kebab-case,
   `SKILL.md` obrigatório com frontmatter, pastas opcionais
   `references/`, `scripts/`, `templates/`, `assets/`.
2. **Contrato de projeto (adaptador)** — `skill.yaml` com identidade,
   versão, capabilities, permissões (deny-by-default), lifecycle,
   dependências, distribuição e conformance.
3. **Camada de plataforma** — Knowledge API, Source Registry, Document
   Model, Lifecycle e Packaging compartilhados por todas as skills, sem
   fundir seus domínios internos.

## Níveis de conformance

- `package-static` — estrutura ORKA, frontmatter, seções e outputs
  declarados.
- `project-contract` — adaptador, permissões, lifecycle, outputs e
  ownership validados contra o contrato do projeto.
- `live-invocation` — invocação real com evidência de permissões,
  saída, fallback e observabilidade.

Um nível inferior nunca é apresentado como prova do seguinte.
