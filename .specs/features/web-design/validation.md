# Validation — web-design: PASS ✅ (T14, iteração 4, excepcional e restrita — WDS-D-035)

**Veredito atual (iteração 4): PASS ✅**. O relatório da iteração 4 está na seção "Iteração 4 (excepcional, autorizada — WDS-D-035)", no fim deste arquivo. O texto abaixo, até essa seção, é o relatório da iteração 3, mantido sem alteração; o título original dele era: "Validation — web-design: FAIL ❌ (T14, iteração 3 de 3 — última antes de escalar ao mantenedor)".


**Veredito: FAIL ❌**

**Date**: 2026-10-13
**Spec**: `.specs/features/web-design/spec/01` a `05` (DSF, DSC, DSP, DSS, DSM); Web V1 em `.specs/features/web/spec/01` a `06`
**Diff range**: `7c30c6a..13e56c5` (base documental `7c30c6a` sobre a `develop` em `cebbdf2`; HEAD `13e56c5c6a21f8a6e58dabb21a153851339e0ff9`). Correções desta iteração, feitas pelo coordenador: `fb62b99` (A1, A2, A3, O3) e `13e56c5` (O1)
**Verifier**: sub-agente independente (não escreveu nenhuma correção). Evidência por `arquivo:linha`, saída de comando ou medição no navegador.

Esta é a **terceira e última iteração**. Todas as lacunas da iteração 2 (A1, A2, A3, O1, O3) estão corrigidas e conferidas, por teste, mutante e navegador. O FAIL vem de quatro itens pequenos, para o mantenedor decidir em cada um entre **corrigir** ou **registrar como conflito ou isenção em `STATE.md`**:

| # | Item | Natureza | Onde |
| --- | --- | --- | --- |
| D1 | Mutantes M28 e M29 sobrevivem: nenhum teste afirma o anel do kit nos links de produto e de devolução (correção O1 desta iteração) | sensor | `features/estoque/produtos/produtos-page.tsx:140`; `features/financeiro/lancamentos/lancamento-detail.tsx:162` |
| D2 (F1) | `Input` e `MoneyField` do espécime sem `name` e, nos `Input`, sem `autoComplete` | `web-design-guidelines` ("inputs need autocomplete and meaningful name"); só em `app/(dev)`, que responde 404 em produção | `app/(dev)/design-system/acesso/example-login-form.tsx:47,60`; `example-listing.tsx:58`; `fields-section.tsx:52,55,58,101,108,115,121,127`; `overlays-section.tsx:69,72` |
| D3 (F2) | Campo de arquivo do comprovante sem `name` | `web-design-guidelines` (`autocomplete` não se aplica a `type="file"`) | `features/financeiro/comprovantes/comprovantes-panel.tsx:176` |
| D4 (F3) | Link da marca no shell sem estado de `hover` | `web-design-guidelines` ("buttons/links need hover state") | `components/app/shell-frame.tsx:155` |

D2 e D4 já estavam no código e apareceram na minha varredura da iteração 2, mas não os reportei lá. A falha foi minha, de consistência: na iteração 1 reprovei placeholders do espécime, e na iteração 2 deixei os campos do espécime e o link da marca de fora. D3 também já existia. Na iteração 2 eu tinha descartado esse campo só quanto a `autocomplete` e não conferi `name`. Nenhum dos quatro foi introduzido pelas correções. D1 é novo: a correção O1 entrou sem teste.

## Histórico

- **Iteração 1** (`7c30c6a..f9de714`): **FAIL**. Gates verdes (110 arquivos, 1774 testes); 203/219 ✅, 3 ❌ (DSP-04 AC6, DSP-04 AC7, DSM-06 AC2), 13 ⚠️; sensor 16/15/1 (M13 sobreviveu); 22 achados da `web-design-guidelines` sem correção; desvios de `lib/forms` e `.gitignore`. Relatório em `0da56c2`.
- **Iteração 2** (`7c30c6a..7499d87`): **FAIL**. Lacunas da iteração 1 fechadas (`87fdd3a`, `63b5e8e`, `ea87164`; decisões WDS-D-034); gates verdes (1785 testes); 204/219 ✅, 3 ❌ (DSP-04 AC6, DSP-07 AC4, DSM-06 AC2), 12 ⚠️ aceitas; sensor 23/23/0; achados novos A1 a A3 (já existiam e a iteração 1 não os viu) e observações O1 a O3. Relatório em `de84949`.
- **Iteração 3** (este relatório): A1 a A3, O1 e O3 corrigidos; restam D1 a D4.

---

## Task Completion

| Task | Status | Notes |
| --- | --- | --- |
| Rodadas 1 a 6, T11, T12, T13 | ✅ | `STATE.md:267-282`; `evidence/` |
| Correções das iterações 1 e 2 | ✅ | `87fdd3a`, `63b5e8e`, `ea87164`, `d050091`, `7499d87`, `fb62b99`, `13e56c5` |
| T14 | ❌ | D1 a D4 |

---

## Gate Check (repositório principal, `web/`, `API_URL=http://localhost:8080`)

| Gate | Saída |
| --- | --- |
| `pnpm lint` | exit 0 |
| `pnpm exec next typegen` | exit 0 |
| `pnpm exec tsc --noEmit` | exit 0 |
| `pnpm test` | exit 0 — **110 arquivos, 1787 testes, 1787 passaram, 0 falharam, 0 pulados** |
| `pnpm build` | exit 0 ("Compiled successfully"), no worktree temporário em `13e56c5` (para não sobrescrever o `.next` servido em `:3000`) |
| `pnpm lint:api` | exit 0 |
| `pnpm gen:api:check` | exit 0 |
| `pnpm audit --audit-level high` | exit 0 — 1 high ignorada (`GHSA-vfj7-8cjw-p6xm`, `web/pnpm-workspace.yaml:10-12`, inalterado desde a `develop`) |
| `git diff --name-only develop -- api/ web/package.json web/pnpm-lock.yaml` | vazio |

Testes antes do ciclo: 69 arquivos / 851 (`STATE.md:271`). Agora: 110 / 1787. Nenhum arquivo de teste removido.

---

## Conferência das correções da iteração 2

| Lacuna | Correção | Teste / guarda | Mutante | Navegador |
| --- | --- | --- | --- | --- |
| A1: selects do lançamento sem `autoComplete` | `lancamento-form.tsx:189,224,276` | `global.guard.test.ts:221` | M23 morto | `name` `tipo`, `conta_id`, `forma_pagamento`, `autocomplete="off"` |
| A2: número e unidade com espaço comum | `produto-detail.tsx:107`, `ajuste-dialog.tsx:120,127`, `movimentacao-dialog.tsx:57` (U+00A0) | `global.guard.test.ts:240`; `produto-detail.test.tsx:50`; `ajuste-dialog.test.tsx:55-65`; `movimentacoes-panel.test.tsx:423-424` | M24, M26, M27 mortos | "3 UN", saldo atual e resultante e nome de "Devolver" com U+00A0 |
| A3: `QuantityField` do ajuste sem `name` | `ajuste-dialog.tsx:96` | `global.guard.test.ts:221` | M25 morto | `name="quantidade"`, `autocomplete="off"` |
| O3: guarda por lista fixa | guarda geral para `NativeSelect`, `Textarea` e `QuantityField` | `global.guard.test.ts:221` | M23, M25 mortos | — (não cobre `Input` nem `MoneyField`, origem de D2) |
| O1: links sem o anel do kit | `produtos-page.tsx:140`, `lancamento-detail.tsx:162` | **nenhum** | **M28, M29 sobrevivem (D1)** | anel preto de 2 px com afastamento branco de 2 px nos dois links, por teclado |

---

## Spec-Anchored Acceptance Criteria

Legenda: ✅ o teste/evidência afirma o valor que a spec define · ❌ lacuna · ⚠️ lacuna de precisão da spec, aceita e registrada em WDS-D-034 (`STATE.md:154`). As linhas sem mudança desde a iteração 2 foram reconferidas contra `13e56c5`; as referências a arquivos alterados em `fb62b99` e `13e56c5` foram remapeadas.

### DSF-01 Tokens em camadas

| AC | Spec-defined outcome | Evidência + asserção | Result |
| --- | --- | --- | --- |
| 1 | três camadas | `web/test/design/tokens.test.ts:100` — primitivos, semânticos e de componente em blocos separados | ✅ |
| 2 | `#111111 #D81E28 #FFFFFF #A7A9AC` | `tokens.test.ts:70` — os quatro primitivos com exatamente esses valores | ✅ |
| 3 | fundo…informação + shell | `tokens.test.ts:135,163` — `@theme` expõe todos os semânticos; cada token com o valor do design | ✅ |
| 4 | só semânticos/componente | `tokens.test.ts:117` (nenhum primitivo no `@theme`, logo nenhum utilitário `--tj-*`); guarda `arbitrary-color` (`test/design/guard.ts:72-75`) | ✅ |
| 5 | um único bloco de tema | `tokens.test.ts:124` | ✅ |
| 6 | sem seletor nem `prefers-color-scheme` | `tokens.test.ts:231,237,241` | ✅ |
| 7 | guarda falha com arquivo e linha | `test/design/guard.test.ts:61,75`; sensor M02 morto | ✅ |

