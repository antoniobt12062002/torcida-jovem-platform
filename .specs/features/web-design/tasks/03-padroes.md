# Padrões Compartilhados — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos critérios de aceite; um commit atômico por task; a task é marcada como concluída neste arquivo antes do commit e no mesmo commit. **Se a skill não puder ser ativada, PARE e avise.**

Antes de escrever código Next: ler `web/AGENTS.md` e a documentação do pacote instalado em `web/node_modules/next/dist/docs/` (em especial `use-search-params.md` e a seção de `window.history` em `04-linking-and-navigating.md`).

Autorização: a aprovação destas tasks autoriza implementação e commits locais na branch da unidade. `git push`, PR, merge e deploy exigem autorização explícita para cada ação.

---

**Specs**: `.specs/features/web-design/spec/03-padroes.md`; os componentes de domínio e do shell atendem também `spec/02-componentes.md` e `spec/04-shell-e-acesso.md`.
**Design**: `.specs/features/web-design/design/03-padroes.md`, `design/02-componentes.md` (domínio) e `design/04-shell-e-acesso.md` (shell).
**Status**: Draft (2026-10-11), aguardando aprovação do mantenedor.

**Pré-requisito**: DS-K e DS-T integradas (`tasks/02-componentes.md`).

**Unidades de execução** (propriedade e DAG em `design/05-migracao-web-v1.md`):

| Unidade | Tasks | Rodada | Entrada | Saída |
| --- | --- | --- | --- | --- |
| DS-L Bibliotecas auxiliares | T1, T2 | 3 | DS-T integrada (mock reativo e `mockViewport`) | Estado na URL e consulta de mídia, com testes de servidor |
| DS-A Componentes compostos | T3 a T10 | 4 | DS-K, DS-T e DS-L integradas | `components/app` com página, estados, formulário, domínio, listagem e shell; guarda passando |

As duas unidades são sequenciais e podem ser do mesmo agente ou de agentes diferentes, em rodadas separadas.

## Restrições comuns a todas as tasks deste arquivo

- Só os arquivos da unidade são alterados (gate Propriedade). `components/app/nav-items.ts` é de DS-S e não é tocado aqui. `package.json` e `pnpm-lock.yaml` não mudam.
- **Compatibilidade**: propriedades e assinaturas em uso pelas telas não migradas continuam aceitas até DS-INT. Os testes existentes de `features/` e `app/` passam sem alteração.
- `components/app` não importa de `features/`. `StatusBadge` deixa de importar tipos de módulo.
- Sem biblioteca nova: estado na URL só com `useSearchParams` e a API de histórico; consulta de mídia só com `matchMedia`.
- Nenhuma mudança em `lib/api/`, `lib/forms/`, `lib/session/` (fora dos arquivos de DS-T) nem em regra de negócio ou de permissão.

## Test Coverage Matrix

> Gerada do código, das diretrizes do projeto e da spec. Diretrizes consultadas: `CLAUDE.md`, `web/AGENTS.md`, `docs/development/web.md`, `web/vitest.config.ts`, `.specs/features/web/tasks/*.md`.

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `lib/url-state/`, `lib/responsive/` | unit | Cada AC de DSP-05 ligado à URL; valores inválidos; renderização no servidor igual à primeira renderização do cliente | `web/lib/**/*.test.ts(x)` | `pnpm test` |
| `components/app/` | component | Cada AC da spec atendido pelo componente; papel, nome e teclado; comportamento em largura estreita com `mockViewport` | `web/components/app/*.test.tsx` | `pnpm test` |
| Guarda de `components/app` | unit | Nenhum valor solto | `web/components/app/*.guard.test.ts` | `pnpm test` |

## Gate Check Commands

