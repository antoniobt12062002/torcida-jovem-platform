# Evidência — revisão estrutural do front (web-design)

- **Data**: 2026-10-12
- **Unidade**: DS-INT (T12 de `tasks/05-migracao-web-v1.md`), requisitos INT-02, DSM-01 AC2, WDS-D-029
- **Ponto conferido**: `663febe` (T11 sobre `445bc68`, rodadas 1 a 6 integradas). O grafo registra `built_at_commit = 663febe…`.
- **Ferramenta**: Graphify local, `C:/Users/Michels/Desktop/torcida-jovem/.claude/tools/graphify/.venv/Scripts/graphify.exe`. `--version` conferido antes do uso: `graphify 0.9.86`.
- **Fonte de verdade**: os `import` reais do código, conferidos com `grep`. O grafo é auxiliar (`docs/development/ai-environment.md`), e specs e `STATE.md` têm precedência.

## Como o grafo foi gerado

A partir da raiz do worktree da integração (`tjw/ds-int`), com o registro de consultas desligado:

```bash
G=C:/Users/Michels/Desktop/torcida-jovem/.claude/tools/graphify/.venv/Scripts/graphify.exe
"$G" --version                                   # graphify 0.9.86
GRAPHIFY_QUERY_LOG_DISABLE=1 "$G" extract . --code-only
```

- Só AST (tree-sitter), sem modelo de linguagem, sem chave de API, sem rede. Nenhum hook, MCP (local ou remoto), serviço ou alteração em `.claude/`, `CLAUDE.md` ou configuração foi feita.
- Saída em `graphify-out/` (`graph.json`, `manifest.json`, `cache/`), ignorada pelo git e não versionada. Não se rodou `cluster-only` (o relatório de comunidades não é necessário para esta revisão).
- Resultado: 600 arquivos de código, 5572 nós, 24206 arestas (23036 `EXTRACTED`, 1170 `INFERRED`). Em `web/`: 1836 nós de 302 arquivos.
- A leitura foi feita por um script descartável fora do repositório: cada aresta é projetada no par (arquivo de origem, arquivo de destino), e as relações abaixo são filtradas por diretório.

## Relações conferidas

| # | Relação esperada | Grafo (`EXTRACTED`) | `import` real (`grep`) | Resultado |
| --- | --- | --- | --- | --- |
| 1 | Nenhuma dependência entre `features/identity`, `features/financeiro` e `features/estoque` | Nenhuma aresta entre os três módulos | Nenhum `@/features/<outro>` nem caminho relativo para outro módulo em nenhum dos três | Confirmado |
| 2 | `components/app` não importa `features`, exceto `nav-items.ts` | Só `nav-items.ts` → `features/{estoque,financeiro,identity}/nav.ts`, e o teste dele (`nav-items.test.ts`) para os mesmos três arquivos | Os mesmos seis `import` (`estoqueNav`, `financeiroNav`, `identityNavSection`), em `nav-items.ts` e `nav-items.test.ts` | Confirmado. O teste de `nav-items.ts` importa o que o próprio arquivo importa |
| 3 | `components/ui` depende só de `lib/utils` (e de pacotes) | Código do kit → `lib/utils.ts` e `package.json` (pacotes); fora do kit, só os testes apontam para `test/design/guard.ts` | Código do kit: só `@/components/ui/*` (entre arquivos do kit) e `@/lib/utils` (20 arquivos). Pacotes: `@base-ui/react/*`, `class-variance-authority`, `cn`, `lucide-react`, `next/link`, `react`. Testes: `@/test/design/guard` em `data.test.tsx` e `tokens.guard.test.ts` | Confirmado |
| 4 | `lib/url-state` e `lib/responsive` sem depender de `features` ou `components` | Nenhuma aresta para `features/` ou `components/` | Código: só `./codecs`, `./use-query-state`, `./breakpoints`, `./use-media-query` e pacotes (`next/navigation`, `react`). Testes: `@/lib/session/navigation-mock` e `@/test/media` | Confirmado |

Observações:

- `button.tsx`, `badge.tsx` e `icon-button.tsx` importam `cn` direto do pacote `cn` (dependência declarada em `web/package.json`), e os outros 20 arquivos do kit importam de `@/lib/utils`, que só reexporta o mesmo pacote (`export { cn } from "cn"`). As duas formas cabem na regra. Unificar fica como melhoria opcional, fora desta task.
- Nenhuma dependência nova entre módulos em relação ao diagnóstico de 2026-10-11 (`design/05-migracao-web-v1.md`, DD-6): as duas únicas arestas entre módulos continuam `INFERRED` e são as mesmas coincidências de nome entre testes (próxima seção).

## Arestas `INFERRED` descartadas

Em `web/` há 62 arestas `INFERRED`, todas `indirect_call`. Nenhuma conta como dependência. As que cruzam as fronteiras conferidas:

- `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` → `row()` de `features/financeiro/contas/contas-page.test.tsx`: coincidência de nome. O teste do estoque não importa nada do financeiro.
- `features/identity/usuarios/dialogs/dialog-patterns.test.tsx` → `confirm()` de `features/estoque/ajustes/ajuste-dialog.tsx`: coincidência com o campo `confirm` da tabela de casos do teste da identidade. Não há `import` do estoque.

A única `INFERRED` fora de testes (`app/(app)/inicio/atalhos.tsx` → `isNavSection()` de `components/app/nav.ts`) é de uma tela para um composto, direção permitida, e não muda nenhuma conclusão.

## Limites do grafo

- **Arestas `INFERRED` não provam dependência** e foram descartadas. Só contam `EXTRACTED` confirmadas pelo `import` real.
- **`.d.ts` fora das relações**: os quatro arquivos gerados em `web/lib/api/*.d.ts` aparecem como nós (7 cada), mas **nenhuma aresta chega a eles**. Quem importa os tipos gerados da API fica invisível no grafo. Essas importações não fazem parte das relações conferidas.
- **Ausência de aresta não prova independência.** Por isso cada linha da tabela foi confirmada também por `grep`, inclusive os caminhos relativos (`../`).
- Importações de pacotes viram arestas para `web/package.json`, sem nomear o pacote. A lista de pacotes acima vem do `grep`.
- CSS, tokens (`app/globals.css`), `import()` dinâmico e `vi.mock` só são vistos parcialmente ou não são vistos. Essas relações ficam com as guardas (`web/test/design/`) e com a leitura do código.
- O grafo é um retrato de `663febe` neste worktree. Não substitui testes, gates, guardas, o smoke (T13) nem o verificador independente (T14).

## Comandos de conferência com `grep` (a partir de `web/`)

```bash
# 1. Entre módulos (alias e caminhos relativos)
for m in identity financeiro estoque; do
  grep -rnE "[\"']@/features/" features/$m | grep -vE "@/features/$m/"
  grep -rnE "[\"'](\.\./)+(identity|financeiro|estoque)/" features/$m | grep -vE "/$m/"
done                                                     # vazio
# 2. components/app -> features
grep -rnE "[\"']@/features/|[\"'](\.\./)+features/" components/app   # só nav-items.ts e nav-items.test.ts
# 3. Kit: importações internas fora do kit, sem testes
grep -rnE "from\s+[\"'](@/|\.)" components/ui | grep -v "\.test\." | grep -vE "@/components/ui/|@/lib/utils"   # vazio
# 4. lib/url-state e lib/responsive
grep -rnE "from\s+[\"'](@/|\.)" lib/url-state lib/responsive | grep -v "\.test\."   # só caminhos ./ internos
```