### DSF-02 Cor e contraste

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1-3 | texto ≥ 4,5:1; gráfico/borda/foco ≥ 3:1; teste calcula cada par | `test/design/contrast.test.ts:30,34,42,55`; sensor M01 morto (5 pares de `muted-foreground` abaixo de 4,5) | ✅ |
| 4 | ação primária em preto | `tokens.test.ts:216`; `components/ui/actions.test.tsx:52` | ✅ |
| 5-6 | vermelho só acento, nunca fundo de página/cartão/cabeçalho | Estático: `grep (bg|text|border…)-brand` em `app/components/features` só acha `components/ui/typography.tsx:31` (`after:bg-brand`, filete do `PageTitle`) e o espécime; item ativo usa `before:bg-sidebar-primary` (`components/app/shell-frame.test.tsx:268-270`) | ✅ |
| 7 | erro ≠ vermelho de marca | `tokens.test.ts:216` | ✅ |
| 8 | aço não é texto nem borda de campo | `tokens.test.ts:222` | ✅ |
| 9 | vermelho não é texto < 24 px no shell | Estático: nenhum `text-brand`/`text-sidebar-primary` no código (grep acima) | ✅ |
| 10 | verde/âmbar só estado | Estático: `text-/bg-/border-(success|warning)` só em `alert.tsx`, `badge.tsx`, `toast.tsx`, `money-value.tsx` | ✅ |

### DSF-03 Tipografia

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | Barlow 600/700, Geist, Geist Mono por `next/font` | `test/design/fonts.test.ts:26,35,43,63` | ✅ |
| 2 | corpo em Geist | `tokens.test.ts:248` (`--font-sans` → Geist) | ✅ |
| 3 | Barlow só nos papéis listados | `test/design/global.guard.test.ts:30-36` (exceção `font-display` só em `typography.tsx`); `components/ui/tokens.guard.test.ts:41` | ✅ |
| 4 | caixa alta só em título de página, rótulo de grupo e sobretítulo | `components/ui/data.test.tsx:390-394` põe `uppercase` também no `DisplayTitle` do nome da marca; design aprovado manda caixa alta (`design/01-fundacao-visual.md:192,335`) | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 5 | só a escala, mínimo 12 px | regra `arbitrary-text-size` (`guard.ts:77`) | ✅ |
| 6 | algarismos tabulares | `components/app/data-list.test.tsx:122`; `components/ui/data.test.tsx:55`; `components/app/money-value.test.tsx:29` | ✅ |
| 7 | toque: campo ≥ 16 px | `tokens.test.ts:294` | ✅ |
| 8 | `text-wrap: balance` em títulos | `components/ui/data.test.tsx:368,382` (`text-balance`) | ✅ |

### DSF-04 Espaço, forma, elevação e movimento

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | escala de espaço | regra `arbitrary-spacing` (`guard.ts:79`) | ✅ |
| 2 | três raios; pílula só em status | `tokens.test.ts:337`; estático: `rounded-full` só em `components/ui/badge.tsx:12` | ✅ |
| 3 | sem sombra em repouso | `tokens.test.ts:344`; `data.test.tsx:183` (Card sem sombra); estático: `shadow-elevation-*` só em dialog, alert-dialog, sheet, dropdown, select, toast | ✅ |
| 4 | sem gradiente/brilho | `tokens.test.ts:392`; grep de `gradient` vazio no código | ✅ |
| 5 | durações dos tokens, sem `transition: all` | `tokens.test.ts:353`; regra `transition-all` | ✅ |
| 6 | redução de movimento | `tokens.test.ts:364`; `evidence/smoke-visual.md:134-140` | ✅ |
| 7 | três larguras nomeadas | `components/app/page.test.tsx:91,103`; `global.guard.test.ts:193` | ✅ |
| 8 | dois pontos de quebra, 768 e 1024 | `lib/responsive/use-media-query.test.tsx:41`; `tokens.test.ts:307`; Dialog em 640 px (DSC-04 AC4) e `min-[400px]` no shell | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |

### DSF-05 Marca

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | traçados e cores idênticos, `viewBox` no círculo | `test/design/brand.test.ts:39,44,49,54`; sensor M03 morto | ✅ |
| 2 | 1:1, sem efeito | `components/app/page.test.tsx:40` (largura = altura) | ✅ |
| 3 | ≥ 32 px e respiro | `page.test.tsx:47` | ✅ |
| 4 | alt vazio com nome; "Torcida Jovem" sozinho | `page.test.tsx:25,32` | ✅ |
| 5 | largura e altura | `page.test.tsx:40` | ✅ |
| 6 | sem substituto; dependência em `STATE.md` | `brand.test.ts:89`; `STATE.md:180-183` | ✅ |
| 7 | sem A FERA na área operacional | `brand.test.ts:89`; `components/app/access-layout.test.tsx:50` | ✅ |
| 8 | sem arquivos do `create-next-app` | `brand.test.ts:81,89,94` | ✅ |

### DSF-06 Espécime

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | todas as seções | `app/(dev)/design-system/page.test.tsx:72,85`; `components.test.tsx:44-255` | ✅ |
| 2 | só componentes reais | `design-system.guard.test.ts:36` | ✅ |
| 3 | repouso, foco, desabilitado, inválido | `components.test.tsx:44,69` | ✅ |
| 4 | tabela nas três larguras | `components.test.tsx:124,137`; `evidence/especime.md` | ✅ |
| 5 | 404 em produção | `page.test.tsx:38,48`; `subpages.test.tsx:38`; sensor M12 morto; Smoke T14: `GET /design-system` = 404, título "Página não encontrada \| TJ Platform" | ✅ |
| 6 | migração só após aprovação | `STATE.md:147-148` (WDS-D-033, 2026-10-12) antes das integrações da rodada 6 (`STATE.md:282`) | ✅ |

### DSC-01 Estados, foco e alvos

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | repouso, hover, pressionado, foco, desabilitado distintos | `actions.test.tsx:39,117,130`; espécime aprovado | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 2 | foco ≥ 2 px, ≥ 3:1 no claro e no shell | `actions.test.tsx:117`; `shell-frame.test.tsx:102`; `contrast.test.ts:55`; campo de dinheiro `numeric-text-field.test.tsx:72`. Links de produto e de devolução passaram ao anel do kit (`features/estoque/produtos/produtos-page.tsx:140`, `features/financeiro/lancamentos/lancamento-detail.tsx:162`), conferido no navegador (anel preto de 2 px com afastamento branco de 2 px), **mas sem teste**: mutantes M28 e M29 sobrevivem | ✅ valor no navegador; ⚠️ teste não discrimina (M28, M29) |
| 3 | foco sem vermelho nem erro | `actions.test.tsx:117`; `components/ui/field.test.tsx:374` | ✅ |
| 4 | toque ≥ 44 × 44 | `tokens.test.ts:294`; `data.test.tsx:269`; `evidence/smoke-visual.md:121-130` | ✅ |
| 5 | ≥ 24 × 24 | `tokens.test.ts:262` (alturas 32/36/44); `field.test.tsx:422` | ✅ |
| 6 | ícones decorativos ocultos | `actions.test.tsx:137`; `status-badge.test.tsx:40` | ✅ |
| 7 | estado não só por cor | `display.test.tsx:104,163,177` | ✅ |

### DSC-02 Ações

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | variantes e tamanhos | `actions.test.tsx:22,52,66,74,89` | ✅ |
| 2 | destrutiva sólida, texto ≥ 4,5:1 | `actions.test.tsx:59`; pares em `contrast.test.ts:34` | ✅ |
| 3 | IconButton com nome obrigatório | `actions.test.tsx:215,242` (`@ts-expect-error` conferido pelo `tsc`); sensor M04 morto (34 testes) | ✅ |
| 4 | desabilitado, `aria-busy`, rótulo "…" | `actions.test.tsx:150,178`; `confirm-dialog.test.tsx:206,230` | ✅ |
| 5 | navegação por link | `app-shell.test.tsx:140`; `overlays.test.tsx:488`; `components.test.tsx:57` | ✅ |

### DSC-03 Campos

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | rótulo associado e clicável | `field.test.tsx:32,196` | ✅ |
| 2 | `aria-invalid`, `aria-describedby`, ícone e texto | `field.test.tsx:42,85,233,264`; sensor M06 morto | ✅ |
| 3 | mesma altura, borda e estados | `field.test.tsx:345,353,380` | ✅ |
| 4 | select nativo no kit com cores explícitas | `field.test.tsx:325`; `global.guard.test.ts:177` | ✅ |
| 5 | Select rico só com busca/conteúdo rico | `global.guard.test.ts:177`; `components.test.tsx:34`; nenhum uso do Select rico em `features/` | ✅ |
| 6 | controle e rótulo num alvo | `field.test.tsx:391,410,422` | ✅ |
| 7 | "(opcional)" | `field.test.tsx:210,220` | ✅ |
| 8 | Switch/Radio | Não implementados (WDS-D-022); `conta-dialogs.tsx:112` usa rádio nativo para 2 opções | ✅ |
| 9 | ausentes no espécime | `components.test.tsx:34` | ✅ |

### DSC-04 Sobreposições

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | foco entra e volta | `overlays.test.tsx:231,353,407,463` | ✅ |
| 2 | Esc fecha | `overlays.test.tsx:243,366,419` | ✅ |
| 3 | `overscroll-behavior: contain` | `overlays.test.tsx:306,396,448` | ✅ |
| 4 | < 640 px tela inteira, ações fixas | `overlays.test.tsx:289,314,328`; correção N1 (`evidence/smoke-visual.md:231`) | ✅ |
| 5 | sem diálogo sobre diálogo | Sem teste; inspeção: um diálogo por estado (`features/identity/usuarios/users-page.tsx:55`) | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 6 | região viva; quatro tons com ícone | `display.test.tsx:252,266`; D5 (toast `role=dialog`) registrado em `STATE.md:299` | ✅ |
| 7 | dispensável; pausa em foco/hover | `display.test.tsx:279,293,349,361` | ✅ |
| 8 | Tooltip | não implementado (WDS-D-022) | ✅ |
| 9 | AlertDialog para destrutivas | `overlays.test.tsx:388`; `confirm-dialog.test.tsx:17` | ✅ |

