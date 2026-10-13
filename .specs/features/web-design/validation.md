# Validation — web-design: FAIL ❌ (T14, iteração 2 de 3)

**Veredito: FAIL ❌**

**Date**: 2026-10-13
**Spec**: `.specs/features/web-design/spec/01` a `05` (DSF, DSC, DSP, DSS, DSM); Web V1 em `.specs/features/web/spec/01` a `06`
**Diff range**: `7c30c6a..7499d87` (base documental `7c30c6a` sobre a `develop` em `cebbdf2`; HEAD `7499d87cb1c4b7a83cefb51bc81c48f3823fe9a7`). Correções desta iteração: `87fdd3a`, `63b5e8e`, `ea87164`, integradas em `d050091`; decisões em `7499d87` (WDS-D-034)
**Verifier**: sub-agente independente (autor ≠ verificador; as correções foram feitas por outro agente). Evidência por `arquivo:linha`, saída de comando ou medição no navegador.

Motivo do FAIL, em uma linha: todas as lacunas apontadas na iteração 1 estão fechadas e verificadas (inclusive por mutante), mas a revisão refeita do zero achou três pontos que já existiam em `f9de714` e que **a iteração 1 deixou passar** (falha do verificador, não regressão das correções): `NativeSelect` do formulário de lançamento sem `autoComplete` (DSP-04 AC6), quantidade e unidade separadas por espaço comum (DSP-07 AC4) e `QuantityField` do ajuste sem `name` (regra da `web-design-guidelines`). Nenhum tem correção nem conflito registrado.

## Histórico

- **Iteração 1** (`7c30c6a..f9de714`, 2026-10-13): **FAIL**. Gates verdes (110 arquivos, 1774 testes); 203/219 ACs ✅, 3 ❌ (DSP-04 AC6, DSP-04 AC7, DSM-06 AC2), 13 ⚠️; sensor 16 injetadas, 15 mortas, M13 sobreviveu; 22 achados da `web-design-guidelines` sem correção nem conflito; desvios de propriedade (`lib/forms`) e de escopo (`.gitignore`). Relatório completo no histórico do git (`0da56c2`).
- **Iteração 2** (este relatório): lacunas da iteração 1 fechadas; três achados novos, listados acima.

---

## Task Completion

| Task | Status | Notes |
| --- | --- | --- |
| Rodadas 1 a 6, T11, T12, T13 | ✅ | `STATE.md:267-282`; `evidence/` |
| Correções da iteração 1 | ✅ | `87fdd3a`, `63b5e8e`, `ea87164`, `d050091`, `7499d87` |
| T14 | ❌ | três achados novos sem correção nem conflito |

---

## Gate Check (repositório principal, `web/`, `API_URL=http://localhost:8080`)

| Gate | Saída |
| --- | --- |
| `pnpm lint` | exit 0 |
| `pnpm exec next typegen` | exit 0 |
| `pnpm exec tsc --noEmit` | exit 0 |
| `pnpm test` | exit 0 — **110 arquivos, 1785 testes, 1785 passaram, 0 falharam, 0 pulados** (+11 sobre a iteração 1) |
| `pnpm build` | exit 0 (Next 16.4.0, "Compiled successfully"). Rodado no worktree temporário em `7499d87`, para não sobrescrever o `.next` servido em `:3000` |
| `pnpm lint:api` | exit 0 |
| `pnpm gen:api:check` | exit 0 |
| `pnpm audit --audit-level high` | exit 0 — 1 high ignorada (`GHSA-vfj7-8cjw-p6xm`, `web/pnpm-workspace.yaml:10-12`, inalterado desde a `develop`) |
| `git diff --name-only develop -- api/ web/package.json web/pnpm-lock.yaml` | vazio |

- Testes antes do ciclo: 69 arquivos / 851 testes (`STATE.md:271`). Agora: 110 / 1785. Nenhum arquivo de teste removido (lista de `7c30c6a` contida na de HEAD).

---

## Conferência das correções da iteração 1

