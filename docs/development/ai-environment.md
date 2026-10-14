# Ambiente de desenvolvimento com IA

As skills de IA são ferramentas do ambiente de cada desenvolvedor. **Não são versionadas** neste repositório (as pastas `.claude/`, `.cursor/`, `.windsurf/` e `.agents/` estão no `.gitignore`). Este documento registra quais são usadas, para quê e quando acioná-las, para que o ambiente seja reproduzível.

## Skills utilizadas

| Skill | Finalidade | Quando acionar |
|---|---|---|
| `tlc-spec-driven` | Fluxo Spec Driven: Specify, Design, Tasks, Implement, Validate, com validadores e verificador independente | Ao iniciar qualquer funcionalidade; mantém `.specs/` |
| `technical-design-doc-creator` | Design docs técnicos completos, por descoberta interativa | Temas transversais ou complexos (autenticação, auditoria, integração de pagamento) antes de implementar |
| `create-adr` | Registro de decisões de arquitetura no formato MADR em `docs/adr/` | Sempre que uma decisão relevante for tomada; nunca para decisões ainda em aberto |
| `security-best-practices` | Revisão de segurança para Go e JavaScript/TypeScript | No design e antes de abrir PR que toque autenticação, sessões, dinheiro, uploads ou entrada de usuário |
| `the-judge` | Revisão de PR baseada em evidências, publicada como review no GitHub | Antes de mesclar qualquer PR relevante |
| `spec-driven-eval` | Avaliação da qualidade do processo Spec Driven | Somente por invocação explícita do mantenedor (`/spec-driven-eval`); o agente não a executa por conta própria |
| `vercel-deploy` | Deploy na Vercel | Não usada agora; só se a hospedagem do front for a Vercel |
| `frontend-blueprint` | Descoberta de design antes do código: briefing, referências visuais, direção de design (paleta, tipografia, layout, ícones) aprovada pelo mantenedor | No Specify e no Design de ciclos de interface (Design System, telas novas, redesenho); a construção segue as tasks do `tlc-spec-driven`, não as fases de build da própria skill |
| `web-design-guidelines` | Revisão de código de interface contra as Web Interface Guidelines (acessibilidade, foco, formulários, estados, hidratação), com achados em `arquivo:linha` | Na verificação de ciclos de interface e antes de abrir PR que altere telas; só leitura |

O plugin `superpowers` (brainstorming, planos, execução) foi usado nas fases iniciais; os artefatos resultantes estão arquivados em [`docs/archive/superpowers/`](../archive/superpowers/plans/2026-09-26-github-cicd.md). A partir da governança, o fluxo oficial é o do `tlc-spec-driven`.

## Instalação

```bash
npx @tech-leads-club/agent-skills install --skill technical-design-doc-creator
npx @tech-leads-club/agent-skills install --skill the-judge
npx @tech-leads-club/agent-skills install --skill spec-driven-eval
npx @tech-leads-club/agent-skills install --skill create-adr
npx @tech-leads-club/agent-skills install --skill frontend-blueprint web-design-guidelines --agent claude-code
```

O pacote `@tech-leads-club/agent-skills` 1.5.0 declara Node >= 24; com o Node 22 do projeto ele instala com um aviso de `engines`.

`tlc-spec-driven`, `security-best-practices` e `vercel-deploy` já constam no lockfile local do ambiente (`.agents/.skill-lock.json`); instale-as da mesma forma, pelo mesmo pacote, se estiverem ausentes.

## Regras de uso