### DSC-05 Dados e navegação interna

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | cabeçalhos, nome, numérico à direita | `components/ui/data.test.tsx:44,55`; `data-list.test.tsx:81,122` | ✅ |
| 2 | tons neutro, sucesso, aviso, erro e informação | `display.test.tsx:95-104`, `status-badge.tsx:9`: 4 tons; informação fundida no neutro (WDS-D-018) | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 3 | status distintos por rótulo e ícone | `status-badge.test.tsx:59`; `lancamentos-page.test.tsx:417-436`; sensor M14 morto | ✅ |
| 4 | kit sem tipos de módulo | `status-badge.test.tsx:67` | ✅ |
| 5 | uso do Card | regra sem valor verificável; `data.test.tsx:183` cobre a estrutura | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 6 | Breadcrumb, último item atual sem link | `data.test.tsx:207,219,233` | ✅ |
| 7 | "Mostrar mais" informa visíveis e total | `data.test.tsx:293,298` | ✅ |
| 8 | Tabs | não implementado (WDS-D-022) | ✅ |

### DSC-06 Domínio

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | `formatBRL`, tabular | `money-value.test.tsx:18,29` | ✅ |
| 2 | sinal de menos; ícone no saldo negativo | `money-value.test.tsx:65`; `saldo-page.test.tsx:43`; sensores M11a e M11b mortos | ✅ |
| 3 | quatro variantes, nenhuma só por cor | `money-value.test.tsx:38,44,51,58,65,76` | ✅ |
| 4 | MoneyField/QuantityField como FND-05 AC4/AC5 | `git diff --stat 7c30c6a f9de714 -- components/app/money-field.test.tsx components/app/quantity-field.test.tsx` vazio, e ambos passam | ✅ |
| 5 | selo de papel neutro em pt-BR | `role-badge.test.tsx:8` | ✅ |

### DSC-07 Testes do kit

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | papel, nome e teclado | `components/ui/tokens.guard.test.ts:75`; `components/app/tokens.guard.test.ts:84` (todo componente referenciado por teste); suítes de `actions`, `field`, `overlays`, `data`, `display` | ✅ |
| 2 | testes existentes passam | gate `pnpm test` | ✅ |
| 3 | alteração cita decisão | `evidence/testes-alterados.md:24-40`, conferido abaixo (Regressão) | ✅ |

### DSP-01 Estrutura de página

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | contêiner com largura nomeada | `page.test.tsx:81,91`; `app/(app)/financeiro/saldo/page.test.tsx:49`; `global.guard.test.ts:193` | ✅ |
| 2 | exatamente um `h1`, descrição, ações | `page.test.tsx:110,126`; `global.guard.test.ts:183` | ✅ |
| 3 | no máximo uma primária | `contas-page.test.tsx:411`; inspeção de `lancamento-detail.tsx` | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 4 | trilha no detalhe | `page.test.tsx:140`; `lancamento-detail.test.tsx:237`; `produto-detail.test.tsx:166` | ✅ |
| 5 | `h1`→`h2`→`h3` sem salto | `states.test.tsx:98,136`; `components/ui/data.test.tsx:358,374` | ✅ |
| 6 | < 768 ações abaixo, primária em largura total | `page.test.tsx:160` | ✅ |

### DSP-02 Estados de tela

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | esqueleto anunciado | `states.test.tsx:14,62,82` | ✅ |
| 2 | vazio com ação de criar | `states.test.tsx:98`; `users-page.layout.test.tsx:122`; `contas-page.test.tsx:468` | ✅ |
| 3 | sem resultado com "Limpar filtros" | `states.test.tsx:113,123` | ✅ |
| 4 | erro com ícone e "Tentar de novo" | `states.test.tsx:34,144`; `app/(app)/layout.test.tsx:120` | ✅ |
| 5 | "Sem acesso" com volta ao início | `require-permission.test.tsx:60,126`; Smoke T14 (conselho em `/admin/usuarios`) | ✅ |
| 6 | estado não desloca o cabeçalho | sem valor verificável | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |

### DSP-03 Feedback e confirmação

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | toast de sucesso e mudança sem recarga | `workflow.test.tsx` (Web V1, inalterado na asserção de toast); Smoke T14 ("Lançamento criado.", "Recebimento registrado.", "Comprovante anexado.") | ✅ |
| 2 | nomeia o registro e a consequência | `confirm-dialog.test.tsx:110`; `dialog-patterns.test.tsx:17`; `workflow.test.tsx:372` | ✅ |
| 3 | verbo da ação | `confirm-dialog.test.tsx:250`; `workflow.test.tsx:66` ("Receber receita", "Pagar despesa") | ✅ |
| 4 | motivo obrigatório (FND-05 AC6) | `confirm-dialog.test.tsx:83,90,98` | ✅ |
| 5 | falha mantém aberto com erro dentro | `confirm-dialog.test.tsx:159,177` | ✅ |
| 6 | indisponível desabilitado com motivo | `ajuste-dialog.test.tsx:197` | ✅ |
| 7 | mensagens FND-06 sem mudança | `git diff 7c30c6a f9de714 -- web/features/*/errors.ts`: só `comprovantes/errors.ts:3,7` (troca `formatBytes` → `formatLimit`, motivada por DSP-07 AC6) | ✅ |

### DSP-04 Formulários

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | coluna única, rótulo acima | `form-layout.test.tsx:33`; `numeric-text-field.test.tsx:55` | ✅ |
| 2 | erro ao sair do campo | `numeric-text-field.test.tsx:127` (só numéricos) | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 3 | foco no primeiro inválido ao enviar | `access-screens.test.tsx:65,81,98`; `produtos-page.test.tsx:185`; `lib/forms/use-focus-invalid-after-submit.test.tsx:40` | ✅ |
| 4 | erro sem campo em alerta no topo | `form-layout.test.tsx:69,85`; `form-error-alert.test.tsx:26` | ✅ |
| 5 | envio desabilita e mostra progresso | `form-layout.test.tsx:137` | ✅ |
| 6 | todo campo declara `autocomplete` e tipo/modo | Telas: todos os campos de texto, seleção e quantidade declaram `autoComplete` (direto ou pelo `NumericTextField`, `numeric-text-field.tsx:100`); os três selects do lançamento em `lancamento-form.tsx:189,224,276`, conferidos no navegador (`autocomplete="off"`, `name` `tipo`, `conta_id`, `forma_pagamento`). Guardas `global.guard.test.ts:213` e `:221` (todo `NativeSelect` e `Textarea` com `name` e `autoComplete`, todo `QuantityField` com `name`); mutantes M18, M23 e M25 mortos. `input type="file"` (`comprovantes-panel.tsx:176`) fica fora de `autocomplete` pela especificação do HTML. Os `Input` e `MoneyField` do espécime não declaram `name`/`autoComplete` (achado F1 da revisão; a guarda não cobre `Input` nem `MoneyField`) | ✅ nas telas (F1 no espécime, ver DSM-06 AC2) |
| 7 | e-mail e código sem correção ortográfica | `create-produto-dialog.tsx:116` (`spellCheck={false}`, `autoCapitalize="none"`); `global.guard.test.ts:253` (todo campo `type="email"` ou `register("email"|"codigo")` com `spellCheck={false}`); `access-screens.test.tsx:55,73`; sensor M17 morto | ✅ |
| 8 | colar não bloqueado | Estático: nenhum `onPaste` em `app/components/features` | ✅ |
| 9 | < 768 ações em largura total, principal por último | `form-layout.test.tsx:119` | ✅ |

### DSP-05 Listagens

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | barra com contagem e "Limpar filtros" | `filter-bar.test.tsx:34,43,48,100` | ✅ |
| 2 | mudança vai à URL | `use-query-state.test.tsx:95,105`; `lancamentos-page.test.tsx:340`; `users-page.layout.test.tsx:65` | ✅ |
| 3 | restaura da URL | `use-query-state.test.tsx:209`; `lancamentos-page.test.tsx:369` | ✅ |
| 4 | inválido ignorado | `codecs.test.ts:38`; `use-query-state.test.tsx:235`; `lancamentos-page.test.tsx:387`; sensor M07 morto | ✅ |
| 5 | ≥ 1024 e conteúdo ≥ 64rem: completa; senão sem P3; sem rolagem do documento | `data-list.test.tsx:140,154,166,173`; Smoke T14 (sem rolagem em 1024 e 1440) | ✅ |
| 6 | 768-1023 por prioridade | `data-list.test.tsx:140,197` | ✅ |
| 7 | < 768 empilhado | `data-list.test.tsx:215,231`; `lancamentos-page.test.tsx:493` | ✅ |
| 8 | tabela pequena mantida | `data-list.test.tsx:302` | ✅ |
| 9 | rolagem com nome, teclado, 1ª coluna fixa | `components/ui/data.test.tsx:149,164` | ✅ |
| 10 | ação de linha identifica o registro | `data-list.test.tsx:319,337`; `lancamentos-page.test.tsx:511-522`; `movimentacoes-panel.test.tsx:417-425`; sensores M05 e M05b mortos | ✅ |
| 11 | uma visível, demais em menu | `data-list.test.tsx:344,356` | ✅ |
| 12 | 50 e +50 | `lancamentos-page.test.tsx:453,462` | ✅ |
| 13 | filtro sobre todos | `lancamentos-page.test.tsx:480` | ✅ |
| 14 | cursor | `users-page.test.tsx` ("Mostrar mais" com `cursor`, `evidence/testes-alterados.md:37`) | ✅ |
| Borda: exatamente 50 | controle não aparece | `lancamentos-page.test.tsx:476` (`[data-slot="show-more"]` nulo e nenhum "Mostrando…"), `produtos-page.test.tsx:366`, `users-page.test.tsx:185`; sensores M13, M13p e M13b mortos | ✅ |