| Lacuna da iteração 1 | Correção | Evidência | Mutante |
| --- | --- | --- | --- |
| DSP-04 AC7: código sem `spellCheck` | `create-produto-dialog.tsx:114-119` | `global.guard.test.ts:221` | M17 morto |
| DSP-04 AC6: motivos e filtros sem `name`/`autocomplete` | `confirm-dialog.tsx:151-152`, `ajuste-dialog.tsx:108-109`, `admin-dialogs.tsx:141-142`, `temporary-password-dialog.tsx:118-119`, `produtos-page.tsx:93-99`, `lancamentos-page.tsx:209-273`, `users-page.tsx:156-177` | `global.guard.test.ts:213`; `confirm-dialog.test.tsx:76` | M18 morto |
| `focus-within` do campo de dinheiro | `numeric-text-field.tsx:77-94` | `numeric-text-field.test.tsx:72`; navegador (recorte abaixo) | M19 morto |
| `tap-highlight` e `touch-action` | `app/globals.css:268-281` | `tokens.test.ts:378` | M21 morto |
| Rótulo longo do botão | `button.tsx:12,26-29` (sem `whitespace-nowrap`, `min-h-*`); `icon-button.tsx:33` mantém `whitespace-nowrap` | `actions.test.tsx:98,231`; navegador (320 e 390) | M20 morto |
| Prévia "Saldo resultante" sem `aria-live` | `ajuste-dialog.tsx:123` | `ajuste-dialog.test.tsx:212` | M22 morto |
| Placeholders sem "…" | espécime | `global.guard.test.ts:232` | — |
| Mutante M13 | testes | `lancamentos-page.test.tsx:476`, `produtos-page.test.tsx:366`, `users-page.test.tsx:185` | M13, M13p mortos |
| `lib/forms` fora da propriedade | registrado | `STATE.md:152` (WDS-D-034) | — |
| DSM-01 AC2 e `.gitignore` | AC revisado | `spec/05-migracao-web-v1.md:67` | — |
| Lacunas de precisão da spec | aceitas | `STATE.md:154` (WDS-D-034) | — |

---

## Spec-Anchored Acceptance Criteria

Legenda: ✅ o teste/evidência afirma o valor que a spec define · ❌ lacuna · ⚠️ lacuna de precisão da spec (aqui todas aceitas e registradas em WDS-D-034, `STATE.md:154`). Linhas sem mudança desde a iteração 1 foram reconferidas contra `7499d87` (os arquivos citados não mudaram, salvo os listados na tabela acima).

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
| 2 | foco ≥ 2 px, ≥ 3:1 no claro e no shell | `actions.test.tsx:117`; `shell-frame.test.tsx:102`; `contrast.test.ts:55`; campo de dinheiro: `numeric-text-field.test.tsx:72` e navegador (anel preto de 2 px com afastamento de 2 px no grupo, `box-shadow … 0 0 0 2px #fff, 0 0 0 4px #111`). Observação: o link do nome do produto (`features/estoque/produtos/produtos-page.tsx:138`) e o da devolução (`lancamento-detail.tsx:160`) usam o contorno padrão do navegador (`outline: auto`, cinza a 50%), não o anel do kit | ✅ (observação O1) |
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
| 6 | todo campo declara `autocomplete` e tipo/modo | Corrigidos e guardados: `test/design/global.guard.test.ts:213` (motivos, filtros e busca com `name` e `autoComplete="off"`), `confirm-dialog.test.tsx:76`; sensor M18 morto. **Restam** três `NativeSelect` do formulário de lançamento com `name` (via `field.name`) e sem `autoComplete`: `features/financeiro/lancamentos/lancamento-form.tsx:186` (Tipo), `:220` (Conta), `:271` (Forma de pagamento). A guarda só cobre a lista fixa `NAMED_CONTROLS` (`global.guard.test.ts:92-100`), por isso não os vê | ❌ |
| 7 | e-mail e código sem correção ortográfica | `create-produto-dialog.tsx:116` (`spellCheck={false}`, `autoCapitalize="none"`); `global.guard.test.ts:221` (todo campo `type="email"` ou `register("email"|"codigo")` com `spellCheck={false}`); `access-screens.test.tsx:55,73`; sensor M17 morto | ✅ |
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
| 4 | "…" e espaço não separável entre número e unidade | "…": `global.guard.test.ts:232` (placeholders), `actions.test.tsx:178`. Espaço não separável só no limite do anexo (`comprovantes-panel.test.tsx:373`, `features/financeiro/comprovantes/limits.ts`). **Quantidade e unidade com espaço comum** no texto visível: `features/estoque/produto-detail.tsx:107` (saldo "3 UN"; no navegador `/ /` falso), `features/estoque/ajustes/ajuste-dialog.tsx:121` (saldo atual) e `:126` (saldo resultante); o teste `features/estoque/produto-detail.test.tsx:50` afirma `"12 UN"` com espaço comum. Valores em reais ficam em `whitespace-nowrap` (`money-value.tsx`), o que equivale | ❌ |
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
| 2 | `web-design-guidelines` com cada achado corrigido ou registrado como conflito | Os 22 achados da iteração 1 estão corrigidos (seção da revisão). **Restam achados novos sem correção nem conflito** (A1 a A3) | ❌ |
| 3 | navegador real, produção, 5 larguras, todas as telas | `evidence/smoke-visual.md:66-85,152-181` | ✅ |
| 4 | percurso só por teclado | `evidence/smoke-visual.md:87-97`; N2 e N3 corrigidos em `2903f32` | ✅ |
| 5 | 200% e 320 px | `evidence/smoke-visual.md:101-117` | ✅ |
| 6 | verificador independente com `validation.md` | este arquivo | ✅ |
| 7 | divergência vira correção | N1, N2, N3 corrigidos (`evidence/smoke-visual.md:229-231`) | ✅ |


