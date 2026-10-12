# Componentes — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos critérios de aceite; um commit atômico por task; a task é marcada como concluída neste arquivo antes do commit e no mesmo commit. **Se a skill não puder ser ativada, PARE e avise.**

Antes de escrever código Next: ler `web/AGENTS.md` e a documentação do pacote instalado em `web/node_modules/next/dist/docs/`. Para componentes do Base UI, conferir a API no pacote instalado (`web/node_modules/@base-ui/react/`), não de memória.

Autorização: a aprovação destas tasks autoriza implementação e commits locais na branch da unidade. `git push`, PR, merge e deploy exigem autorização explícita para cada ação.

---

**Spec**: `.specs/features/web-design/spec/02-componentes.md`
**Design**: `.specs/features/web-design/design/02-componentes.md`
**Status**: Draft (2026-10-11), aguardando aprovação do mantenedor.

**Pré-requisito**: DS-F integrada (`tasks/01-fundacao-visual.md`, T1 a T4).

**Unidades de execução** (propriedade e DAG em `design/05-migracao-web-v1.md`):

| Unidade | Tasks | Rodada | Entrada | Saída |
| --- | --- | --- | --- | --- |
| DS-T Infraestrutura compartilhada de testes | T1, T2 | 2, paralela a DS-K | DS-F integrada | Mock de navegação reativo e apoio a consulta de mídia; todos os testes existentes passando sem alteração |
| DS-K Kit de componentes | T3 a T8 | 2, paralela a DS-T | DS-F integrada | `components/ui` nos tokens, com variantes, estados e alias de compatibilidade; guarda do kit passando |

DS-T e DS-K rodam em agentes e worktrees separados. São paralelas porque os arquivos são disjuntos e nenhum teste de `components/ui` importa a infraestrutura de DS-T (conferido com os `import` reais em 2026-10-11).

## Restrições comuns a todas as tasks deste arquivo

- Só os arquivos da unidade são alterados (gate Propriedade). `package.json` e `pnpm-lock.yaml` não mudam; se um componente exigir dependência nova, a task para e reporta.
- **Compatibilidade**: nenhuma mudança pode quebrar as telas ainda não migradas. Nomes de variante, propriedades e assinaturas em uso continuam aceitos como alias até DS-INT. Os testes existentes de `features/` e `app/` passam sem alteração.
- Radio, Switch, Tabs e Tooltip não são criados (WDS-D-022).
- Nenhum arquivo de `api/`, nenhum contrato, nenhuma regra de negócio ou de permissão.
- Nenhuma cor literal, nenhum valor arbitrário, nenhuma classe `dark:`, nenhum `transition-all` nos arquivos alterados.

## Test Coverage Matrix

> Gerada do código, das diretrizes do projeto e da spec. Diretrizes consultadas: `CLAUDE.md`, `web/AGENTS.md`, `docs/development/web.md`, `web/vitest.config.ts`, `.specs/features/web/tasks/*.md`.

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| Infraestrutura de testes (`test/`, mocks de `lib/session`) | unit | Comportamento novo coberto; todo arquivo que a consome continua passando sem alteração | `web/test/*.test.ts(x)`, `web/lib/session/*.test.ts(x)` | `pnpm test` |
| `components/ui/` | component | Papel acessível, nome acessível, teclado e cada estado definido na spec; alias de compatibilidade cobertos | `web/components/ui/*.test.tsx` | `pnpm test` |
| Guarda do kit | unit | Nenhum valor solto em `components/ui` | `web/components/ui/*.guard.test.ts` | `pnpm test` |

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

Nenhum gate pode ser enfraquecido, pulado ou mascarado. Teste não é apagado nem afrouxado para passar.

---

## Execution Plan

```
T1 → T2
T3 → T4
T4 → T5
T5 → T6
T6 → T7
T7 → T8
```

Duas cadeias independentes: T1 e T2 (DS-T), T3 a T8 (DS-K). Cada cadeia é de um agente.

---

## Task Breakdown

### T1: Mock de navegação reativo ao histórico

**Unidade**: DS-T
**What**: fazer o substituto de `next/navigation` usado nos testes reagir a `history.pushState` e `history.replaceState`, sem mudar a API atual.
**Where**: `web/lib/session/navigation-mock.ts`
**Arquivos**: `web/lib/session/navigation-mock.ts`, `web/lib/session/navigation-mock.test.tsx`
**Depends on**: None
**Reuses**: `router`, `setLocation`, `navigations` e `resetNavigation` atuais, com os mesmos nomes e assinaturas.
**Requirement**: DSP-05 (AC2, AC3), DSC-07 (AC2)