- Decisões arquiteturais viram ADR; decisões de projeto que valem para todas as features entram também em `.specs/STATE.md`.
- Nenhuma skill grava segredos ou dados pessoais no repositório (ele é público).
- Ações remotas (push, PR, merge, deploy) exigem autorização explícita do mantenedor; a aprovação de uma spec autoriza apenas implementação e commits locais.
- O `frontend-blueprint` alimenta o Specify e o Design; o processo continua sendo o do `tlc-spec-driven` (spec, design, tasks, implementação e verificador independente). A integração opcional com o Google Stitch (servidor MCP remoto com chave de API) não é usada sem autorização do mantenedor, e a chave nunca entra no repositório.
- As regras do `web-design-guidelines` são parâmetro de revisão, não requisito: decisões registradas nas specs e em `STATE.md` prevalecem (por exemplo, interface em português sem Title Case e nenhuma dependência nova sem justificativa).
- O `the-judge` publica a revisão com a conta do mantenedor; como o GitHub não permite aprovar o próprio PR, o resultado funciona como comentário, e não como aprovação formal.
- Instruções para os agentes ficam no [`CLAUDE.md`](../../CLAUDE.md) da raiz.

## Graphify (análise estrutural)

O [Graphify](https://github.com/Graphify-Labs/graphify) (pacote `graphifyy`, licença Apache-2.0) gera um grafo do código por análise sintática local. É ferramenta **auxiliar** do `tlc-spec-driven`, para análise de dependências, impacto de mudanças, relações entre componentes, testes relacionados, apoio à propriedade de arquivos entre agentes e revisão estrutural antes de PR.

**Specs e `STATE.md` têm precedência sobre o Graphify.** A hierarquia é: Specs/STATE → Design → Tasks → ferramentas auxiliares → código → testes → Verification. O Graphify está entre as ferramentas auxiliares e não é autoridade sobre decisões do projeto. O grafo é análise, não prova de correção funcional, e seus resultados não viram requisito. Ele não substitui `validate_spec`, `validate_tasks`, testes, gates, smoke, `web-design-guidelines`, o verificador independente nem a revisão humana.

Quando acionar: análise estrutural, de dependências e de impacto; descoberta de quem chama ou importa um arquivo; investigação de componentes compartilhados; planejamento de propriedade de arquivos; verificação de dependências entre tasks. Não é passo obrigatório para perguntas comuns sobre o código.

### Instalação em escopo de projeto

Nada é versionado: o ambiente virtual e a skill ficam em `.claude/` e a saída em `graphify-out/`, ambos no `.gitignore`.

```bash
# CLI num ambiente virtual do projeto (versão fixada)
uv venv .claude/tools/graphify/.venv --python 3.14
uv pip install --python .claude/tools/graphify/.venv/Scripts/python.exe "graphifyy==0.9.86"

# Só a skill, em .claude/skills/graphify/ (sem hooks e sem regras permanentes)
.claude/tools/graphify/.venv/Scripts/python.exe -c "from pathlib import Path; from graphify.install import install; install(platform='claude', project=True, project_dir=Path('.'))"
```

**Não usar `graphify install --project` nem `graphify claude install`.** Na versão 0.9.86 esses comandos, além da skill, acrescentam uma seção `## graphify` ao `CLAUDE.md` versionado da raiz ("para perguntas sobre o código, rode primeiro `graphify query`") e gravam hooks `PreToolUse` em `.claude/settings.json` que interceptam buscas e leituras de arquivo. As duas coisas contrariam a precedência acima. Também não se usa `graphify hook install` (hooks de git que reconstroem o grafo a cada commit).

### Personalização local da skill

A descrição original da skill manda tratar qualquer pergunta sobre o código como consulta ao grafo. Em `.claude/skills/graphify/SKILL.md` duas coisas foram alteradas, e só elas:

1. O campo `description` do cabeçalho, que passa a restringir o acionamento a `/graphify` ou a pedido explícito de análise estrutural.
2. Um bloco "Regras deste projeto", logo abaixo do cabeçalho, com a hierarquia acima, o uso do executável local, o modo só local e as limitações.

O restante do arquivo e as referências são os do pacote. **Reinstalar ou atualizar a skill sobrescreve a personalização**: depois de qualquer reinstalação, reaplique os dois pontos. O original está em `.claude/tools/graphify/.venv/Lib/site-packages/graphify/skill.md`, para comparação.

### Versão e executável

A versão do projeto é a **0.9.86** do ambiente virtual local. Pode existir outra instalação no `PATH` da máquina (por exemplo uma global mais antiga, com servidor MCP próprio): ela não é usada neste projeto, nem o servidor MCP global `graphify`. Resolva o executável pelo caminho e confirme a versão antes de qualquer comando.

PowerShell:

```powershell
$G = Join-Path (git rev-parse --show-toplevel) ".claude\tools\graphify\.venv\Scripts\graphify.exe"
if (-not (Test-Path $G)) { throw "Graphify local ausente: refaça a instalação em escopo de projeto" }
$v = & $G --version
if ($v -ne "graphify 0.9.86") { throw "Versão inesperada: $v" }
```

Git Bash:

```bash
G="$(git rev-parse --show-toplevel)/.claude/tools/graphify/.venv/Scripts/graphify.exe"
[ "$("$G" --version)" = "graphify 0.9.86" ] || { echo "Graphify local ausente ou em versão inesperada"; exit 1; }
```

### Uso

Com `$G` resolvido como acima (em PowerShell, prefixe com `&`):

```powershell
& $G extract . --code-only      # gera ou regenera graphify-out/graph.json (só AST, sem LLM)
& $G update .                   # atualização incremental depois de mudar código
& $G affected "button.tsx"      # quem é afetado por uma mudança
& $G path "A" "B"               # caminho entre dois nós
& $G god-nodes --top 15         # nós mais conectados
```

Sempre `--code-only`: sem esse modo, documentos e imagens passariam por extração semântica com um modelo de linguagem.

### Atualização do grafo

Manual e controlada, sem hooks. O grafo é um retrato do momento em que foi gerado.

- Momentos de atualização: antes de fechar a DAG de tasks; depois de integrar unidades que alterem dependências; antes da revisão estrutural final.
- O coordenador é o responsável pelo retrato consolidado em `graphify-out/` da árvore principal.
- Agentes em execução paralela não atualizam o grafo. Cada worktree tem a própria pasta `graphify-out/` ignorada; se um agente precisar consultar, usa o retrato do coordenador só para leitura.

### Comunicação externa

- A extração de código é local (tree-sitter) e não usa chave de API. O pacote declara não ter telemetria.
- Não se usa o Graphify Cloud (`app.graphify.com`), OAuth, MCP remoto, `graphify add <url>`, `graphify clone` nem os modos com LLM (`extract` sem `--code-only`, `label`, `--mode deep`) sem autorização do mantenedor.
- O registro de consultas em `~/.cache/graphify-queries.log` só existe se for ligado por variável de ambiente; não é ligado.

### Limitações conhecidas (verificadas em 2026-10-11)

**Nenhuma decisão de paralelização depende só do grafo.**

- Arestas `INFERRED` não são prova de dependência. Em `web/` todas as 42 encontradas eram `indirect_call` por coincidência de nomes entre arquivos de teste.
- Para propriedade de arquivos, valem relações `EXTRACTED` confirmadas com o `import` real no código.
- Arquivos `.d.ts` não estão representados. Importações dos tipos gerados da API (`web/lib/api/*.d.ts`) ficam invisíveis.
- Dependências em Go aparecem ampliadas: a importação de um pacote vira aresta para todos os arquivos dele. Chamadas por método de campo (por exemplo `Authz.Require`) não são ligadas à função. A fronteira entre módulos da API continua sendo verificada por `api/internal/architecture_test.go`.
- Chamadas dinâmicas e dependências não detectadas continuam exigindo leitura de código.
- **Ausência de aresta não comprova independência.**
- Arquivos `.sql` (migrations) não são analisados sem o extra `graphifyy[sql]`, que não foi instalado.
- CSS e comportamento em tempo de execução estão fora do grafo.