**Status**: 219 ACs — **204 ✅, 3 ❌ (DSP-04 AC6, DSP-07 AC4, DSM-06 AC2), 12 ⚠️ aceitas em WDS-D-034** (DSF-03 AC4, DSF-04 AC8, DSC-01 AC1, DSC-04 AC5, DSC-05 AC2, DSC-05 AC5, DSP-01 AC3, DSP-02 AC6, DSP-04 AC2, DSP-07 AC1, AC2, AC3). Saíram de ❌: DSP-04 AC7; de ⚠️: DSM-01 AC2 e a borda de 50 de DSP-05. Entrou em ❌: DSP-07 AC4 (na iteração 1 marquei ✅ olhando só o limite do anexo; o AC vale para todo número com unidade).

---

## Regressão da Web V1

- Os 69 arquivos de teste da base passam (gate `pnpm test`, 1785/1785); nenhum removido.
- Arquivos da base alterados depois de `f9de714`: só acréscimos (`confirm-dialog.test.tsx`, `ajuste-dialog.test.tsx`, `produtos-page.test.tsx`, `lancamentos-page.test.tsx`, `users-page.test.tsx`) e, em `components/ui/actions.test.tsx` (arquivo do ciclo), a troca `h-control-*` → `min-h-control-*` nos tamanhos, motivada por DSM-06 AC2 (rótulo longo) e citada no próprio teste (`actions.test.tsx:96-97`); a altura mínima segue o mesmo token, e no navegador os botões continuam com 32, 36 e 44 px.
- O julgamento da iteração 1 sobre os 15 arquivos da base alterados continua válido: nenhuma dessas linhas mudou de novo. Os casos removidos na remoção de alias seguem cobertos (`lancamentos-page.test.tsx:422,431,435`; `status-badge.test.tsx:15-21,76`).

---

## Discrimination Sensor

Worktree temporário em `7499d87 --detach`, `pnpm install --frozen-lockfile`; uma falha por vez, `pnpm exec vitest run` nos testes relevantes, `git checkout -- .` entre uma e outra (status do worktree vazio depois de cada reversão, registrado no log).