**Mudanças previstas**: `useSearchParams` e `usePathname` do mock passam a assinar mudanças de histórico (`useSyncExternalStore`), de modo que um componente que escreve na URL é renderizado de novo com o valor novo. `setLocation` também notifica. Nada mais muda.

**Done when**:
- [x] Um componente de teste que chama `replaceState` vê o novo `useSearchParams` na renderização seguinte; o mesmo para `pushState` e para `setLocation`
- [x] `resetNavigation` limpa assinaturas e chamadas registradas
- [x] Os 33 arquivos que importam o mock passam **sem nenhuma alteração**
- [x] Gates passam

**Tests**: unit
**Verificação visual**: não se aplica
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `test(web): mock de navegacao reativo ao historico`

---

### T2: Apoio de teste para consulta de mídia

**Unidade**: DS-T
**What**: oferecer aos testes uma forma única de simular largura de tela e tipo de ponteiro, sem alterar o ambiente padrão.
**Where**: `web/test/`
**Arquivos**: `web/test/media.ts`, `web/test/media.test.ts`; `web/test/setup.ts` só se for preciso restaurar o estado entre testes
**Depends on**: T1
**Reuses**: `web/test/setup.ts`, `web/test/render.tsx`.
**Requirement**: DSP-05 (AC5 a AC8), DSC-01 (AC4)

**Mudanças previstas**: `mockViewport({ width, coarse })` instala um `matchMedia` controlado e devolve a função de restauração; por padrão o jsdom continua sem `matchMedia`, o que os hooks tratam como desktop. `test/msw/fixtures.ts`, `test/msw/server.ts`, `test/render.tsx` e `lib/session/test-utils.tsx` **não mudam** nesta task: pertencem a DS-T e ficam congelados.

**Done when**:
- [x] `mockViewport` responde a consultas de largura mínima e de `pointer: coarse`, e notifica ouvintes quando a largura muda
- [x] O estado é restaurado entre testes, sem vazar para outros arquivos
- [x] Nenhum teste existente muda; a contagem de testes só cresce
- [x] Gates passam, incluindo Audit (fim da unidade DS-T)

**Tests**: unit
**Verificação visual**: não se aplica
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `test(web): apoio de teste para largura de tela e tipo de ponteiro`

---

### T3: Button e IconButton

**Unidade**: DS-K
**What**: ajustar o botão às variantes, tamanhos e estados do design e criar o botão de ícone com nome acessível obrigatório.
**Where**: `web/components/ui/button.tsx`
**Arquivos**: `web/components/ui/button.tsx`, `web/components/ui/icon-button.tsx`, `web/components/ui/actions.test.tsx`
**Depends on**: None
**Reuses**: `buttonVariants` com `class-variance-authority`; `Button` do Base UI.
**Requirement**: DSC-01 (AC1 a AC7), DSC-02 (AC1 a AC5)

**Mudanças previstas**: variantes `primary`, `secondary`, `ghost`, `destructive` (sólida) e `link`; tamanhos `sm`, `md` e `lg` pelos tokens de altura; propriedade `loading` com `loadingLabel`, `aria-busy` e ícone; foco por anel de 2 px; `touch-action: manipulation`; transição só de cor. Alias mantidos: `default` (igual a `primary`), `outline` (igual a `secondary`), tamanhos `icon` e `icon-sm`. Saem `xs`, `icon-xs` e a variante `secondary` antiga de preenchimento cinza (sem uso fora do kit).

**Done when**:
- [x] Cada variante e tamanho tem teste de papel, nome e estado desabilitado
- [x] `loading` desabilita, expõe `aria-busy` e mostra o rótulo de progresso
- [x] `IconButton` sem `label` não compila (teste de tipo) e expõe o `label` como nome acessível
- [x] Os alias `default`, `outline`, `icon` e `icon-sm` continuam renderizando; teste cobre cada um
- [x] Nenhum arquivo fora de `components/ui` precisou mudar para compilar
- [x] Gates passam