### DSP-06 Permissões

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | sem permissão, ação não aparece | `require-permission.test.tsx:102,113`; `users-page.permissions.test.tsx` | ✅ |
| 2 | só leitura: sem controles de escrita | Smoke T14 (conselho): 0 botões, 0 campos de arquivo e 0 formulários em `main` de início, lançamentos, contas, saldo, produtos e detalhe do lançamento | ✅ |
| 3 | negação da API mostra mensagem de permissão | `components/app/api-error.test.tsx:14` | ✅ |
| 4 | "Sem acesso" igual em todas as áreas | `require-permission.test.tsx:124`; `app/(app)/sem-acesso/page.test.tsx:7` | ✅ |

### DSP-07 Texto

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | pt-BR, frase comum | sem teste nem guarda | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 2 | rótulo começa por verbo | sem teste; link "Detalhes" (`lancamentos-page.tsx:294`) | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 3 | erro do front diz o que houve e o próximo passo | sem teste geral | ⚠️ (aceito em WDS-D-034, `STATE.md:154`) |
| 4 | "…" e espaço não separável entre número e unidade | "…": `global.guard.test.ts:264`; `actions.test.tsx:178`. Espaço não separável: `features/estoque/produto-detail.tsx:107`, `features/estoque/ajustes/ajuste-dialog.tsx:120,127`, `features/estoque/movimentacoes/movimentacao-dialog.tsx:57` (U+00A0 conferido nos bytes); guarda `global.guard.test.ts:240`; testes `produto-detail.test.tsx:50`, `ajuste-dialog.test.tsx:55-65` e `movimentacoes-panel.test.tsx:423-424` afirmam U+00A0; mutantes M24, M26 e M27 mortos; navegador: saldo "3 UN", saldo atual e resultante do ajuste e nome de "Devolver" com U+00A0 | ✅ |
| 5 | progresso termina em "…" | `actions.test.tsx:178` (tipo `${string}…`) | ✅ |
| 6 | limite em MB inteiros, para baixo | `comprovantes-panel.test.tsx:196` ("9 MB") | ✅ |
| 7 | tipo + código curto monoespaçado | `format.test.ts:51,55`; `lancamento-detail.test.tsx:237` | ✅ |

### DSS-01 Shell

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | tokens do shell e do conteúdo | `shell-frame.test.tsx:192` | ✅ |
| 2 | escudo + "Torcida Jovem", link ao início, sem "TJ Platform" | `shell-frame.test.tsx:163,187`; `app-shell.test.tsx:116` | ✅ |
| 3 | grupos com rótulo; ícone e texto | `shell-frame.test.tsx:219,231,250`; `nav-items.test.ts:17` | ✅ |
| 4 | `aria-current`, indicador vermelho, peso maior | `shell-frame.test.tsx:261-279` (`aria-current="page"`, `font-semibold`, `before:bg-sidebar-primary`) | ✅ |
| 5 | primeiro focável "Pular para o conteúdo" | `shell-frame.test.tsx:84,95,109`; `app-shell.test.tsx:127`; sensor M10 morto; `evidence/smoke-visual.md:24` | ✅ |
| 6 | uma região principal; navs nomeadas | `shell-frame.test.tsx:145,151`; `app/(app)/layout.test.tsx:105` | ✅ |
| 7 | nome; "Trocar senha" link; "Sair" botão | `shell-frame.test.tsx:403`; `app-shell.test.tsx:140` | ✅ |
| 8 | FND-05 AC1 e AC8 | `app-shell.test.tsx:55,66,177` | ✅ |
| 9 | texto do shell ≥ 4,5:1 | pares `sidebar-*` em `contrast.test.ts:34` | ✅ |

### DSS-02 Navegação responsiva (inclui FND-05 AC2 revisado por WDS-D-019)

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | ≥ 1024 fixa | `shell-frame.test.tsx:303`; `app-shell.test.tsx:101`; sensor M09 morto; Smoke T14: 1440 e 1024 com "Navegação principal" visível e botão de menu oculto | ✅ |
| 2 | < 1024 gaveta, botão 44 × 44 nomeado | `shell-frame.test.tsx:310,317`; Smoke T14: 820, 768 e 390 sem navegação fixa, botão "Abrir menu" 44 × 44 | ✅ |
| 3 | escolher item fecha | `shell-frame.test.tsx:332`; Smoke T14 (390: gaveta `dialog` "Menu" com 6 links; após escolher, gaveta fechada) | ✅ |
| 4 | foco volta ao botão | `shell-frame.test.tsx:342,354`; Smoke T14 (768, Esc → foco em "Abrir menu") | ✅ |
| 5 | cabeçalho visível; nada inacessível ≥ 320 | `evidence/smoke-visual.md:66-85` (320 a 1440) | ✅ |
| 6 | áreas seguras | `shell-frame.test.tsx:209`; `viewportFit` sem `cover` registrado (`STATE.md:288`) | ✅ |
| 7 | navegador real em 390, 768, 820, 1024, 1440 | `evidence/smoke-visual.md:24,70-85`; Smoke T14 nas cinco larguras | ✅ |

### DSS-03 Acesso

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | escudo e nome | `access-screens.test.tsx:28`; `access-layout.test.tsx:23` | ✅ |
| 2 | mesmo layout estreito | `access-layout.test.tsx:60,71,80` | ✅ |
| 3 | FND-03, ACS-01 a 03 | `app/(public)/entrar/login-form.test.tsx` e testes de acesso da V1, todos passando | ✅ |
| 4 | `autocomplete` de autenticação; e-mail sem correção | `access-screens.test.tsx:38,55,73,89` | ✅ |
| 5 | sem ilustração | `access-layout.test.tsx:50` | ✅ |
| 6 | < 768 sem rolagem, campos em largura total | `access-screens.test.tsx:46`; `evidence/smoke-visual.md:72` | ✅ |

### DSS-04 Erros

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | erro no shell com "Tentar de novo" e "Voltar ao início" | `app/error-pages.test.tsx:38,48,56` | ✅ |
| 2 | erro raiz com escudo e "Tentar de novo" | `error-pages.test.tsx:77,89` | ✅ |
| 3 | "Página não encontrada" com link | `error-pages.test.tsx:65`; Smoke T14 (`/design-system` em produção) | ✅ |
| 4 | sem detalhe técnico | `error-pages.test.tsx:14,38,77` | ✅ |
| 5 | `h1` com foco | `error-pages.test.tsx:38,65,77` | ✅ |

### DSS-05 Metadados

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | `lang="pt-BR"`, `theme-color` do shell | `fonts.test.ts:75,84` | ✅ |
| 2 | "Página \| TJ Platform" | `fonts.test.ts:84`; Smoke T14 ("Início \| TJ Platform") | ✅ |
| 3 | ícone derivado do escudo | `brand.test.ts:70,74` | ✅ |
| 4 | dependência de asset reduzido | `STATE.md:180` | ✅ |

### DSS-06 Início

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | atalhos da lista da navegação | `app/(app)/inicio/atalhos.test.tsx:53,69` | ✅ |
| 2 | links com ícone, nome e descrição | `atalhos.test.tsx:53` | ✅ |
| 3 | sem indicadores | `atalhos.test.tsx:87` | ✅ |
| 4 | < 768 uma coluna | `atalhos.test.tsx:76` | ✅ |

### DSM-01 Migração sem regressão

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | ACs da Web V1 válidos | 1774/1774 testes; testes alterados conferidos (seção Regressão) | ✅ |
| 2 | nada fora de `web/`, `.specs/`, `docs/development/` e `.gitignore` (AC revisado em `7499d87`, `spec/05-migracao-web-v1.md:67`) | `git diff --name-only 7c30c6a 7499d87` fora de `web/`: só `.specs/` (`features/web-design/`, `LESSONS.md`, `lessons.json`), `docs/development/web.md` e `.gitignore` | ✅ |
| 3 | sem dependência nova | `git diff --name-only develop -- web/package.json web/pnpm-lock.yaml` vazio | ✅ |
| 4 | alteração de teste cita decisão; lista no relatório | `evidence/testes-alterados.md:24-40`; conferido abaixo | ✅ |
| 5 | nenhum teste removido ou enfraquecido | seção Regressão | ✅ |
| 6 | espécime aprovado antes | `STATE.md:147-148` | ✅ |

### DSM-02 Identity

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | cabeçalho, filtros, lista responsiva | `users-page.layout.test.tsx:21` | ✅ |
| 2 | filtro na URL e `active`/`role` na API | `users-page.layout.test.tsx:65-82` (`{ limit: "50", active: "false", role: "TESOURARIA" }`) | ✅ |
| 3 | menu nomeado pelo usuário | `users-page.layout.test.tsx:176`; `evidence/smoke-visual.md:31` | ✅ |
| 4 | troca de senha no layout de acesso | `app/(app)/conta/senha/page.test.tsx:41` | ✅ |
| 5 | USR-01 a 05, ACS-01 a 03 | testes da V1 passando | ✅ |