| # | Falha | Mutação | Resultado |
| --- | --- | --- | --- |
| M01 | token quebra contraste | `app/globals.css:178` `--muted-foreground` → `--tj-neutral-400` | **MORTO** (5) |
| M02 | cor literal numa tela | `lancamentos-page.tsx` + `text-[#D81E28]` | **MORTO** (3) |
| M03 | traçado do escudo | um dígito do primeiro `d=` do SVG | **MORTO** (3) |
| M04 | IconButton sem nome | `icon-button.tsx:32` sem `aria-label` | **MORTO** (35) |
| M05 | `rowLabel` removido da linha | `lancamentos-page.tsx` `rowLabel=""` | **MORTO** (3) |
| M05b | idem no componente | `row-actions.tsx:60` `name = label` | **MORTO** (32) |
| M06 | Field sem `aria-invalid`/`describedby` | `field.tsx` força `invalid={false}` | **MORTO** (6) |
| M07 | `useQueryState` aceita inválido | `codecs.ts:34` | **MORTO** (5) |
| M08 | parâmetro não limpo no padrão | `use-query-state.ts:69-70` | **MORTO** (8) |
| M09 | shell em `md` | `shell-frame.tsx` `lg:` → `md:` | **MORTO** (3) |
| M10 | link de pular removido | `shell-frame.tsx:109` | **MORTO** (7) |
| M11a | MoneyValue sem sinal | `money-value.tsx:33` | **MORTO** (3) |
| M11b | MoneyValue sem ícone | `money-value.tsx:50` | **MORTO** (2) |
| M12 | espécime em produção | `dev-only.ts:8` | **MORTO** (5) |
| M13 | "Mostrar mais" com exatamente 50 (lançamentos) | `lancamentos-page.tsx` `>` → `>=` | **MORTO** (1: "com exatamente 50, o controle não aparece") |
| M13p | idem em produtos | `produtos-page.tsx:160` `>` → `>=` | **MORTO** (1) |
| M13b | idem no componente | `show-more.tsx:33` `<` → `<=` | **MORTO** (2) |
| M14 | status com rótulo e ícone iguais | `labels.ts:45` | **MORTO** (5) |
| M17 | código com correção ortográfica | `create-produto-dialog.tsx:116` sem `spellCheck` | **MORTO** (1) |
| M18 | motivo sem `name` | `confirm-dialog.tsx:151` | **MORTO** (2) |
| M19 | sem `focus-within` no campo de dinheiro | `numeric-text-field.tsx:80` | **MORTO** (2) |
| M20 | botão volta a `whitespace-nowrap` | `button.tsx:12` | **MORTO** (3) |
| M21 | sem `tap-highlight` | `globals.css:269` | **MORTO** (1) |
| M22 | prévia do ajuste sem `aria-live` | `ajuste-dialog.tsx:123` | **MORTO** (1) |

**Resultado do sensor**: 23 injetadas (as 16 da iteração 1, M13p e seis sobre as correções), **23 mortas, 0 sobreviventes** → PASS.

**Isolamento**: `git status --porcelain` do principal antes: vazio (0 bytes); depois de `git worktree remove --force` (diretório apagado por `Remove-Item -LiteralPath \\?\…`, caminho longo do `node_modules`) e `git worktree prune`: vazio (0 bytes); `git worktree list` sem o temporário; HEAD `7499d87`.

---

## Revisão `web-design-guidelines` (refeita do zero)

Aplicada a `web/app`, `web/components` e `web/features` (arquivos não-teste), regra a regra. Além das buscas da iteração 1, desta vez: todos os controles de formulário (`Input`, `Textarea`, `NativeSelect`, `MoneyField`, `QuantityField`, nativos) quanto a `name`/`autocomplete`; todo `Link`/`a`/`button` sem variante do kit quanto a `hover:` e foco; números com unidade; aspas retas em texto de interface; `truncate` em flex.

### Achados sem correção e sem conflito registrado (reprovam)