Todos rodam a partir de `web/`, com o código de saída capturado de verdade. Um gate só passa com saída `0`.

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Contract | toda task | `pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Dependências intactas | toda task | `git diff --name-only develop -- web/package.json web/pnpm-lock.yaml` vazio |
| Ownership | toda task | `git diff --name-only <base da unidade>` só contém caminhos da unidade |
| Testes existentes intactos | toda task | `git diff --name-only <base da unidade> -- 'web/features/**/*.test.*' 'web/app/**/*.test.*'` vazio |
| Audit | última task de cada unidade | `pnpm audit --audit-level high` |
| Revisão de interface | fim da rodada 4, pelo coordenador | skill `web-design-guidelines` sobre `web/components/`; achado corrigido ou casado com conflito registrado |

Nenhum gate pode ser enfraquecido, pulado ou mascarado. Teste não é apagado nem afrouxado para passar.

---

## Execution Plan

```
T1 → T2
T2 → T3
T3 → T4
T4 → T5
T5 → T6
T6 → T7
T7 → T8
T8 → T9
T9 → T10
```

T1 e T2 são a rodada 3 (DS-L). T3 a T10 são a rodada 4 (DS-A).

---

## Task Breakdown

### T1: Estado na URL

**Unidade**: DS-L
**What**: hook que lê e escreve um valor na query string, com codificadores e histórico configurável.
**Where**: `web/lib/url-state/`
**Arquivos**: `web/lib/url-state/use-query-state.ts`, `codecs.ts`, `index.ts`, `use-query-state.test.tsx`, `codecs.test.ts`
**Depends on**: None
**Reuses**: `useSearchParams` do Next; mock reativo de DS-T; nenhum código de `lib/session`.
**Requirement**: DSP-05 (AC2 a AC4), WDS-D-011

**Mudanças previstas**: `useQueryState(key, codec, { history })` e os codificadores `string`, `oneOf` e `isoDate`. Escrita por `history.replaceState` ou `pushState`; valor padrão remove o parâmetro; escrita adiada em 300 ms quando pedida (busca).

**Done when**:
- [x] Alterar o valor atualiza a query string; voltar ao padrão remove o parâmetro e preserva os demais
- [x] Abrir com o parâmetro presente restaura o valor; valor inválido cai no padrão, sem erro
- [x] `history: "replace"` não cria entrada de histórico; `"push"` cria uma
- [x] Com espera, várias mudanças seguidas geram uma única escrita
- [x] Gates passam

**Tests**: unit
**Verificação visual**: não se aplica
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): estado de filtros e busca na url com apis nativas`

---

### T2: Consulta de mídia segura para SSR

**Unidade**: DS-L
**What**: hook de consulta de mídia cujo resultado no servidor coincide com a primeira renderização do cliente.
**Where**: `web/lib/responsive/`
**Arquivos**: `web/lib/responsive/use-media-query.ts`, `breakpoints.ts`, `index.ts`, `use-media-query.test.tsx`
**Depends on**: T1
**Reuses**: `mockViewport` de DS-T; `useSyncExternalStore` do React.
**Requirement**: DSP-05 (AC5 a AC8), WDS-D-026 (DD-2), DSF-04 (AC8)

**Mudanças previstas**: `useMediaQuery(query)` com `useSyncExternalStore` e retrato de servidor fixo; `useIsNarrow()` (abaixo de 768) e `useIsCoarsePointer()`; constantes 768 e 1024. Sem `matchMedia` disponível, o resultado é o de desktop.

**Done when**:
- [x] Renderizado no servidor (`renderToString`), o hook devolve o valor de desktop; hidratado com `mockViewport` estreito, atualiza sem aviso de hidratação
- [x] Sem `matchMedia`, devolve desktop
- [x] Mudar a largura com `mockViewport` atualiza o componente
- [x] Gates passam, incluindo Audit (fim da unidade DS-L)