**Tests**: component
**Verificação visual**: conferida no espécime (rodada 5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): botoes do design system com alias de compatibilidade`

---

### T4: Campos

**Unidade**: DS-K
**What**: unificar medidas e estados de Input, Textarea, Checkbox e Label, criar o select nativo do kit e completar o Field.
**Where**: `web/components/ui/field.tsx`
**Arquivos**: `web/components/ui/input.tsx`, `textarea.tsx`, `checkbox.tsx`, `label.tsx`, `field.tsx`, `native-select.tsx`, `field.test.tsx` (ampliado)
**Depends on**: T3
**Reuses**: `Field` do Base UI; testes atuais de `field.test.tsx`.
**Requirement**: DSC-03 (AC1 a AC7), DSF-03 (AC7)

**Mudanças previstas**: altura e texto pelos tokens de controle; estados inválido, desabilitado e somente leitura; `NativeSelect` com aparência própria, seta e cores de fundo e de texto explícitas; `FieldLabel` com `optional`; `FieldError` com ícone; Checkbox com alvo único com o rótulo.

**Done when**:
- [x] `getByLabelText` encontra Input, Textarea e NativeSelect dentro de `Field` (associação de rótulo confirmada; corrigida se falhar)
- [x] Campo com erro tem `aria-invalid`, `aria-describedby` e mensagem com ícone e texto
- [x] Input, Textarea e NativeSelect usam o mesmo token de altura e de borda
- [x] Rótulo com `optional` mostra "(opcional)"; clicar no rótulo do Checkbox alterna o controle
- [x] Os testes atuais de `field.test.tsx` continuam passando
- [x] Gates passam

**Tests**: component
**Verificação visual**: conferida no espécime (rodada 5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): campos do design system e select nativo do kit`

---

### T5: Indicadores e feedback

**Unidade**: DS-K
**What**: ajustar Badge, Alert, Skeleton e Toast aos tons e regras do design.
**Where**: `web/components/ui/toast.tsx`
**Arquivos**: `web/components/ui/badge.tsx`, `alert.tsx`, `skeleton.tsx`, `toast.tsx`, `display.test.tsx` (ampliado)
**Depends on**: T4
**Reuses**: `Toast` do Base UI; API imperativa `toast.show`, `toast.success`, `toast.error`.
**Requirement**: DSC-01 (AC7), DSC-04 (AC6, AC7), DSC-05 (AC2)

**Mudanças previstas**: Badge e Alert com os tons `neutral`, `success`, `warning` e `error`, ícone e sem `transition-all`; alias `destructive` (tom de erro) e `default`, `secondary`, `outline` do Badge mantidos; Toast com ícone por tom, `toast.warning`, durações de 5 e 8 segundos, botão de fechar com alvo adequado e posição pela área segura; Skeleton com as formas de linha, linha de tabela e bloco.

**Done when**:
- [x] Cada tom de Badge e de Alert tem teste de texto e de ícone; nenhum depende só de cor
- [x] `toast.success`, `toast.error` e `toast.show` mantêm a assinatura; `toast.warning` existe
- [x] O toast é anunciado por região viva e pode ser dispensado pelo teclado
- [x] `lib/session` e as features que usam `toast` e `Alert` compilam e passam sem alteração
- [x] Gates passam

**Tests**: component
**Verificação visual**: conferida no espécime (rodada 5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): indicadores e feedback do design system`

---

### T6: Sobreposições

**Unidade**: DS-K
**What**: ajustar Dialog, AlertDialog, Sheet e DropdownMenu ao comportamento de foco, rolagem e tela estreita do design.
**Where**: `web/components/ui/dialog.tsx`
**Arquivos**: `web/components/ui/dialog.tsx`, `alert-dialog.tsx`, `sheet.tsx`, `dropdown-menu.tsx`, `overlays.test.tsx` (ampliado)
**Depends on**: T5
**Reuses**: primitivos do Base UI; testes atuais de `overlays.test.tsx`.
**Requirement**: DSC-04 (AC1 a AC5, AC9), DSC-01 (AC2)

**Mudanças previstas**: Dialog em tela cheia abaixo de 640 px, com rodapé fixo ao contêiner e área segura; `overscroll-behavior: contain`; sombra de nível 2 sobre o véu; Sheet com os lados `left` e `bottom`; item de menu com altura pelos tokens, variante destrutiva e possibilidade de ser link; botão de fechar como `IconButton`.

**Done when**:
- [x] Abrir leva o foco para dentro; fechar devolve o foco ao gatilho; Esc fecha (testes para Dialog, AlertDialog, Sheet e menu)
- [x] Clique fora fecha o Dialog e não fecha o AlertDialog
- [x] Item de menu que navega é renderizado como link
- [x] Os testes atuais de `overlays.test.tsx` continuam passando
- [x] Gates passam

**Tests**: component
**Verificação visual**: diálogo em 390 px e gaveta inferior conferidos no espécime (rodada 5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): sobreposicoes do design system`

---

### T7: Tabela, cartão, trilha, "Mostrar mais" e tipografia

**Unidade**: DS-K
**What**: ajustar a tabela e criar os componentes de dados e de título que faltam.
**Where**: `web/components/ui/table.tsx`
**Arquivos**: `web/components/ui/table.tsx`, `card.tsx`, `breadcrumb.tsx`, `show-more.tsx`, `typography.tsx`, `data.test.tsx`
**Depends on**: T6
**Reuses**: estrutura atual da tabela; `Button` e ícones do Lucide.
**Requirement**: DSC-05 (AC1, AC5 a AC7), DSF-03 (AC3, AC4, AC8), DSP-01 (AC5)