```text
## features/financeiro/lancamentos/lancamento-form.tsx
features/financeiro/lancamentos/lancamento-form.tsx:186 - NativeSelect "Tipo" sem autoComplete (DSP-04 AC6)
features/financeiro/lancamentos/lancamento-form.tsx:220 - NativeSelect "Conta" sem autoComplete (DSP-04 AC6)
features/financeiro/lancamentos/lancamento-form.tsx:271 - NativeSelect "Forma de pagamento" sem autoComplete (DSP-04 AC6)

## features/estoque/produto-detail.tsx
features/estoque/produto-detail.tsx:107 - "3 UN": espaço comum entre número e unidade → espaço não separável (DSP-07 AC4)

## features/estoque/ajustes/ajuste-dialog.tsx
features/estoque/ajustes/ajuste-dialog.tsx:95 - QuantityField "Quantidade" sem name (o NumericTextField já põe autoComplete="off")
features/estoque/ajustes/ajuste-dialog.tsx:121 - saldo atual com espaço comum entre número e unidade (DSP-07 AC4)
features/estoque/ajustes/ajuste-dialog.tsx:126 - saldo resultante com espaço comum entre número e unidade (DSP-07 AC4)
```

Agrupados no relatório como **A1** (selects do formulário de lançamento), **A2** (número e unidade) e **A3** (`name` da quantidade do ajuste). Os três já existiam em `f9de714`; a iteração 1 não os apontou.

### Achados da iteração 1: todos corrigidos

Os 22 da iteração 1 (código, motivos, busca, filtros, `focus-within`, `tap-highlight`, rótulo longo, `aria-live` da prévia, placeholders do espécime) estão corrigidos; ver "Conferência das correções".

### Achados casados com conflito ou decisão registrada (não reprovam)

| Achado | Registro |
| --- | --- |
| Title Case | `STATE.md:193` |
| Envio desabilitado sem motivo | `STATE.md:194` |
| Estado na URL sem `nuqs` | `STATE.md:195` |
| Listas > 50 sem virtualização | `STATE.md:196` |
| Dark mode | `STATE.md:197` |
| `safe-area` sem `viewportFit: cover` | `STATE.md:286-292` (conflitos da rodada 4) |
| Placeholder `0,00` do campo de dinheiro (por propriedade, fora da guarda de literais) | idem |
| Toast anima `height` | idem |
| Data de calendário sem `Intl` | idem |
| Anel do item de menu também no hover | idem |
| Toast `role="dialog"` (D5) e toast de erro `polite` | `STATE.md:299` e `STATE.md:304` |
| Sem aviso de alterações não salvas | `STATE.md:262` ("Fora deste ciclo") |
| `lib/forms/use-focus-invalid-after-submit.ts` | `STATE.md:152` |

### Observações (não reprovam)

- **O1** `features/estoque/produtos/produtos-page.tsx:138` e `features/financeiro/lancamentos/lancamento-detail.tsx:160`: links com o contorno padrão do navegador no foco (medido: `outline: auto 1px` cinza a 50%, visível no recorte `1440-produto-link-foco.png`), em vez do anel do kit. Atende "foco visível ou equivalente"; para o padrão de DSC-01 AC2, vale passar ao anel de 2 px do token `ring`.
- **O2** `app/(dev)/design-system/acesso/example-login-form.tsx:52`: o placeholder ficou "pessoa@exemplo.com…", que cumpre a regra mas lê como e-mail incompleto. Só no espécime.
- **O3** A guarda de `name`/`autoComplete` (`global.guard.test.ts:92-100,213`) é uma lista fixa de arquivos; um campo novo fora dela (como A1 e A3) não é apontado.

### Conferido e sem achado

`outline-none` só em alvos de foco programático com `tabIndex={-1}`; nenhum `onClick` em `div`/`span`; imagem com dimensões e prioridade; nenhum `autoFocus`, `user-scalable`, `maximum-scale` ou `onPaste`; nenhum `...` literal; nenhuma aspa reta em texto exibido; `touch-action: manipulation` e `-webkit-tap-highlight-color` na camada base; links e cartões com `hover:` (o link do cartão de início herda o `hover:` do `Card`, `atalhos.tsx:74`); rádio nativo do plano de contas com `name` por `register`; `input type="file"` fora da regra de `autocomplete`; valores em reais com `whitespace-nowrap`.