### DSM-03 Financeiro

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | padrões nas telas | `contas-page.test.tsx:399`; `lancamento-detail.test.tsx:237`; `saldo-page.test.tsx:79`; `comprovantes-panel.test.tsx:389` | ✅ |
| 2 | MoneyValue tabular | `lancamento-detail.test.tsx:74,237`; `saldo-page.test.tsx:79` | ✅ |
| 3 | quatro status distintos | `lancamentos-page.test.tsx:417-436`; sensor M14 morto | ✅ |
| 4 | tipo, status, conta, período na URL | `lancamentos-page.test.tsx:340-410` | ✅ |
| 5 | < 768 data, tipo, conta, líquido, status | `lancamentos-page.test.tsx:493-506` | ✅ |
| 6 | link de detalhe nomeado | `lancamentos-page.test.tsx:511-522` | ✅ |
| 7 | hierarquia por recuo | `contas-page.test.tsx:418-419` | ✅ |
| 8 | FWB-01 a 05 | testes da V1 passando; Smoke T14 (criar, receber, anexar) | ✅ |

### DSM-04 Estoque

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | padrões nas telas | `produtos-page.test.tsx:306,371`; `produto-detail.test.tsx:166` | ✅ |
| 2 | busca na URL | `produtos-page.test.tsx:264` | ✅ |
| 3 | < 768 histórico empilhado | `movimentacoes-panel.test.tsx:429` | ✅ |
| 4 | "Devolver" nomeia a origem | `movimentacoes-panel.test.tsx:417-425` | ✅ |
| 5 | diálogos em tela inteira < 640 | `overlays.test.tsx:289,314`; `evidence/smoke-visual.md:179` | ✅ |
| 6 | tabular; negativo com ícone e texto | `ajuste-dialog.test.tsx:197`; `produto-detail.test.tsx:166` | ✅ |
| 7 | EWB-01 a 04 | testes da V1 passando | ✅ |

### DSM-05 Limpeza

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | sem contêiner, largura ou título próprios | `global.guard.test.ts:183,188,193,208` | ✅ |
| 2 | sem `<select>` fora do kit | `global.guard.test.ts:177` | ✅ |
| 3 | `select-field` e filtro de data substituídos | `git ls-files | grep select-field` vazio | ✅ |
| 4 | guarda em todo o front | `global.guard.test.ts:163,168`; sensor M02 morto | ✅ |
| 5 | sem variantes `dark:` | regra `dark-variant` (`guard.ts:87`) | ✅ |

### DSM-06 Verificação

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | gates passam | seção Gate Check | ✅ |
| 2 | `web-design-guidelines` com cada achado corrigido ou registrado como conflito | A1 a A3 e O1 da iteração 2 corrigidos. **Restam F1 a F3** (seção da revisão), sem correção nem conflito | ❌ |
| 3 | navegador real, produção, 5 larguras, todas as telas | `evidence/smoke-visual.md:66-85,152-181` | ✅ |
| 4 | percurso só por teclado | `evidence/smoke-visual.md:87-97`; N2 e N3 corrigidos em `2903f32` | ✅ |
| 5 | 200% e 320 px | `evidence/smoke-visual.md:101-117` | ✅ |
| 6 | verificador independente com `validation.md` | este arquivo | ✅ |
| 7 | divergência vira correção | N1, N2, N3 corrigidos (`evidence/smoke-visual.md:229-231`) | ✅ |



**Status**: 219 ACs. **206 ✅, 1 ❌ (DSM-06 AC2: D2 a D4), 12 ⚠️ aceitas em WDS-D-034** (DSF-03 AC4, DSF-04 AC8, DSC-01 AC1, DSC-04 AC5, DSC-05 AC2, DSC-05 AC5, DSP-01 AC3, DSP-02 AC6, DSP-04 AC2, DSP-07 AC1, AC2, AC3).
- DSP-04 AC6 e DSP-07 AC4 passaram a ✅.
- DSC-01 AC2 tem o valor certo no navegador, mas a correção O1 está sem teste (D1).

---

## Regressão da Web V1 e testes alterados

- Os 69 arquivos de teste da base passam (gate `pnpm test`, 1787/1787) e nenhum foi removido.
- **Alterados em `fb62b99`** (arquivos da base): `features/estoque/produto-detail.test.tsx` (5 asserções), `features/estoque/ajustes/ajuste-dialog.test.tsx` (10) e `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` (13).
  - Conferido por comparação: com U+00A0 trocado por espaço comum, cada arquivo de HEAD é **byte a byte igual** ao de `7499d87`. A única mudança é o espaço não separável no valor esperado.
  - Motivo: DSP-07 AC4, citado na mensagem do commit.
  - A asserção fica mais estrita, não mais fraca. Prova disso: M24, M26 e M27 (voltar ao espaço comum) derrubam esses testes.
  - Exceção: `movimentacoes-panel.test.tsx:461` usa `getByText` com espaço comum, porque o normalizador do Testing Library junta qualquer espaço. O nome acessível da mesma movimentação é afirmado com U+00A0 em `movimentacoes-panel.test.tsx:423-424`.
  - DSM-01 AC4 pede essa lista no relatório de verificação, e ela está aqui. `evidence/testes-alterados.md` não foi atualizado depois de `663febe`. Vale acrescentar as linhas de `fb62b99` e das correções da iteração 1 (acréscimos e `min-h-control-*` em `actions.test.tsx`), mas isso não bloqueia, porque o AC aponta o relatório de verificação.
- As alterações anteriores (os 15 arquivos de `evidence/testes-alterados.md` e as correções da iteração 1) não mudaram de novo. Os casos removidos com a remoção de alias continuam cobertos (`lancamentos-page.test.tsx:422,431,435`; `status-badge.test.tsx:15-21,76`).

---

## FND-05 AC2 (revisado por WDS-D-019)

- **Teste**: `shell-frame.test.tsx:303,310`; `app-shell.test.tsx:101`.
- **Mutante**: M09 morto.
- **Navegador** (build `13e56c5`):
  - Em 390, 768 e 820 px não há navegação fixa e o botão "Abrir menu" mede 44×44.
  - Em 1024 e 1440 px a "Navegação principal" fica fixa e o botão está oculto.
  - Nenhuma das cinco larguras rola na horizontal.

---

## Discrimination Sensor

Worktree temporário em `13e56c5 --detach`, com `pnpm install --frozen-lockfile`. Cada falha foi injetada sozinha, os testes relevantes rodaram com `pnpm exec vitest run` e o worktree voltou com `git checkout -- .` antes da próxima. O log registra o status do worktree vazio após cada reversão.

| # | Falha | Resultado |
| --- | --- | --- |
| M01-M14 (16 falhas do design e variantes: M01, M02, M03, M04, M05, M05b, M06, M07, M08, M09, M10, M11a, M11b, M12, M13, M13b, M14) e M13p | as mesmas das iterações 1 e 2 | **todas MORTAS** |
| M17-M22 | correções da iteração 1 (`spellCheck`, `name` do motivo, `focus-within`, `whitespace-nowrap`, `tap-highlight`, `aria-live`) | **todas MORTAS** |
| M23 | `lancamento-form.tsx:189` sem `autoComplete` | **MORTO** (guarda geral) |
| M24 | `produto-detail.tsx:107` volta ao espaço comum | **MORTO** (18 testes) |
| M25 | `ajuste-dialog.tsx:96` sem `name` no `QuantityField` | **MORTO** (guarda geral) |
| M26 | `ajuste-dialog.tsx:127` volta ao espaço comum | **MORTO** (2) |
| M27 | `movimentacao-dialog.tsx:57` volta ao espaço comum | **MORTO** (2) |
| M28 | `produtos-page.tsx:140` sem o anel do kit | **SOBREVIVEU** (`features/estoque` e `test/design`: 247/247) |
| M29 | `lancamento-detail.tsx:162` sem o anel do kit | **SOBREVIVEU** (`features/financeiro` e `test/design`: 435/435) |

**Resultado do sensor**: 31 injetadas, 29 mortas, 2 sobreviventes (M28 e M29).

**Isolamento**: o `git status --porcelain` do principal estava vazio (0 bytes) antes. Depois de `git worktree remove --force` (com o diretório apagado por `\\?\…`) e `git worktree prune`, continuou vazio (0 bytes). `git worktree list` não lista mais o temporário. HEAD continua `13e56c5`.

---

## Revisão `web-design-guidelines` (refeita do zero)

Apliquei a revisão a `web/app`, `web/components` e `web/features` (arquivos não-teste), regra a regra. As varreduras mecânicas foram:
- todos os controles de formulário, quanto a `name` e `autocomplete`;
- todo `Link`, `a` e `button` sem variante do kit, quanto a `hover:` e foco;
- placeholders, `outline-none`, números com unidade, aspas e `...` literais;
- `onClick` em `div` e `span`, imagens, `autoFocus`, zoom e `onPaste`.

Cada achado foi lido no trecho.

### Achados sem correção e sem conflito registrado (reprovam)