**Mudanças previstas**: tabela com cabeçalho em `muted`, régua de 2 px, `scope="col"`, propriedade `numeric`, quebra por coluna e variante `scrollable`; modo legado de rolagem embutida mantido como padrão até DS-INT. `Card` sem sombra; `Breadcrumb` com `aria-current` e forma reduzida; `ShowMore` com contagem em região viva; `DisplayTitle`, `PageTitle`, `SectionTitle` e `Eyebrow` com `text-wrap: balance`.

**Done when**:
- [x] Coluna `numeric` alinha à direita e usa algarismos tabulares
- [x] A tabela usada como hoje (sem propriedades novas) mantém a rolagem embutida; teste cobre o modo legado
- [x] `ShowMore` mostra "Mostrando X de Y" e anuncia a mudança; sem total, mostra só a quantidade visível
- [x] `Breadcrumb` marca o último item como página atual, sem link
- [x] Os componentes de título são o único lugar de `components/ui` com `font-display` e `uppercase`
- [x] Gates passam

**Tests**: component
**Verificação visual**: conferida no espécime (rodada 5)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `feat(web): tabela, cartao, trilha e tipografia do design system`

---

### T8: Limpeza do kit e guarda de tokens

**Unidade**: DS-K
**What**: levar os componentes restantes aos tokens e travar o diretório com a guarda de design.
**Where**: `web/components/ui/`
**Arquivos**: `web/components/ui/select.tsx`, `separator.tsx`, demais arquivos do kit com sobras, `web/components/ui/tokens.guard.test.ts`
**Depends on**: T7
**Reuses**: biblioteca de guarda de DS-F (`web/test/design/guard.ts`).
**Requirement**: DSF-01 (AC4, AC7), DSC-07 (AC1 a AC3), DSM-05 (AC5)

**Mudanças previstas**: remoção das classes `dark:` e dos valores arbitrários restantes em todo o kit; `Select` rico e `Separator` nos tokens; teste de guarda que aplica todas as regras a `components/ui`, com exceção declarada só para `typography.tsx` (fonte de identidade e caixa alta).

**Done when**:
- [x] A guarda passa em todo `components/ui`, com a lista de exceções explícita e justificada
- [x] Todo arquivo de componente do kit é referenciado por pelo menos um teste
- [x] A lista de testes alterados nesta unidade, com a decisão que motiva cada um, consta no relatório da unidade
- [x] Todos os testes existentes de `features/` e `app/` passam sem alteração
- [x] Gates passam, incluindo Audit (fim da unidade DS-K)

**Tests**: unit + component
**Verificação visual**: uma tela autenticada e `/entrar` abertas com `pnpm dev`, sem quebra de layout (estado intermediário esperado)
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Testes existentes intactos
**Commit**: `refactor(web): kit de interface sem valores soltos, com guarda de tokens`

---

## Phase Execution Map

```
T1 → T2
T3 → T4
T4 → T5
T5 → T6
T6 → T7
T7 → T8
```

Rodada 2: as duas cadeias em paralelo. Integração pelo coordenador ao fim, com conferência de propriedade e gates completos com as duas juntas.

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | 1 arquivo de mock | ✅ Granular |
| T2 | 1 utilitário de teste | ✅ Granular |
| T3 | botão e botão de ícone (mesmo contrato de variantes) | ✅ Aceitável |
| T4 | família de campos, que precisa das mesmas medidas | ✅ Aceitável |
| T5 | 4 componentes de indicação, mesma tabela de tons | ✅ Aceitável |
| T6 | 4 sobreposições, mesmas regras de foco e rolagem | ✅ Aceitável |
| T7 | tabela e 4 componentes pequenos novos | ✅ Aceitável |
| T8 | limpeza e guarda de 1 diretório | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | None | None | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | T4 | T4 | ✅ Match |
| T6 | T5 | T5 | ✅ Match |
| T7 | T6 | T6 | ✅ Match |
| T8 | T7 | T7 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | infraestrutura de testes | unit | unit | ✅ OK |
| T2 | infraestrutura de testes | unit | unit | ✅ OK |
| T3 | `components/ui` | component | component | ✅ OK |
| T4 | `components/ui` | component | component | ✅ OK |
| T5 | `components/ui` | component | component | ✅ OK |
| T6 | `components/ui` | component | component | ✅ OK |
| T7 | `components/ui` | component | component | ✅ OK |
| T8 | `components/ui` e guarda | unit + component | unit + component | ✅ OK |