**Tests**: unit
**Verificação visual**: não se aplica
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): consulta de midia compativel com ssr e hidratacao`

---

### T3: Marca e estrutura de página

**Unidade**: DS-A
**What**: componente do escudo e os dois componentes que dão a mesma estrutura a toda página.
**Where**: `web/components/app/page-header.tsx`
**Arquivos**: `web/components/app/brand.tsx`, `page-container.tsx`, `page-header.tsx`, `page.test.tsx`
**Depends on**: T2
**Reuses**: componentes de título e `Breadcrumb` do kit; tokens de largura e margem.
**Requirement**: DSF-05 (AC2 a AC5), DSP-01 (AC1 a AC6), DSF-04 (AC7)

**Mudanças previstas**: `BrandMark` (tamanhos 32, 40, 64 e 96; com ou sem nome; `alt` conforme o caso; largura e altura declaradas); `PageContainer` com `main#conteudo` e as três larguras; `PageHeader` com um único `h1`, filete de marca, descrição, trilha, ações e `meta`.

**Done when**:
- [x] `BrandMark` com nome tem `alt=""`; sem nome, `alt="Torcida Jovem"`; sempre com `width` e `height`
- [x] `PageHeader` renderiza exatamente um `h1` e marca a trilha como navegação
- [x] `PageContainer` aceita só as três larguras nomeadas e expõe o alvo do link de pular
- [x] Gates passam

**Tests**: component
**Verificação visual**: conferida no espécime (rodada 5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): marca e estrutura de pagina do design system`

---

### T4: Estados de tela, sem acesso e confirmação

**Unidade**: DS-A
**What**: completar os estados de tela e a confirmação de ação com o que as specs pedem.
**Where**: `web/components/app/states.tsx`
**Arquivos**: `web/components/app/states.tsx`, `require-permission.tsx`, `api-error.tsx`, `confirm-dialog.tsx` e os testes correspondentes
**Depends on**: T3
**Reuses**: lógica atual dos quatro arquivos; `Alert`, `Skeleton`, `AlertDialog` do kit; catálogos de erro de `lib/api/problem.ts`.
**Requirement**: DSP-02 (AC1 a AC6), DSP-03 (AC2 a AC5), DSP-06 (AC4), FND-05 (AC6, preservado)

**Mudanças previstas**: `LoadingState` com `shape`; `EmptyState` com `kind` (`empty`, `no-results`), ícone, orientação e ação; `ErrorState` com alerta de erro e "Tentar de novo"; `NoAccess` no layout padrão; `ConfirmDialog` com `subject`, `consequence` e `error`, e `confirmLabel` obrigatório. Propriedades atuais continuam aceitas.

**Done when**:
- [x] Vazio e sem resultado são distinguíveis por título e ação ("Limpar filtros" no segundo)
- [x] `ConfirmDialog` mostra o registro e a consequência, mantém aberto e mostra o erro quando a ação falha, e mantém confirmar desabilitado com motivo vazio ou só espaços
- [x] As seis chamadas existentes de `ConfirmDialog` compilam sem alteração
- [x] A lógica de `RequirePermission` não muda; seus testes passam sem alteração de asserção
- [x] Gates passam

**Tests**: component
**Verificação visual**: conferida no espécime (rodada 5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): estados de tela e confirmacao padronizados`

---

### T5: Layout de formulário e campos numéricos

**Unidade**: DS-A
**What**: criar o layout de formulário e levar os campos de dinheiro e quantidade à aparência do kit, sem mudar comportamento.
**Where**: `web/components/app/form-layout.tsx`
**Arquivos**: `web/components/app/form-layout.tsx`, `money-field.tsx`, `quantity-field.tsx`, `numeric-text-field.tsx`, `form-layout.test.tsx`, `numeric-text-field.test.tsx`
**Depends on**: T4
**Reuses**: `Field` do kit; `parseBRL` e `formatBRL`; testes atuais de `money-field` e `quantity-field`.
**Requirement**: DSP-04 (AC1, AC4, AC5, AC9), DSC-06 (AC4), FND-05 (AC4, AC5, preservados)