```text
## app/(dev)/design-system (espécime, 404 em produção) - D2
app/(dev)/design-system/acesso/example-login-form.tsx:47 - Input sem name
app/(dev)/design-system/acesso/example-login-form.tsx:60 - Input sem name
app/(dev)/design-system/example-listing.tsx:58 - Input de busca sem name nem autoComplete
app/(dev)/design-system/fields-section.tsx:52 - MoneyField sem name
app/(dev)/design-system/fields-section.tsx:55 - MoneyField sem name
app/(dev)/design-system/fields-section.tsx:58 - MoneyField sem name
app/(dev)/design-system/fields-section.tsx:101 - Input sem name nem autoComplete
app/(dev)/design-system/fields-section.tsx:108 - Input sem name nem autoComplete
app/(dev)/design-system/fields-section.tsx:115 - Input sem name nem autoComplete
app/(dev)/design-system/fields-section.tsx:121 - Input sem name nem autoComplete
app/(dev)/design-system/fields-section.tsx:127 - Input sem name nem autoComplete
app/(dev)/design-system/overlays-section.tsx:69 - Input sem name nem autoComplete
app/(dev)/design-system/overlays-section.tsx:72 - MoneyField sem name

## features/financeiro/comprovantes/comprovantes-panel.tsx - D3
features/financeiro/comprovantes/comprovantes-panel.tsx:176 - input type="file" sem name

## components/app/shell-frame.tsx - D4
components/app/shell-frame.tsx:155 - link da marca (escudo + nome) sem estado hover
```

### Achados das iterações 1 e 2: todos corrigidos

- Os 22 achados da iteração 1 e os A1 a A3 da iteração 2 estão corrigidos (ver as tabelas de conferência).
- O1, o anel nos links, também está corrigido, mas sem teste (D1).

### Achados casados com conflito ou decisão registrada (não reprovam)

| Achado | Registro |
| --- | --- |
| Title Case; envio desabilitado sem motivo; URL sem `nuqs`; listas > 50 sem virtualização; dark mode | `STATE.md:193-197` |
| `safe-area` sem `viewportFit: cover`; placeholder `0,00` do campo de dinheiro; toast anima `height`; data de calendário sem `Intl`; anel do item de menu no hover | `STATE.md:286-292` |
| Toast `role="dialog"` (D5) e toast de erro `polite` | `STATE.md:299`, `STATE.md:304` |
| Sem aviso de alterações não salvas | `STATE.md:262` |
| `lib/forms/use-focus-invalid-after-submit.ts` | `STATE.md:152` |

### Conferido e sem achado

- **Foco**:
  - `outline-none` só aparece em alvos de foco programático com `tabIndex={-1}`.
  - Os links de produto e de devolução têm o anel do kit.
  - O link do cartão de início usa anel `after:` e herda o `hover:` do `Card` (`atalhos.tsx:75`).
  - O link de pular só aparece no foco.
- **Formulários e eventos**:
  - Todos os campos das telas declaram `autoComplete`; `type="file"` está fora dessa regra.
  - Nenhum `onClick` em `div` ou `span`, e nenhum `autoFocus`, `user-scalable`, `maximum-scale` ou `onPaste`.
- **Texto**:
  - Todos os placeholders literais terminam em "…".
  - Não há `...` literal nem aspa reta em texto exibido.
  - Toda quantidade com unidade usa espaço não separável.
- **Base e toque**:
  - A camada base tem `touch-action` e `tap-highlight`.
  - Valores em reais ficam em `whitespace-nowrap`.

---

## SSR e hidratação

- Nenhum Client Component novo.
- `useSearchParams` continua sob `Suspense` e `useMediaQuery` continua com snapshot de servidor fixo.
- Nenhum `window`, `Date.now`, `Math.random` ou `new Date()` em renderização.
- O U+00A0 entra em template literal, idêntico no servidor e no cliente.
- No navegador: 0 avisos de hidratação, 0 exceções e 0 avisos de console.

---

## Propriedade e escopo

- Fora de `web/` mudaram só `.specs/`, `docs/development/web.md` e `.gitignore` (DSM-01 AC2 revisado, `spec/05-migracao-web-v1.md:67`).
- Sem diff em `api/`, `web/package.json`, `web/pnpm-lock.yaml`, `web/lib/api/`, `web/lib/session/session-provider.tsx` e `features/*/hooks.ts`.
- `fb62b99` e `13e56c5` mexem só em atributos, texto de quantidade, classes de foco, testes e espécime. Nenhuma regra de negócio, permissão ou contrato foi alterada.

---

## Smoke (amostra do verificador, build de produção `13e56c5` em `:3000`, API em `:8080`)

Usei o Edge headless por CDP, com scripts em `…/scratchpad/verify3/smoke/` e perfil temporário encerrado no fim. A senha ficou só em memória. As capturas estão no scratchpad, fora do repositório.

A primeira tentativa de login do script falhou: o envio saiu antes de o formulário hidratar, com o servidor recém-iniciado, e o script não chegou a mandar o `POST /auth/login`. Com espera da hidratação, o login deu `200` e foi para `/inicio`. Foi um problema de tempo do script, não da aplicação.

| Conferência | Resultado |
| --- | --- |
| Login presidente | `/inicio`, "Início \| TJ Platform" |
| FND-05 AC2 em 390, 768, 820, 1024 e 1440 | gaveta com botão de 44×44 abaixo de 1024; navegação fixa em 1024 e 1440; sem rolagem horizontal |
| Anel no link de produto, por teclado | "Bandeira grande" com `:focus-visible` e `box-shadow 0 0 0 2px #fff, 0 0 0 4px #111` (recorte `1440-produto-link-foco.png`) |
| Anel no link de devolução, por teclado | link `f655839f` com o mesmo anel (recorte `1440-devolucao-link-foco.png`) |
| Quantidades | saldo "3 UN", saldo atual e resultante do ajuste e nome de "Devolver" ("entrada de 5 UN em …") com U+00A0; `QuantityField` do ajuste com `name="quantidade"` e `autocomplete="off"` |
| Selects de "Novo lançamento" | Tipo, Conta e Forma de pagamento com `autocomplete="off"` |
| Console | 0 avisos de hidratação, 0 exceções, 0 avisos |

Não repeti nesta iteração a comparação de pixels de lançamentos, usuários e produtos. As únicas mudanças visíveis desde a iteração 2, que deu 0,00% em usuários e produtos e só diferenças de dados em lançamentos, são o anel de foco, que não aparece em repouso, e o U+00A0, que não altera o desenho.

---

## Code Quality

| Principle | Status |
| --- | --- |
| Minimum code / surgical changes | ✅ |
| No scope creep | ✅ |
| Matches patterns | ✅ |
| Spec-anchored outcome check | ❌ DSM-06 AC2 |
| Per-layer coverage | ⚠️ anel dos links sem teste (D1); guarda de atributos sem `Input` e `MoneyField` (origem de D2) |
| Every test maps to a requirement | ✅ |
| Documented guidelines followed | ❌ D2 a D4 |

---

## O que falta (decisão do mantenedor, iteração 3 de 3)

Cada item pode ser corrigido ou registrado em `STATE.md`. Todos são pequenos e nenhum toca regra de negócio, API ou permissão.

1. **D1, sensor M28 e M29** (DS-ES e DS-FB). Os links de produto e de devolução estão com o anel do kit, mas nenhum teste o afirma.
   - Correção: um teste por link que afirme `focus-visible:ring-2` e `focus-visible:ring-ring`, no padrão de `actions.test.tsx:117`.
   - Alternativa: aceitar e registrar que esse estilo é coberto só pelo navegador.
2. **D2** (DS-E). Os `Input` e `MoneyField` do espécime estão sem `name` e, nos `Input`, sem `autoComplete` (13 ocorrências, lista acima).
   - Correção: acrescentar os atributos e estender a guarda de `global.guard.test.ts:221` a `Input` e `MoneyField`.
   - Alternativa: registrar em `STATE.md` que o espécime, rota interna e 404 em produção (WDS-D-017), fica isento das regras de `name` e `autocomplete`.
3. **D3** (DS-FA). O campo de arquivo do comprovante está sem `name` (`comprovantes-panel.tsx:176`).
   - Correção: `name="comprovante"`.
   - Alternativa: registrar como conflito, porque o campo não pertence a formulário enviado e a regra de `autocomplete` não se aplica.
4. **D4** (DS-S). O link da marca no shell está sem `hover` (`shell-frame.tsx:155`).
   - Correção: um realce de `hover` com tokens do shell.
   - Alternativa: registrar como conflito, porque a marca não muda de aparência ao passar o ponteiro.

---

## Requirement Traceability Update

| Requirement | New Status |
| --- | --- |
| DSF-01 a DSF-06 | ✅ Verified (⚠️ aceitas: DSF-03 AC4, DSF-04 AC8) |
| DSC-01 a DSC-07 | ✅ Verified (⚠️ aceitas: DSC-01 AC1, DSC-04 AC5, DSC-05 AC2/AC5; D1 no teste de DSC-01 AC2) |
| DSP-01 a DSP-07 | ✅ Verified (⚠️ aceitas indicadas) |
| DSS-01 a DSS-06 | ✅ Verified (FND-05 AC2 revisado reconfirmado) |
| DSM-01 a DSM-05 | ✅ Verified |
| DSM-06 | ❌ Needs Fix ou decisão (AC2: D2 a D4) |

---

## Summary

**Overall**: ❌ Not Ready. É a iteração 3 de 3, a última; a decisão sobre D1 a D4 vai ao mantenedor.

- **Spec-anchored check**: 206/219 ✅; 1 ❌; 12 lacunas de precisão aceitas em WDS-D-034.
- **Sensor**: 31 injetadas, 29 mortas, 2 sobreviventes.
- **Gate**: todos com saída 0 (1787/1787 testes).

**What works**:
- Fundação, kit, padrões, shell e telas da Web V1 estão migrados e cobertos.
- Todas as lacunas das iterações 1 e 2 estão fechadas, com teste, mutante e navegador, exceto o teste do anel dos links (D1).
- FND-05 AC2 revisado foi reconfirmado.
- Permissões, API e regras de negócio estão intocadas.