---

## SSR e hidratação

Sem mudança de padrão nas correções (nenhum Client Component novo; `useSearchParams` continua sob `Suspense` nas quatro páginas; `useMediaQuery` com `useSyncExternalStore` e snapshot de servidor fixo; nenhum `window`, `Date.now`, `Math.random` ou `new Date()` em renderização). No navegador: 0 avisos de hidratação, 0 exceções, 0 avisos de console.

---

## Propriedade e escopo

- Fora de `web/`: só `.specs/` (`features/web-design/`, `LESSONS.md`, `lessons.json`), `docs/development/web.md` e `.gitignore`, como permite DSM-01 AC2 revisado.
- `api/`, `web/package.json`, `web/pnpm-lock.yaml`, `web/lib/api/`, `web/lib/session/session-provider.tsx`, `features/*/hooks.ts`: sem diff. `RequirePermission`: só a aparência de `NoAccess` (já conferido na iteração 1; sem mudança depois).
- As correções desta iteração não tocam hooks, clientes, catálogos de erro, permissões nem regras de negócio: só atributos de campo, classes, CSS da camada base e testes.
- `lib/forms/use-focus-invalid-after-submit.ts`: registrado em `STATE.md:152`.

---

## Smoke (amostra do verificador, build de produção `7499d87` em `:3000`, API em `:8080`)

Edge headless por CDP, scripts em `…/scratchpad/verify2/smoke/`, perfil temporário encerrado. Senha só em memória. Capturas no scratchpad, fora do repositório.

| Conferência | Resultado |
| --- | --- |
| Login presidente | `/inicio` |
| Lançamentos, usuários e produtos em 390, 1024 e 1440 | `scrollWidth === clientWidth` nas 9; nenhum botão cortado (`scrollWidth > clientWidth`) nem fora do viewport; alturas de botão 32, 36 e 44 px, as mesmas dos tokens |
| Comparação com `evidence/smoke-final/` | Usuários 390, 1024, 1440 e produtos 390, 1440: **0,00%** de pixels diferentes. Lançamentos: 9,39% (390), 5,98% (1024), 3,37% (1440), todos de dados: a lista tem 9 lançamentos contra 4 na captura de base (os criados nos smokes), e em 1024 a barra de rolagem vertical desloca o conteúdo 15 px; inspecionadas lado a lado, cabeçalho, barra de filtros, colunas, selos e botões "Detalhes" iguais. Produtos 1024 sem captura de base |
| Rótulo longo em 320 e 390 | botão real da página ("Novo produto", classes do build) clonado com "Gerar senha temporária para a pessoa selecionada agora": 288 × 54 px em 320 e 358 × 54 px em 390 (duas linhas), `white-space: normal`, nada cortado, dentro do viewport, sem rolagem do documento; o original segue com 36 px |
| Anel do campo de dinheiro por teclado | "Novo lançamento" → Tab até "Valor bruto": campo `:focus-visible`, grupo `[data-slot=field-group]` com `box-shadow 0 0 0 2px #fff, 0 0 0 4px #111` (anel de 2 px afastado 2 px) envolvendo prefixo "R$" e campo; o campo não repete o anel. Recorte `1440-dinheiro-foco-recorte.png` |
| Prévia do ajuste | "Saldo resultante" com `aria-live="polite"`; saldo do produto "3 UN" sem espaço não separável (A2) |
| Foco do link de produto | `:focus-visible` com contorno padrão do navegador (O1) |
| Console | 0 hidratação, 0 exceções, 0 avisos; só o 401 de `/auth/me` antes do login |

---

## Code Quality