**Mudanças previstas**: `FormLayout` com alerta de erro no topo que recebe foco, e rodapé de ações responsivo; os três campos numéricos passam a compor `Field`. `numeric-text-field.tsx` ganha o teste próprio que faltava.

**Done when**:
- [x] Os testes atuais de `money-field.test.tsx` e `quantity-field.test.tsx` passam sem alteração de asserção
- [x] `FormLayout` mostra o erro de formulário como alerta e leva o foco a ele
- [x] `numeric-text-field` tem teste de rótulo, erro e modo de entrada
- [x] Gates passam

**Tests**: component
**Verificação visual**: conferida no espécime (rodada 5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): layout de formulario e campos numericos no design system`

---

### T6: Status, valor monetário e selo de papel

**Unidade**: DS-A
**What**: desacoplar o indicador de status dos módulos e criar a exibição padronizada de dinheiro e de papel.
**Where**: `web/components/app/status-badge.tsx`
**Arquivos**: `web/components/app/status-badge.tsx`, `money-value.tsx`, `role-badge.tsx`, `format.ts`, e os testes correspondentes
**Depends on**: T5
**Reuses**: `Badge` do kit; `formatBRL`; `shortId` de `format.ts`.
**Requirement**: DSC-05 (AC2 a AC4), DSC-06 (AC1 a AC3, AC5), DSP-07 (AC7), WDS-D-024, WDS-D-026 (DD-3)

**Mudanças previstas**: `StatusBadge` recebe `{ label, tone, icon }` e não importa tipos de módulo; a forma antiga (`status` como texto) continua aceita, com o mapa atual, até DS-INT. `MoneyValue` com as variantes `neutral`, `income`, `expense` e `balance` e o tamanho `display`. `RoleBadge` recebe o rótulo pronto. `format.ts` ganha a apresentação de identificador (tipo e código curto).

**Done when**:
- [x] `components/app/status-badge.tsx` não importa de `lib/api/financeiro` nem de `features/`
- [x] Dois status distintos têm rótulo e ícone distintos; nenhum depende só de cor
- [x] `MoneyValue`: receita com "+", despesa com "−", saldo negativo com sinal, ícone e texto para leitor de tela; algarismos tabulares; valores `-1050`, `0` e `19850` cobertos
- [x] A forma antiga de `StatusBadge` continua funcionando; os testes das features passam sem alteração
- [x] Gates passam

**Tests**: component + unit
**Verificação visual**: conferida no espécime (rodada 5), incluindo a comparação de DD-4
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): status, valor monetario e selo de papel do design system`

---

### T7: Listagem responsiva e ações de linha

**Unidade**: DS-A
**What**: componente único de listagem que mostra tabela ou lista empilhada, com ações de linha nomeadas.
**Where**: `web/components/app/data-list.tsx`
**Arquivos**: `web/components/app/data-list.tsx`, `row-actions.tsx`, `data-list.test.tsx`
**Depends on**: T6
**Reuses**: `Table`, `DropdownMenu` e `IconButton` do kit; `useIsNarrow` de DS-L.
**Requirement**: DSP-05 (AC5 a AC11), DSC-05 (AC1), WDS-D-006, WDS-D-026 (DD-2)

**Mudanças previstas**: `DataList` com colunas tipadas (`priority`, `numeric`, `wrap`, `inList`), `rowLabel`, `primaryAction`, `moreActions` e `layout` (`auto`, `table`, `list`). Uma única árvore renderizada: tabela a partir de 768 px (prioridade 3 oculta por CSS até 1024) ou lista `<ul>` abaixo disso. `RowActions` com menu "Ações de {rowLabel}".

**Done when**:
- [x] Sem `matchMedia` (padrão dos testes), renderiza tabela com cabeçalhos de coluna e nome acessível
- [x] Com `mockViewport` estreito, renderiza lista com título, até três campos, status e ações, e nenhuma tabela
- [x] Coluna de prioridade 3 tem a classe que a oculta abaixo de 1024 px; coluna de prioridade 1 nunca é ocultada
- [x] `layout="table"` mantém a tabela em qualquer largura
- [x] Toda ação de linha tem nome acessível que contém o texto visível e identifica a linha; nomes são únicos entre linhas
- [x] Renderização no servidor igual à primeira renderização do cliente (sem aviso de hidratação)
- [x] Gates passam