**Next steps**: o mantenedor decide D1 a D4 (corrigir ou registrar). Se escolher corrigir, uma reverificação curta, restrita a esses itens, fecha o ciclo.

---

## Fechamento determinístico

`python .claude/skills/tlc-spec-driven/scripts/validate_state.py web-design --root C:/Users/Michels/Desktop/torcida-jovem` → exit 1: "web-design: validation.md verdict is FAIL - route the ranked gaps to fix tasks, then re-verify (feature is not done)". Coerente com o veredito: a feature não está concluída.

---

## Iteração 4 (excepcional, autorizada — WDS-D-035)

- **Data**: 2026-10-13
- **Verificador**: agente independente; não escreveu o código nem as evidências.
- **Escopo**: restrito por WDS-D-035 (`STATE.md:156`):
  - D1 a D4;
  - item da T2 da DS-S;
  - `evidence/testes-alterados.md`;
  - matriz dos 219 critérios;
  - sensores relacionados.

  As iterações 1 a 3 acima ficam como estão.
- **Commits verificados**: `13f5ad7` (D1), `9e95fb9` (D3), `131a5c4` (D2), `023558b` (D4) e `86f6a30` (evidências e matriz), sobre `1056707`. HEAD `86f6a30`.
- **Veredito**: ✅ **PASS**.
  - D1 a D4 estão corrigidos, com teste que discrimina.
  - Reproduzi a T2 da DS-S de forma independente.
  - A matriz confere; corrigi só uma linha de formatação (abaixo).
  - Gates verdes. Nenhum defeito material novo.

### 1. D1: anel de foco nos links de produto e de devolução

- **Código**: `features/estoque/produtos/produtos-page.tsx:140` e `features/financeiro/lancamentos/lancamento-detail.tsx:162` têm `rounded-sm outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset`.
- **Testes**: `produtos-page.test.tsx:336` e `lancamento-detail.test.tsx:351` exigem essas classes e recusam `destructive` ou `brand` no foco.
- **Mutantes**: reaplicados por mim num worktree descartável em `86f6a30` (`<scratchpad>/v4wt`, `pnpm install --frozen-lockfile`). Cada um foi aplicado sozinho e revertido antes do seguinte.

| # | Mutação | Testes rodados | Resultado |
| --- | --- | --- | --- |
| M28 | tira as classes do anel do link de produto | `produtos-page.test.tsx` | **MORTO** (1 falha, 26 passam) |
| M29 | tira as classes do anel do link de devolução | `lancamento-detail.test.tsx` | **MORTO** (1 falha, 27 passam) |
| M28r | `focus-visible:ring-ring` → `focus-visible:ring-destructive` no link de produto | `produtos-page.test.tsx` | **MORTO** (1 falha) |
| M29o | tira só o afastamento (`ring-offset-*`) do link de devolução | `lancamento-detail.test.tsx` | **MORTO** (1 falha) |

- **Navegador** (build de `86f6a30` em `:3142`, foco só por Tab):
  - "Bandeira grande" (12 Tabs) e `f655839f` (14 Tabs) ficam com `:focus-visible`.
  - `box-shadow` computado: `rgb(255,255,255) 0 0 0 2px, rgb(17,17,17) 0 0 0 4px`.
  - `outline: none`; raio de 4 px.

### 2. D2: campos do espécime

- **Atributos**:
  - Os 13 campos listados na iteração 3 têm `name` próprio e `autoComplete`.
  - Login de exemplo: `email`/`username` e `senha`/`current-password`, nunca `off`.
  - Busca e campos de demonstração: `off`.
  - O diff de `131a5c4` só acrescenta `name` e `autoComplete`. Rótulos, ids e valores ficam iguais.
- **Componentes**: `git diff fb62b99 HEAD -- web/components` só mostra `shell-frame.tsx` e o teste dele, que são de D4. `Input`, `MoneyField` e `NumericTextField` estão intactos; o `MoneyField` já repassava `name` (`numeric-text-field.tsx:22,97`).
- **Guarda** (`global.guard.test.ts:246`): cobre todo `Input` (`name` e `autoComplete`) e todo `MoneyField` (`name`) de `app`, `components` e `features`. As exceções são justificadas:
  - `components/ui/` fica fora porque é onde o componente é definido;
  - `type="file"` só exige `name`, porque o HTML não aplica `autocomplete` a esse tipo;
  - `{...field}` e `{...register()}` contam como `name`.
- **Testes de página**:
  - `page.test.tsx:77`: todo campo do espécime tem `name`, sem repetição, e os de texto têm `autocomplete`.
  - `subpages.test.tsx:75`.
- **Mutantes da guarda**:

| # | Mutação | Resultado |
| --- | --- | --- |
| G1 | tira `name` do `Input` "Descrição" do espécime (`fields-section.tsx:101`) | **MORTO** (guarda e `page.test.tsx:77`) |
| G2 | tira `autoComplete` de um `Input` real (`features/financeiro/contas/conta-dialogs.tsx:103`, com `{...register}`) | **MORTO** (guarda) |
| G3 | tira `name` do `MoneyField` do diálogo do espécime | **MORTO** (guarda) |
| G4 | tira `name="comprovante"` | **MORTO** (guarda e `comprovantes-panel.test.tsx`) |
| G5 | troca `current-password` por `off` no login de exemplo | **MORTO** (`subpages.test.tsx:75`) |
| G6 | `name` repetido entre dois `MoneyField` do espécime | **MORTO** (`page.test.tsx:77`) |

- **Produção**: `/design-system` em `:3142` responde `404`, com título "Página não encontrada | TJ Platform" e `h1` "Página não encontrada".

### 3. D3: campo de arquivo do comprovante

- **Código e teste**: `comprovantes-panel.tsx:179` tem `name="comprovante"`; o teste "declara name=…" de `comprovantes-panel.test.tsx` passa.
- **Navegador**:
  - `name="comprovante"`, sem `autocomplete`.
  - Rótulo "Anexar comprovante".
  - `accept=".pdf,.jpg,.jpeg,.png,.webp"`.
  - Fora de `<form>`; há um único campo com esse nome.
- **Contrato inalterado**:
  - O envio continua sendo `FormData.append("file", …)` (`features/financeiro/comprovantes/hooks.ts:41`).
  - Os demais testes do painel passam sem mudança.
  - Não há diff em `hooks.ts` nem em `lib/api`.

### 4. D4: hover da marca do shell

- **Código** (`shell-frame.tsx:157`): acrescenta `transition-colors duration-fast hover:bg-sidebar-accent`, os mesmos tokens e a mesma transição dos itens da navegação (`shell-frame.tsx:53,254,258`).
  - Nenhum token, cor ou animação nova.
  - Escudo, raio, alvo e foco não mudam.
- **Teste** (`shell-frame.test.tsx:182`): exige as classes e o anel do shell, e recusa outro `hover:` e qualquer animação além da transição de cor. Mutantes:
  - H1, tirar o `hover:`: **MORTO**.
  - H2, acrescentar `hover:scale-105`: **MORTO**.
- **Navegador** (`:3142`, 1440 px):
  - Repouso: fundo `rgba(0,0,0,0)` sobre o cabeçalho `rgb(17,17,17)`.
  - Hover (`:hover` verdadeiro): fundo `rgb(42,44,46)`, que é `--sidebar-accent` `#2a2c2e`; texto branco; `animation: none`; `transform: none`.
  - Escudo continua com 32 px; caixa de 156 × 44.
  - Foco no 2º Tab: `:focus-visible`, anel branco de 2 px, `outline: none`.
  - O contraste do branco sobre `sidebar-accent` já está nos pares calculados (`lib/design/contrast-pairs.ts`).

### 5. T2 da DS-S: erro dentro do shell numa exceção real

**O que conferi na evidência**:
- `evidence/erro-no-shell.md` descreve um worktree descartável com rotas temporárias de cliente e de servidor.
- As capturas de cliente e de servidor são idênticas byte a byte (mesmo md5). Isso é coerente com a mesma página de erro desenhada nos dois casos, mas por isso mesmo reproduzi o teste em vez de confiar nelas.
- Não há rota `zz-falha`, `__zzRenders` nem `SIMULACAO` em `web/`, nem no histórico (`git log --all -S zz-falha -- web` vazio).
- `git worktree list` não lista `err-sim`. O worktree `ds-v3fix` está limpo e sem `zz*`.

**Reprodução independente**:
- Worktree descartável `v4wt`, com duas rotas minhas:
  - `app/(app)/zz-v4/cliente`: Client Component que lança erro;
  - `zz-v4/servidor`: Server Component `force-dynamic` que lança erro.
- Build de produção e `next start -p 3143`, com a API em `:8080`.
- Sessão do presidente aberta por `:3000`; a senha ficou só em memória.

| Verificação | Cliente | Servidor |
| --- | --- | --- |
| Boundary acionado | `app/(app)/error.tsx` ("Não foi possível mostrar esta página…"); não é o `global-error` nem a tela padrão do Next | igual |
| Shell | `[data-slot=shell]`, "Navegação principal", 1 `main` | igual |
| `h1` | "Algo deu errado", com foco ao abrir e depois de "Tentar de novo"; título "Algo deu errado \| TJ Platform" | igual |
| Detalhe técnico | nenhum no texto nem no HTML (`V4SIM`, `digest`, `Error:`, `stack`) | nenhum; o `digest` aparece só no log do servidor |
| Sem loop | 5 s parado: 0 requisições, renderizações 6 → 6. "Tentar de novo": 1 `GET ?_rsc`, renderizações 6 → 12. Mais 5 s: 0 requisições, 12 → 12 | 5 s parado: 0 requisições. "Tentar de novo": 1 `GET ?_rsc`, 2 renderizações no servidor no total. Depois: 0 requisições |
| 390 px | página de erro no shell, foco no `h1` | igual |
| "Voltar ao início" | `/inicio`, `h1` "Início", sem erro | igual |
| Console | só o `console.error` do React com a exceção capturada | só o "Minified React error #441" |