| Principle | Status |
| --- | --- |
| Minimum code / surgical changes | ✅ correções restritas a atributos, classes e CSS base |
| No scope creep | ✅ |
| Matches patterns | ✅ |
| Spec-anchored outcome check | ❌ DSP-04 AC6, DSP-07 AC4, DSM-06 AC2 |
| Per-layer coverage | ⚠️ guarda de atributos por lista fixa (O3) |
| Every test maps to a requirement | ✅ |
| Documented guidelines followed | ❌ três achados da `web-design-guidelines` |

---

## Lacunas ordenadas (com unidade dona)

1. **A1, DSP-04 AC6** — `NativeSelect` "Tipo", "Conta" e "Forma de pagamento" sem `autoComplete` em `features/financeiro/lancamentos/lancamento-form.tsx:186,220,271`. Dono: DS-FB. Correção: `autoComplete="off"`; de preferência, tornar a guarda de `global.guard.test.ts:213` geral (todo `Input`/`Textarea`/`NativeSelect` não de autenticação com `autoComplete`), em vez da lista fixa (O3).
2. **A2, DSP-07 AC4** — número e unidade com espaço comum em `features/estoque/produto-detail.tsx:107` e `features/estoque/ajustes/ajuste-dialog.tsx:121,126`. Dono: DS-ES. Correção: espaço não separável (`\u00a0`), com o teste `features/estoque/produto-detail.test.tsx:50` e os do ajuste atualizados citando DSP-07 AC4.
3. **A3, `web-design-guidelines`** (`name` com significado) — `QuantityField` do ajuste sem `name` em `features/estoque/ajustes/ajuste-dialog.tsx:95`. Dono: DS-ES. Correção: `name="quantidade"`.
4. **O1** (não bloqueia) — links de produto e de devolução com o contorno padrão do navegador; alinhar ao anel do kit. Donos: DS-ES e DS-FB.

---

## Requirement Traceability Update

| Requirement | New Status |
| --- | --- |
| DSF-01 a DSF-06 | ✅ Verified (⚠️ aceitas: DSF-03 AC4, DSF-04 AC8) |
| DSC-01 a DSC-07 | ✅ Verified (⚠️ aceitas: DSC-01 AC1, DSC-04 AC5, DSC-05 AC2/AC5) |
| DSP-01, DSP-02, DSP-03, DSP-05, DSP-06 | ✅ Verified (⚠️ aceitas indicadas) |
| DSP-04 | ❌ Needs Fix (AC6: A1) |
| DSP-07 | ❌ Needs Fix (AC4: A2) |
| DSS-01 a DSS-06 | ✅ Verified (FND-05 AC2 revisado reconfirmado) |
| DSM-01 a DSM-05 | ✅ Verified |
| DSM-06 | ❌ Needs Fix (AC2: A1 a A3) |

---

## Summary

**Overall**: ❌ Not Ready (iteração 2 de 3)

**Spec-anchored check**: 204/219 ✅; 3 ❌; 12 lacunas de precisão aceitas em WDS-D-034
**Sensor**: 23 injetadas, 23 mortas, 0 sobreviventes
**Gate**: todos com saída 0 (1785/1785 testes)

**What works**: todas as lacunas da iteração 1 fechadas e defendidas por teste e por mutante; rótulo longo quebra sem estourar em 320 e 390; anel do campo de dinheiro no grupo por teclado; nenhuma regressão visual nas três telas nas três larguras (diferenças só de dados); escopo e propriedade conforme o AC revisado.

**Next steps**: corrigir A1 a A3 (mudança pequena, em dois arquivos de tela e um teste) e nova verificação (iteração 3 de 3, a última antes de escalar ao mantenedor).

---

## Fechamento determinístico

`python .claude/skills/tlc-spec-driven/scripts/validate_state.py web-design --root C:/Users/Michels/Desktop/torcida-jovem` → exit 1: "web-design: validation.md verdict is FAIL - route the ranked gaps to fix tasks, then re-verify (feature is not done)". Coerente com o veredito: a feature não está concluída. Observação: o script lê o veredito do título `# Validation` e das linhas de resultado com dois-pontos; o título traz o veredito e a linha do sensor foi rotulada "Resultado do sensor" para não ser lida como veredito.