**Tests**: component
**Verificação visual**: os três comportamentos conferidos no espécime em 390, 768, 820, 1024 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): listagem responsiva com tabela e lista empilhada`

---

### T8: Barra de filtros

**Unidade**: DS-A
**What**: barra de filtros com contagem de resultados, limpeza e forma recolhida em tela estreita.
**Where**: `web/components/app/filter-bar.tsx`
**Arquivos**: `web/components/app/filter-bar.tsx`, `filter-bar.test.tsx`
**Depends on**: T7
**Reuses**: `Sheet` (lado inferior), `Button`, `Field` do kit; `useIsNarrow`.
**Requirement**: DSP-05 (AC1), DSP-02 (AC3)

**Mudanças previstas**: `FilterBar` com `activeCount`, `resultCount`, `resultNoun` e `onClear`; em tela estreita, a busca fica visível e os demais filtros vão para uma gaveta inferior aberta por "Filtros (n)". O estado dos filtros vem de quem usa (o componente não conhece a URL).

**Done when**:
- [ ] "Limpar filtros" só aparece com filtro ativo e chama `onClear`
- [ ] A contagem de resultados fica em região viva e respeita singular e plural
- [ ] Com `mockViewport` estreito, os filtros ficam na gaveta e o botão mostra a quantidade de filtros ativos
- [ ] Gates passam

**Tests**: component
**Verificação visual**: conferida no espécime (rodada 5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): barra de filtros com contagem e forma recolhida`

---

### T9: Shell com identidade e layout de acesso

**Unidade**: DS-A
**What**: separar a apresentação do shell da sessão, aplicar a identidade e o novo ponto de quebra, e criar o layout das telas de acesso.
**Where**: `web/components/app/shell-frame.tsx`
**Arquivos**: `web/components/app/shell-frame.tsx`, `app-shell.tsx`, `nav.ts`, `access-layout.tsx`, `app-shell.test.tsx`, `shell-frame.test.tsx`, `access-layout.test.tsx`
**Depends on**: T8
**Reuses**: `NavList`, `NavLink` e `visibleItems` atuais; `Sheet`, `DropdownMenu`, `IconButton`; `BrandMark`.
**Requirement**: DSS-01 (AC1 a AC9), DSS-02 (AC1 a AC6), DSS-03 (AC1, AC2, AC5, AC6), FND-05 (AC1, AC8 preservados; AC2 revisado por WDS-D-019)

**Mudanças previstas**: `ShellFrame` sem sessão (entradas, caminho atual, nome, menu); `AppShell` vira adaptador com a mesma interface de hoje; link "Pular para o conteúdo"; marca "Torcida Jovem" com escudo; tokens do shell escuro; item ativo com `aria-current`, peso e barra; navegação fixa a partir de 1024 px e gaveta abaixo; "Trocar senha" como link; `NavItem` ganha `icon` e `description` opcionais; `AccessLayout` em duas colunas no desktop e coluna única abaixo de 1024, sem "Desde 2013".

**Done when**:
- [ ] O primeiro elemento focável é o link de pular, que leva o foco a `main#conteudo`
- [ ] A marca exibida é "Torcida Jovem" e o texto "TJ Platform" não aparece no shell (teste atualizado, citando WDS-D-020)
- [ ] A navegação fixa usa o ponto de quebra de 1024 px (classes `lg`), e o botão de menu existe abaixo dele (teste atualizado, citando WDS-D-019)
- [ ] Item ativo tem `aria-current="page"`; "Trocar senha" é link e "Sair" é botão
- [ ] Com `must_change_password`, só "Sair" aparece (FND-05 AC8)
- [ ] `AppShell` mantém a interface `items` e `children`; `app/(app)/layout.tsx` compila sem alteração
- [ ] `AccessLayout` mostra escudo e nome e nenhum texto "Desde 2013"
- [ ] Gates passam