**Limpeza**:
- O `next start` de 3143 foi encerrado.
- `git worktree remove --force` falhou por caminho longo; apaguei o diretório com `robocopy /MIR` e rodei `git worktree prune`.
- `git worktree list` sem `v4wt`; `git status --porcelain` do principal vazio.

**Decisão**: o checkbox da T2 (`tasks/04-shell-e-acesso.md`) está **justificado**. A exceção foi real, no cliente e no servidor, sem código artificial no produto.
- Ressalva: a exceção no layout raiz (`global-error`, DSS-04 AC2) continua coberta só por teste (`error-pages.test.tsx:77,89`), como a própria evidência registra. Isso não faz parte do item da T2.

### 6. Documentação

**`testes-alterados.md`**:
- Cobre as trocas de `fb62b99`. Conferi com `git show fb62b99`:
  - `produto-detail.test.tsx`: 5 asserções;
  - `ajuste-dialog.test.tsx`: 10;
  - `movimentacoes-panel.test.tsx`: 13.

  As contagens e os trechos citados batem.
- Explica a exceção de `movimentacoes-panel.test.tsx:461`: o `getByText` normaliza `\s`, que inclui U+00A0. O caractere certo é exigido pela guarda `global.guard.test.ts:267` e pelas linhas 423-424.
- Lista os 7 testes acrescentados nos quatro commits de correção (1787 → 1794).
- Lacuna menor, não bloqueia: os testes que `fb62b99` acrescentou à guarda não estão na tabela de acréscimos. Não são asserções trocadas, e o que DSM-01 AC4 exige são as trocas.

**`matriz-criterios.md`**:
- Contei por script os itens numerados de cada bloco "Acceptance Criteria" de `spec/01` a `05`: **219** (DSF 47, DSC 46, DSP 53, DSS 35, DSM 38). A contagem bate com a matriz por requisito.
- Há exatamente uma linha por AC, numeradas de 1 a N em cada requisito, sem linhas a mais.
- Cada linha tem uma única situação:
  - comprovado: 207 (15 com a anotação "mudou nesta rodada");
  - atendido por decisão documentada: 12;
  - pendente: 0;
  - reprovado: 0.
- Os 12 "por decisão" são os 12 ⚠️ desta validação e todos citam WDS-D-034. Nenhum AC com lacuna aceita está marcado como comprovado.
- Amostra: conferi 40 referências de linhas "comprovado". Todas existem em HEAD e o teste no local cobre o AC. Critérios amostrados:
  - DSF-05 AC1, DSF-06 AC5;
  - DSC-01 AC2, DSC-02 AC3, DSC-03 AC7, DSC-04 AC2, DSC-06 AC2;
  - DSP-02 AC3, DSP-03 AC4, DSP-04 AC6 e AC7, DSP-05 AC4, AC12 e borda, DSP-07 AC4;
  - DSS-01 AC2, AC4, AC5 e AC7, DSS-02 AC1, AC3 e AC4, DSS-04 AC1, AC4 e AC5, DSS-05 AC1, DSS-06 AC3;
  - DSM-02 AC2, DSM-03 AC7, DSM-05 AC2.
- As referências remapeadas em `global.guard.test.ts`, `shell-frame.test.tsx` e `produtos-page.test.tsx` estão corretas.
- **Correção feita nesta iteração**: na tabela "Resumo por requisito", a linha **Total** estava colada à linha DSM-06 (`…| 0 || **Total**…`), o que quebrava a tabela. Separei as duas linhas; os números não mudaram.

### 7. Gates e escopo (repositório principal, `web/`, `API_URL=http://localhost:8080`)

| Gate | Saída |
| --- | --- |
| `pnpm lint` | exit 0 |
| `pnpm exec next typegen` | exit 0 |
| `pnpm exec tsc --noEmit` | exit 0 |
| `pnpm test` | exit 0: **110 arquivos, 1794 testes, 1794 passaram**, 0 falharam (iteração 3: 1787; +7 acréscimos) |
| `pnpm build` | exit 0 ("Compiled successfully") |
| `pnpm lint:api` | exit 0 |
| `pnpm gen:api:check` | exit 0 |
| `pnpm audit --audit-level high` | exit 0 (1 high ignorada, `GHSA-vfj7-8cjw-p6xm`, a mesma da iteração 3) |
| `git diff develop -- api web/package.json web/pnpm-lock.yaml web/openapi* api/openapi` | vazio |
| `git diff 1056707..HEAD --name-only` fora de `.specs/` | só os 4 arquivos do espécime e os 2 testes dele, `shell-frame.tsx` e o teste, `comprovantes-panel.tsx` e o teste, os testes de produtos e do detalhe do lançamento, e `global.guard.test.ts` |

- Nada mudou em `web/lib`, `features/*/hooks.ts`, permissões, regras de negócio ou payloads.
- **Ambiente, não produto**: o build regravou `web/.next`, de onde o processo antigo de `:3000` serve.
  - Esse processo já servia um `.next` recompilado depois do seu início (BUILD_ID de 19:57; processo iniciado às 14:49).
  - O navegador registra `ChunkLoadError` em `:3000`.
  - O processo não foi tocado.

### 8. Navegador

- Edge headless por CDP, com perfil temporário, apagado no fim.
- Login pelo formulário em `:3000`, na conta do presidente. A senha foi lida do arquivo local e não foi impressa nem gravada.
- O cookie `tj_session` (`localhost`, `httpOnly`) valeu em `:3142` e `:3143`.
- As medidas de D1, D3, D4 e de `/design-system` estão nos itens acima.
- Console em `:3142`: nenhum erro nas telas medidas, além do `404` esperado de `/design-system`.
- `:3142`, `:3143` e o Edge foram encerrados; `:3000` continua rodando.

### Revisão `web-design-guidelines` (restrita aos 8 arquivos alterados desde a iteração 3)

Arquivos revisados contra as regras de formulários, foco e hover:
- `example-login-form.tsx`
- `example-listing.tsx`
- `fields-section.tsx`
- `overlays-section.tsx`
- `shell-frame.tsx`
- `comprovantes-panel.tsx`
- `produtos-page.tsx`
- `lancamento-detail.tsx`

```text
Nenhum achado novo. F1 (D2), F2 (D3) e F3 (D4) corrigidos; O1 com teste (D1).
```

Notas, sem achado:
- O hover da marca deixa o fundo menos escuro e reduz o contraste do branco em relação ao repouso. É o mesmo padrão já aprovado nos itens do shell.
- O `Input` "Código" do espécime é `readOnly`, sem edição, e por isso fica fora da regra de correção ortográfica.

### Situação final por AC

`evidence/matriz-criterios.md`, conferida e corrigida só na formatação:

| Situação | ACs |
| --- | --- |
| comprovado | 207 |
| atendido por decisão documentada (WDS-D-034) | 12 |
| pendente | 0 |
| reprovado | 0 |
| **Total** | **219** |

| Requirement | Status |
| --- | --- |
| DSF-01 a DSF-06 | ✅ Verified (WDS-D-034: DSF-03 AC4, DSF-04 AC8) |
| DSC-01 a DSC-07 | ✅ Verified (WDS-D-034: DSC-01 AC1, DSC-04 AC5, DSC-05 AC2 e AC5); DSC-01 AC2 agora tem teste e os mutantes morrem |
| DSP-01 a DSP-07 | ✅ Verified (WDS-D-034: DSP-01 AC3, DSP-02 AC6, DSP-04 AC2, DSP-07 AC1 a AC3) |
| DSS-01 a DSS-06 | ✅ Verified; DSS-04 AC1, AC4 e AC5 reproduzidos com exceção real |
| DSM-01 a DSM-05 | ✅ Verified |
| DSM-06 | ✅ Verified (AC2: D2 a D4 corrigidos; revisão restrita sem achado) |

### Sensor (iteração 4)

- **Resultado**: 14 mutações injetadas (M28, M29, M28r, M29o, G1 a G6, H1, H2); **14 mortas**, 0 sobreviventes.
- **Somado à iteração 3**: lá morreram 29 (M01 a M27 e variantes). M28 e M29, que lá sobreviveram, agora morrem. Nenhum sobrevivente em aberto.
- **Isolamento**: worktree descartável.
  - As reversões regravaram os arquivos com LF no lugar de CRLF.
  - Conferi com `git diff --ignore-cr-at-eol` (vazio) e restaurei com `git checkout -- .`.
  - O status ficou vazio antes de remover o worktree.

### Pendências e exceções que ficam

- **12 lacunas de precisão das specs** aceitas em WDS-D-034: entram na próxima revisão das specs.
- **`global-error`** (DSS-04 AC2): sem simulação no navegador; coberto por teste.
- **D5** (toast com `role="dialog"`) e os conflitos registrados em `STATE.md`: sem mudança.
- **Execução**: nenhuma pendência. Push, PR e merge continuam não autorizados.

### Fechamento determinístico (iteração 4)

`python .claude/skills/tlc-spec-driven/scripts/validate_state.py web-design --root C:/Users/Michels/Desktop/torcida-jovem` → exit 0, "validate_state: 0 error(s) across [web-design]".