**Tests**: component
**Verificação visual**: shell e layout de acesso conferidos no espécime; a revalidação completa de FND-05 AC2 no navegador é de DS-S (`tasks/04-shell-e-acesso.md`, T5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): shell com identidade da torcida e layout de acesso`

---

### T10: Guarda de tokens de `components/app`

**Unidade**: DS-A
**What**: remover as sobras de estilo antigo de `components/app` e travar o diretório com a guarda.
**Where**: `web/components/app/`
**Arquivos**: `web/components/app/tokens.guard.test.ts` e arquivos do diretório com sobras
**Depends on**: T9
**Reuses**: biblioteca de guarda de DS-F.
**Requirement**: DSF-01 (AC4, AC7), DSC-07 (AC1 a AC3)

**Mudanças previstas**: teste de guarda sobre `components/app`, com exceção declarada só para os arquivos do shell e de marca que usam a fonte de identidade.

**Done when**:
- [ ] A guarda passa em todo `components/app`, com exceções explícitas e justificadas
- [ ] Todo arquivo de componente do diretório tem teste que o referencia
- [ ] A lista de testes alterados na unidade, com a decisão que motiva cada um, consta no relatório da unidade
- [ ] Gates passam, incluindo Audit (fim da unidade DS-A)

**Tests**: unit + component
**Verificação visual**: uma tela autenticada e `/entrar` abertas com `pnpm dev`, com o shell novo e sem quebra de layout
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `refactor(web): componentes de aplicacao sem valores soltos, com guarda de tokens`

---

## Phase Execution Map

```
T1 → T2
T2 → T3
T3 → T4
T4 → T5
T5 → T6
T6 → T7
T7 → T8
T8 → T9
T9 → T10
```

Rodada 3: T1 e T2. Rodada 4: T3 a T10. Ao fim da rodada 4, o coordenador atualiza o grafo do Graphify e roda a revisão de interface.

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | 1 hook e seus codificadores | ✅ Granular |
| T2 | 1 hook | ✅ Granular |
| T3 | marca e 2 componentes de estrutura | ✅ Aceitável |
| T4 | estados e confirmação (mesmo conjunto de padrões de DSP-02 e DSP-03) | ✅ Aceitável |
| T5 | 1 layout e a família de campos numéricos | ✅ Aceitável |
| T6 | 3 componentes de domínio pequenos | ✅ Aceitável |
| T7 | listagem e ações de linha (a segunda só existe dentro da primeira) | ✅ Aceitável |
| T8 | 1 componente | ✅ Granular |
| T9 | shell em duas partes e layout de acesso | ✅ Aceitável (a maior task; dividir separaria o shell do teste que o cobre) |
| T10 | guarda de 1 diretório | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | T2 | T2 | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | T4 | T4 | ✅ Match |
| T6 | T5 | T5 | ✅ Match |
| T7 | T6 | T6 | ✅ Match |
| T8 | T7 | T7 | ✅ Match |
| T9 | T8 | T8 | ✅ Match |
| T10 | T9 | T9 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | `lib/url-state` | unit | unit | ✅ OK |
| T2 | `lib/responsive` | unit | unit | ✅ OK |
| T3 | `components/app` | component | component | ✅ OK |
| T4 | `components/app` | component | component | ✅ OK |
| T5 | `components/app` | component | component | ✅ OK |
| T6 | `components/app` | component | component + unit | ✅ OK |
| T7 | `components/app` | component | component | ✅ OK |
| T8 | `components/app` | component | component | ✅ OK |
| T9 | `components/app` | component | component | ✅ OK |
| T10 | `components/app` e guarda | unit + component | unit + component | ✅ OK |
