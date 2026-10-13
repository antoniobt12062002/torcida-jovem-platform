# Matriz consolidada dos critérios — web-design

- **Data**: 2026-10-13
- **Ponto conferido**: `023558b` (branch `ds/v3-fix`, sobre `feature/web-design` em `1056707`)
- **Base**: tabela por AC de `validation.md` (iteração 3, sobre `13e56c5`), mais as correções desta rodada (WDS-D-035): `13f5ad7` (D1), `9e95fb9` (D3), `131a5c4` (D2), `023558b` (D4) e a evidência da T2 da DS-S (`evidence/erro-no-shell.md`)
- **Autor**: agente de correção (não é o verificador). Esta matriz é entrada para a 4ª verificação independente, que confere cada linha.

## Situações

Cada critério tem exatamente uma situação:

- **comprovado**: um teste ou uma evidência que passou, com referência a arquivo e linha, saída de comando ou medição no navegador.
- **atendido por decisão documentada**: lacuna de precisão da spec aceita pelo mantenedor em WDS-D-034 (`STATE.md:154`). Não é teste que passou; fica como entrada para a próxima revisão das specs.
- **pendente**: sem evidência ainda.
- **reprovado**: a evidência contradiz o critério.

## Contagem

| Situação | Critérios |
| --- | --- |
| comprovado | 207 |
| atendido por decisão documentada | 12 |
| pendente | 0 |
| reprovado | 0 |
| **Total** | **219** |

Os 12 atendidos por decisão documentada são os mesmos 12 ⚠️ do `validation.md`: DSF-03 AC4, DSF-04 AC8, DSC-01 AC1, DSC-04 AC5, DSC-05 AC2, DSC-05 AC5, DSP-01 AC3, DSP-02 AC6, DSP-04 AC2, DSP-07 AC1, DSP-07 AC2 e DSP-07 AC3.

## Reconciliação com o `validation.md`

- **219 = só os ACs numerados** das specs `spec/01` a `05`: DSF 47 (7 + 10 + 8 + 8 + 8 + 6), DSC 46 (7 + 5 + 9 + 9 + 8 + 5 + 3), DSP 53 (6 + 6 + 7 + 9 + 14 + 4 + 7), DSS 35 (9 + 7 + 6 + 5 + 4 + 4), DSM 38 (6 + 5 + 8 + 7 + 5 + 7). A contagem foi feita sobre os itens numerados de cada bloco **Acceptance Criteria** das specs e bate com o total do `validation.md`.
- **Casos de borda não entram nos 219.** As seções "Edge Cases" das cinco specs não são numeradas como AC. O `validation.md` tem uma linha extra, "Borda: exatamente 50" em DSP-05, que não conta no total (204 + 3 + 12 na iteração 2 e 206 + 1 + 12 na iteração 3 somam 219 sem ela). Ela segue comprovada (`lancamentos-page.test.tsx:476`, `produtos-page.test.tsx:385`, `users-page.test.tsx:185`; mutantes M13, M13p e M13b mortos) e fica fora da soma desta matriz.
- **Linhas agrupadas do `validation.md` foram separadas**: "DSF-02 AC1-3" virou três linhas (AC1, AC2, AC3), com as mesmas referências.
- **Iteração 3 → agora**: 206 ✅ + 1 ❌ + 12 ⚠️ passa a 207 comprovados + 12 por decisão. O ❌ (DSM-06 AC2) passou a comprovado com as correções de D2 a D4. DSC-01 AC2 estava contado entre os ✅ do `validation.md` com a ressalva de que o teste não discriminava (M28 e M29); agora tem teste e mutantes mortos.
- **Referências de linha**: remapeadas para `023558b` nos arquivos que esta rodada mudou (`global.guard.test.ts` +27 a partir da linha 240; `shell-frame.test.tsx` +19 a partir da 180; `produtos-page.test.tsx` +19 a partir da 334; `app/(dev)/design-system/page.test.tsx` +18 a partir da 72). As demais são as do `validation.md`, em arquivos que não mudaram.

## O que mudou desde a iteração 3

| Item | Correção | Critérios afetados | Prova |
| --- | --- | --- | --- |
| D1 (M28, M29) | `13f5ad7`: testes que exigem o anel do kit nos links de produto e de devolução | DSC-01 AC2, AC3 | os dois testes falham com o mutante e passam sem ele; navegador: anel `#111111` de 2 px com afastamento branco de 2 px (`evidence/foco-e-hover.md`) |
| D2 (F1) | `131a5c4`: `name` e `autoComplete` nos 13 campos do espécime; login de exemplo com `username` e `current-password`; guarda geral para `Input` e `MoneyField` | DSP-04 AC6, DSM-06 AC2 | a guarda falhava com 14 ocorrências (13 do espécime e a de D3) e passa; mutantes G1 a G5 da guarda mortos; testes de página do espécime |
| D3 (F2) | `9e95fb9`: `name="comprovante"` no campo de arquivo | DSP-04 AC6, DSM-06 AC2 | teste novo falha antes e passa depois; os 32 testes do painel (seleção, validação, limites, envio no campo `file`) passam sem mudança |
| D4 (F3) | `023558b`: `hover:bg-sidebar-accent` e a transição de cor dos itens no link da marca | DSS-01 AC2, DSM-06 AC2 | teste novo falha antes e passa depois; navegador: fundo `rgba(0,0,0,0)` → `rgb(42, 44, 46)` no hover, foco com anel branco do shell |
| T2 da DS-S | exceção real em build de produção, cliente e servidor, num worktree descartável | DSS-04 AC1, AC2, AC4, AC5 | `evidence/erro-no-shell.md` |
| Gates | `023558b` | DSM-06 AC1, DSM-01 AC1, DSC-07 AC2 | 110 arquivos, 1794/1794 testes; lint, typegen, tsc, build, `lint:api`, `gen:api:check` com saída 0 |
| Testes alterados | `evidence/testes-alterados.md` | DSM-01 AC4, DSC-07 AC3 | inclui as trocas de `fb62b99` e os acréscimos desta rodada |

## Resumo por requisito

| Requisito | ACs | comprovado | por decisão | pendente | reprovado |
| --- | --- | --- | --- | --- | --- |
| DSF-01 | 7 | 7 | 0 | 0 | 0 |
| DSF-02 | 10 | 10 | 0 | 0 | 0 |
| DSF-03 | 8 | 7 | 1 | 0 | 0 |
| DSF-04 | 8 | 7 | 1 | 0 | 0 |
| DSF-05 | 8 | 8 | 0 | 0 | 0 |
| DSF-06 | 6 | 6 | 0 | 0 | 0 |
| DSC-01 | 7 | 6 | 1 | 0 | 0 |
| DSC-02 | 5 | 5 | 0 | 0 | 0 |
| DSC-03 | 9 | 9 | 0 | 0 | 0 |
| DSC-04 | 9 | 8 | 1 | 0 | 0 |
| DSC-05 | 8 | 6 | 2 | 0 | 0 |
| DSC-06 | 5 | 5 | 0 | 0 | 0 |
| DSC-07 | 3 | 3 | 0 | 0 | 0 |
| DSP-01 | 6 | 5 | 1 | 0 | 0 |
| DSP-02 | 6 | 5 | 1 | 0 | 0 |
| DSP-03 | 7 | 7 | 0 | 0 | 0 |
| DSP-04 | 9 | 8 | 1 | 0 | 0 |
| DSP-05 | 14 | 14 | 0 | 0 | 0 |
| DSP-06 | 4 | 4 | 0 | 0 | 0 |
| DSP-07 | 7 | 4 | 3 | 0 | 0 |
| DSS-01 | 9 | 9 | 0 | 0 | 0 |
| DSS-02 | 7 | 7 | 0 | 0 | 0 |
| DSS-03 | 6 | 6 | 0 | 0 | 0 |
| DSS-04 | 5 | 5 | 0 | 0 | 0 |
| DSS-05 | 4 | 4 | 0 | 0 | 0 |
| DSS-06 | 4 | 4 | 0 | 0 | 0 |
| DSM-01 | 6 | 6 | 0 | 0 | 0 |
| DSM-02 | 5 | 5 | 0 | 0 | 0 |
| DSM-03 | 8 | 8 | 0 | 0 | 0 |
| DSM-04 | 7 | 7 | 0 | 0 | 0 |
| DSM-05 | 5 | 5 | 0 | 0 | 0 |
| DSM-06 | 7 | 7 | 0 | 0 | 0 |
| **Total** | **219** | **207** | **12** | **0** | **0** |

## Matriz por critério

### DSF-01 — Tokens em camadas

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | três camadas | `test/design/tokens.test.ts:100` | comprovado |
| 2 | `#111111 #D81E28 #FFFFFF #A7A9AC` | `tokens.test.ts:70` | comprovado |
| 3 | semânticos de fundo a informação, mais shell | `tokens.test.ts:135,163` | comprovado |
| 4 | componentes usam só semânticos/componente | `tokens.test.ts:117`; guarda `arbitrary-color` (`test/design/guard.ts:72-75`) | comprovado |
| 5 | um único bloco de tema | `tokens.test.ts:124` | comprovado |
| 6 | sem seletor de tema nem `prefers-color-scheme` | `tokens.test.ts:231,237,241` | comprovado |
| 7 | guarda falha com arquivo e linha | `test/design/guard.test.ts:61,75`; mutante M02 morto | comprovado |

### DSF-02 — Cor e contraste

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | texto ≥ 4,5:1 | `test/design/contrast.test.ts:30,34`; mutante M01 morto | comprovado |
| 2 | gráfico, borda de campo e foco ≥ 3:1 | `contrast.test.ts:42,55` | comprovado |
| 3 | o teste calcula cada par | `contrast.test.ts:30,34,42,55` (pares de `lib/design/contrast-pairs.ts`) | comprovado |
| 4 | ação primária em preto | `tokens.test.ts:216`; `components/ui/actions.test.tsx:52` | comprovado |
| 5 | vermelho só como acento | inspeção estática do `validation.md` (só o filete do `PageTitle` e o espécime); `components/app/shell-frame.test.tsx:287-289` | comprovado |
| 6 | vermelho nunca como fundo de página, cartão ou cabeçalho | idem AC5 | comprovado |
| 7 | erro diferente do vermelho de marca | `tokens.test.ts:216` | comprovado |
| 8 | aço não é texto nem borda de campo | `tokens.test.ts:222` | comprovado |
| 9 | vermelho não é texto < 24 px no shell | inspeção estática: nenhum `text-brand`/`text-sidebar-primary` | comprovado |
| 10 | verde e âmbar só para estado | inspeção estática: só em `alert`, `badge`, `toast`, `money-value` | comprovado |

### DSF-03 — Tipografia

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | Barlow 600/700, Geist e Geist Mono por `next/font` | `test/design/fonts.test.ts:26,35,43,63` | comprovado |
| 2 | corpo em Geist | `tokens.test.ts:248` | comprovado |
| 3 | Barlow só nos papéis listados | `test/design/global.guard.test.ts:30-36`; `components/ui/tokens.guard.test.ts:41` | comprovado |
| 4 | caixa alta só nos papéis listados | nome da marca em caixa alta (`components/ui/data.test.tsx:390-394`), aprovado no espécime (WDS-D-031). WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 5 | só a escala, mínimo 12 px | regra `arbitrary-text-size` (`guard.ts:77`) | comprovado |
| 6 | algarismos tabulares | `components/app/data-list.test.tsx:122`; `components/ui/data.test.tsx:55`; `components/app/money-value.test.tsx:29` | comprovado |
| 7 | campo ≥ 16 px em toque | `tokens.test.ts:294` | comprovado |
| 8 | `text-wrap: balance` nos títulos | `data.test.tsx:368,382` | comprovado |

### DSF-04 — Espaço, forma, elevação e movimento

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | escala de espaço | regra `arbitrary-spacing` (`guard.ts:79`) | comprovado |
| 2 | três raios; pílula só em status | `tokens.test.ts:337`; `rounded-full` só em `badge.tsx:12` | comprovado |
| 3 | sem sombra em repouso | `tokens.test.ts:344`; `data.test.tsx:183` | comprovado |
| 4 | sem gradiente nem brilho | `tokens.test.ts:392` | comprovado |
| 5 | durações dos tokens, sem `transition: all` | `tokens.test.ts:353`; regra `transition-all` | comprovado |
| 6 | redução de movimento | `tokens.test.ts:364`; `evidence/smoke-visual.md:134-140` | comprovado |
| 7 | três larguras nomeadas | `components/app/page.test.tsx:91,103`; `global.guard.test.ts:193` | comprovado |
| 8 | dois pontos de quebra, 768 e 1024 | `lib/responsive/use-media-query.test.tsx:41`; `tokens.test.ts:307`; o diálogo usa 640 px e o shell 400 px. WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |

### DSF-05 — Marca

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | traçados e cores idênticos, `viewBox` no círculo | `test/design/brand.test.ts:39,44,49,54`; mutante M03 morto | comprovado |
| 2 | 1:1, sem efeito | `components/app/page.test.tsx:40` | comprovado |
| 3 | ≥ 32 px e respiro | `page.test.tsx:47` | comprovado |
| 4 | `alt` vazio com nome; "Torcida Jovem" sozinho | `page.test.tsx:25,32` | comprovado |
| 5 | largura e altura declaradas | `page.test.tsx:40` | comprovado |
| 6 | sem substituto; dependência em `STATE.md` | `brand.test.ts:89`; `STATE.md:180-183` | comprovado |
| 7 | sem A FERA na área operacional | `brand.test.ts:89`; `components/app/access-layout.test.tsx:50` | comprovado |
| 8 | sem arquivos do `create-next-app` | `brand.test.ts:81,89,94` | comprovado |

### DSF-06 — Espécime

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | todas as seções | `app/(dev)/design-system/page.test.tsx:90,103`; `components.test.tsx:44-255` | comprovado |
| 2 | só componentes reais | `design-system.guard.test.ts:36` | comprovado |
| 3 | repouso, foco, desabilitado e inválido | `components.test.tsx:44,69` | comprovado |
| 4 | tabela nas três larguras | `components.test.tsx:124,137`; `evidence/especime.md` | comprovado |
| 5 | 404 em produção | `page.test.tsx:38,48`; `subpages.test.tsx:38`; mutante M12 morto; smoke T14 | comprovado |
| 6 | migração só depois da aprovação | `STATE.md:147-148` (WDS-D-033) antes da rodada 6 | comprovado |

### DSC-01 — Estados, foco e alvos

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | repouso, hover, pressionado, foco e desabilitado distintos | `actions.test.tsx:39,117,130`; espécime aprovado; sem valor verificável preciso na spec. WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 2 | foco ≥ 2 px e ≥ 3:1 no claro e no shell | `actions.test.tsx:117`; `shell-frame.test.tsx:102`; `contrast.test.ts:55`; `numeric-text-field.test.tsx:72`; **novo**: `features/estoque/produtos/produtos-page.test.tsx:336` e `features/financeiro/lancamentos/lancamento-detail.test.tsx:351` (`13f5ad7`), mutantes M28 e M29 mortos; navegador: anel `#111111` de 2 px com afastamento branco de 2 px (`evidence/foco-e-hover.md`) | comprovado (mudou nesta rodada) |
| 3 | foco sem vermelho nem cor de erro | `actions.test.tsx:117`; `components/ui/field.test.tsx:374`; os dois testes novos de D1 recusam `destructive`/`brand` no foco | comprovado (mudou nesta rodada) |
| 4 | toque ≥ 44 × 44 | `tokens.test.ts:294`; `data.test.tsx:269`; `evidence/smoke-visual.md:121-130` | comprovado |
| 5 | ≥ 24 × 24 | `tokens.test.ts:262`; `field.test.tsx:422` | comprovado |
| 6 | ícones decorativos ocultos | `actions.test.tsx:137`; `status-badge.test.tsx:40` | comprovado |
| 7 | estado não só por cor | `display.test.tsx:104,163,177` | comprovado |

### DSC-02 — Ações

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | variantes e tamanhos | `actions.test.tsx:22,52,66,74,89` | comprovado |
| 2 | destrutiva sólida, texto ≥ 4,5:1 | `actions.test.tsx:59`; `contrast.test.ts:34` | comprovado |
| 3 | IconButton com nome obrigatório | `actions.test.tsx:215,242` (`@ts-expect-error` pelo `tsc`); mutante M04 morto | comprovado |
| 4 | desabilitado, `aria-busy`, rótulo com "…" | `actions.test.tsx:150,178`; `confirm-dialog.test.tsx:206,230` | comprovado |
| 5 | navegação por link | `app-shell.test.tsx:140`; `overlays.test.tsx:488`; `components.test.tsx:57` | comprovado |

### DSC-03 — Campos

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | rótulo associado e clicável | `field.test.tsx:32,196` | comprovado |
| 2 | `aria-invalid`, `aria-describedby`, ícone e texto | `field.test.tsx:42,85,233,264`; mutante M06 morto | comprovado |
| 3 | mesma altura, borda e estados | `field.test.tsx:345,353,380` | comprovado |
| 4 | select nativo do kit com cores explícitas | `field.test.tsx:325`; `global.guard.test.ts:177` | comprovado |
| 5 | Select rico só com busca ou conteúdo rico | `global.guard.test.ts:177`; `components.test.tsx:34` | comprovado |
| 6 | controle e rótulo num alvo | `field.test.tsx:391,410,422` | comprovado |
| 7 | "(opcional)" | `field.test.tsx:210,220` | comprovado |
| 8 | Switch e Radio | não implementados (WDS-D-022); `conta-dialogs.tsx:112` usa rádio nativo | comprovado |
| 9 | ausentes no espécime | `components.test.tsx:34` | comprovado |

### DSC-04 — Sobreposições

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | foco entra e volta | `overlays.test.tsx:231,353,407,463` | comprovado |
| 2 | Esc fecha | `overlays.test.tsx:243,366,419` | comprovado |
| 3 | `overscroll-behavior: contain` | `overlays.test.tsx:306,396,448` | comprovado |
| 4 | < 640 px em tela inteira, ações fixas | `overlays.test.tsx:289,314,328`; `evidence/smoke-visual.md:231` | comprovado |
| 5 | sem diálogo sobre diálogo | sem teste; inspeção parcial (`users-page.tsx:55`). WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 6 | região viva; quatro tons com ícone | `display.test.tsx:252,266`; D5 em `STATE.md:299` | comprovado |
| 7 | dispensável; pausa em foco e hover | `display.test.tsx:279,293,349,361` | comprovado |
| 8 | Tooltip | não implementado (WDS-D-022) | comprovado |
| 9 | AlertDialog para destrutivas | `overlays.test.tsx:388`; `confirm-dialog.test.tsx:17` | comprovado |

### DSC-05 — Dados e navegação interna

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | cabeçalhos, nome, numérico à direita | `components/ui/data.test.tsx:44,55`; `data-list.test.tsx:81,122` | comprovado |
| 2 | tons neutro, sucesso, aviso, erro e informação | 4 tons; informação fundida no neutro (WDS-D-018); `display.test.tsx:95-104`. WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 3 | status distintos por rótulo e ícone | `status-badge.test.tsx:59`; `lancamentos-page.test.tsx:417-436`; mutante M14 morto | comprovado |
| 4 | kit sem tipos de módulo | `status-badge.test.tsx:67` | comprovado |
| 5 | uso do Card | regra sem valor verificável; `data.test.tsx:183` cobre só a estrutura. WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 6 | Breadcrumb com o último item atual sem link | `data.test.tsx:207,219,233` | comprovado |
| 7 | "Mostrar mais" informa visíveis e total | `data.test.tsx:293,298` | comprovado |
| 8 | Tabs | não implementado (WDS-D-022) | comprovado |

### DSC-06 — Domínio

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | `formatBRL`, tabular | `money-value.test.tsx:18,29` | comprovado |
| 2 | sinal de menos; ícone no saldo negativo | `money-value.test.tsx:65`; `saldo-page.test.tsx:43`; mutantes M11a e M11b mortos | comprovado |
| 3 | quatro variantes, nenhuma só por cor | `money-value.test.tsx:38,44,51,58,65,76` | comprovado |
| 4 | MoneyField e QuantityField como FND-05 AC4/AC5 | testes de `money-field` e `quantity-field` sem alteração desde a base, passando | comprovado |
| 5 | selo de papel neutro em pt-BR | `role-badge.test.tsx:8` | comprovado |

### DSC-07 — Testes do kit

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | papel, nome e teclado | `components/ui/tokens.guard.test.ts:75`; `components/app/tokens.guard.test.ts:84`; suítes do kit | comprovado |
| 2 | testes existentes passam | gate `pnpm test`: 110 arquivos, 1794/1794 em `023558b` | comprovado (mudou nesta rodada) |
| 3 | alteração de teste cita decisão | `evidence/testes-alterados.md` (inclui `fb62b99` e esta rodada) | comprovado (mudou nesta rodada) |

### DSP-01 — Estrutura de página

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | contêiner com largura nomeada | `page.test.tsx:81,91`; `app/(app)/financeiro/saldo/page.test.tsx:49`; `global.guard.test.ts:193` | comprovado |
| 2 | exatamente um `h1`, descrição e ações | `page.test.tsx:110,126`; `global.guard.test.ts:183` | comprovado |
| 3 | no máximo uma ação primária | `contas-page.test.tsx:411`; inspeção de `lancamento-detail.tsx` (evidência parcial). WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 4 | trilha no detalhe | `page.test.tsx:140`; `lancamento-detail.test.tsx:237`; `produto-detail.test.tsx:166` | comprovado |
| 5 | `h1`→`h2`→`h3` sem salto | `states.test.tsx:98,136`; `data.test.tsx:358,374` | comprovado |
| 6 | < 768 ações abaixo, primária em largura total | `page.test.tsx:160` | comprovado |

### DSP-02 — Estados de tela

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | esqueleto anunciado | `states.test.tsx:14,62,82` | comprovado |
| 2 | vazio com ação de criar | `states.test.tsx:98`; `users-page.layout.test.tsx:122`; `contas-page.test.tsx:468` | comprovado |
| 3 | sem resultado com "Limpar filtros" | `states.test.tsx:113,123` | comprovado |
| 4 | erro com ícone e "Tentar de novo" | `states.test.tsx:34,144`; `app/(app)/layout.test.tsx:120` | comprovado |
| 5 | "Sem acesso" com volta ao início | `require-permission.test.tsx:60,126`; smoke T14 | comprovado |
| 6 | estado não desloca o cabeçalho | sem valor verificável na spec. WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |

### DSP-03 — Feedback e confirmação

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | toast de sucesso, mudança sem recarga | `workflow.test.tsx`; smoke T14 | comprovado |
| 2 | confirmação nomeia o registro e a consequência | `confirm-dialog.test.tsx:110`; `dialog-patterns.test.tsx:17`; `workflow.test.tsx:372` | comprovado |
| 3 | verbo da ação | `confirm-dialog.test.tsx:250`; `workflow.test.tsx:66` | comprovado |
| 4 | motivo obrigatório (FND-05 AC6) | `confirm-dialog.test.tsx:83,90,98` | comprovado |
| 5 | falha mantém aberto com erro dentro | `confirm-dialog.test.tsx:159,177` | comprovado |
| 6 | indisponível desabilitado com motivo | `ajuste-dialog.test.tsx:197` | comprovado |
| 7 | mensagens FND-06 sem mudança | só `comprovantes/errors.ts:3,7` (DSP-07 AC6) | comprovado |

### DSP-04 — Formulários

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | coluna única, rótulo acima | `form-layout.test.tsx:33`; `numeric-text-field.test.tsx:55` | comprovado |
| 2 | erro ao sair do campo | `numeric-text-field.test.tsx:127` (só numéricos; evidência parcial). WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 3 | foco no primeiro inválido ao enviar | `access-screens.test.tsx:65,81,98`; `produtos-page.test.tsx:185`; `lib/forms/use-focus-invalid-after-submit.test.tsx:40` | comprovado |
| 4 | erro sem campo em alerta no topo | `form-layout.test.tsx:69,85`; `form-error-alert.test.tsx:26` | comprovado |
| 5 | envio desabilita e mostra progresso | `form-layout.test.tsx:137` | comprovado |
| 6 | todo campo declara `autocomplete` e tipo/modo | guardas `global.guard.test.ts:213,221` e **nova** `global.guard.test.ts:246` (todo `Input` com `name` e `autoComplete`, o de arquivo só `name`; todo `MoneyField` com `name`), que falhava com 14 ocorrências antes de `9e95fb9` e `131a5c4`; mutantes M18, M23, M25 e G1 a G5 mortos; `app/(dev)/design-system/page.test.tsx:77`; `subpages.test.tsx:75` | comprovado (mudou nesta rodada) |
| 7 | e-mail e código sem correção ortográfica | `create-produto-dialog.tsx:116`; `global.guard.test.ts:280`; `access-screens.test.tsx:55,73`; mutante M17 morto | comprovado |
| 8 | colar não bloqueado | inspeção estática: nenhum `onPaste` | comprovado |
| 9 | < 768 ações em largura total, principal por último | `form-layout.test.tsx:119` | comprovado |

### DSP-05 — Listagens

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | barra com contagem e "Limpar filtros" | `filter-bar.test.tsx:34,43,48,100` | comprovado |
| 2 | mudança vai à URL | `use-query-state.test.tsx:95,105`; `lancamentos-page.test.tsx:340`; `users-page.layout.test.tsx:65` | comprovado |
| 3 | restaura da URL | `use-query-state.test.tsx:209`; `lancamentos-page.test.tsx:369` | comprovado |
| 4 | valor inválido ignorado | `codecs.test.ts:38`; `use-query-state.test.tsx:235`; `lancamentos-page.test.tsx:387`; mutante M07 morto | comprovado |
| 5 | ≥ 1024 completa, sem P3 quando estreito, sem rolagem do documento | `data-list.test.tsx:140,154,166,173`; smoke T14 | comprovado |
| 6 | 768-1023 por prioridade | `data-list.test.tsx:140,197` | comprovado |
| 7 | < 768 empilhado | `data-list.test.tsx:215,231`; `lancamentos-page.test.tsx:493` | comprovado |
| 8 | tabela pequena mantida | `data-list.test.tsx:302` | comprovado |
| 9 | rolagem com nome, teclado e 1ª coluna fixa | `data.test.tsx:149,164` | comprovado |
| 10 | ação de linha identifica o registro | `data-list.test.tsx:319,337`; `lancamentos-page.test.tsx:511-522`; `movimentacoes-panel.test.tsx:417-425`; mutantes M05 e M05b mortos | comprovado |
| 11 | uma ação visível, demais em menu | `data-list.test.tsx:344,356` | comprovado |
| 12 | 50 e +50 | `lancamentos-page.test.tsx:453,462` | comprovado |
| 13 | filtro sobre todos | `lancamentos-page.test.tsx:480` | comprovado |
| 14 | cursor | `users-page.test.tsx` ("Mostrar mais" com `cursor`) | comprovado |

### DSP-06 — Permissões

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | sem permissão, ação não aparece | `require-permission.test.tsx:102,113`; `users-page.permissions.test.tsx` | comprovado |
| 2 | só leitura sem controles de escrita | smoke T14 (conselho) | comprovado |
| 3 | negação da API mostra mensagem de permissão | `components/app/api-error.test.tsx:14` | comprovado |
| 4 | "Sem acesso" igual em todas as áreas | `require-permission.test.tsx:124`; `app/(app)/sem-acesso/page.test.tsx:7` | comprovado |

### DSP-07 — Texto

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | pt-BR, frase comum | sem teste nem guarda; sem valor verificável. WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 2 | rótulo começa por verbo | sem teste; link "Detalhes" (verbo implícito). WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 3 | erro do front diz o que houve e o próximo passo | sem teste geral. WDS-D-034 (`STATE.md:154`): lacuna de precisão da spec aceita; não é teste que passou | atendido por decisão documentada |
| 4 | "…" e espaço não separável entre número e unidade | `global.guard.test.ts:267,291`; `actions.test.tsx:178`; `produto-detail.test.tsx:50`; `ajuste-dialog.test.tsx:55-65`; `movimentacoes-panel.test.tsx:423-424`; mutantes M24, M26 e M27 mortos; navegador (iteração 3) | comprovado |
| 5 | progresso termina em "…" | `actions.test.tsx:178` | comprovado |
| 6 | limite em MB inteiros, para baixo | `comprovantes-panel.test.tsx:196` | comprovado |
| 7 | tipo + código curto monoespaçado | `format.test.ts:51,55`; `lancamento-detail.test.tsx:237` | comprovado |

### DSS-01 — Shell

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | tokens do shell e do conteúdo | `shell-frame.test.tsx:211` | comprovado |
| 2 | escudo + "Torcida Jovem", link ao início, sem "TJ Platform" | `shell-frame.test.tsx:163,206`; `app-shell.test.tsx:116`; **novo**: hover da marca `shell-frame.test.tsx:182` (`023558b`, D4), medido no navegador (`evidence/foco-e-hover.md`) | comprovado (mudou nesta rodada) |
| 3 | grupos com rótulo; ícone e texto | `shell-frame.test.tsx:238,250,269`; `nav-items.test.ts:17` | comprovado |
| 4 | `aria-current`, indicador vermelho, peso maior | `shell-frame.test.tsx:280-298` | comprovado |
| 5 | primeiro focável "Pular para o conteúdo" | `shell-frame.test.tsx:84,95,109`; `app-shell.test.tsx:127`; mutante M10 morto | comprovado |
| 6 | uma região principal; navegações nomeadas | `shell-frame.test.tsx:145,151`; `app/(app)/layout.test.tsx:105` | comprovado |
| 7 | nome; "Trocar senha" link; "Sair" botão | `shell-frame.test.tsx:422`; `app-shell.test.tsx:140` | comprovado |
| 8 | FND-05 AC1 e AC8 | `app-shell.test.tsx:55,66,177` | comprovado |
| 9 | texto do shell ≥ 4,5:1 | pares `sidebar-*` em `contrast.test.ts:34` | comprovado |

### DSS-02 — Navegação responsiva

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | ≥ 1024 fixa | `shell-frame.test.tsx:322`; `app-shell.test.tsx:101`; mutante M09 morto; smoke T14 | comprovado |
| 2 | < 1024 gaveta, botão 44 × 44 nomeado | `shell-frame.test.tsx:329,336`; smoke T14 | comprovado |
| 3 | escolher item fecha | `shell-frame.test.tsx:351`; smoke T14 | comprovado |
| 4 | foco volta ao botão | `shell-frame.test.tsx:361,373`; smoke T14 | comprovado |
| 5 | cabeçalho visível; nada inacessível ≥ 320 | `evidence/smoke-visual.md:66-85` | comprovado |
| 6 | áreas seguras | `shell-frame.test.tsx:228`; `STATE.md:288` | comprovado |
| 7 | navegador real em 390, 768, 820, 1024 e 1440 | `evidence/smoke-visual.md:24,70-85`; smoke T14 | comprovado |

### DSS-03 — Acesso

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | escudo e nome | `access-screens.test.tsx:28`; `access-layout.test.tsx:23` | comprovado |
| 2 | mesmo layout estreito | `access-layout.test.tsx:60,71,80` | comprovado |
| 3 | FND-03, ACS-01 a 03 | testes de acesso da V1 passando | comprovado |
| 4 | `autocomplete` de autenticação; e-mail sem correção | `access-screens.test.tsx:38,55,73,89` | comprovado |
| 5 | sem ilustração | `access-layout.test.tsx:50` | comprovado |
| 6 | < 768 sem rolagem, campos em largura total | `access-screens.test.tsx:46`; `evidence/smoke-visual.md:72` | comprovado |

### DSS-04 — Páginas de erro

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | erro no shell com "Tentar de novo" e "Voltar ao início" | `app/error-pages.test.tsx:38,48,56`; **novo**: exceção real no cliente e no servidor em build de produção, dentro do shell, sem loop (`evidence/erro-no-shell.md`) | comprovado (mudou nesta rodada) |
| 2 | erro raiz com escudo e "Tentar de novo" | `error-pages.test.tsx:77,89`; na simulação, o `global-error` não é acionado por erro de rota | comprovado (mudou nesta rodada) |
| 3 | "Página não encontrada" com link | `error-pages.test.tsx:65`; smoke T14 | comprovado |
| 4 | sem detalhe técnico | `error-pages.test.tsx:14,38,77`; simulação: nenhuma mensagem, pilha ou `digest` na tela (`evidence/erro-no-shell.md`) | comprovado (mudou nesta rodada) |
| 5 | `h1` com foco | `error-pages.test.tsx:38,65,77`; simulação: foco no `h1` ao abrir e depois de "Tentar de novo" | comprovado (mudou nesta rodada) |

### DSS-05 — Metadados e ícones

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | `lang="pt-BR"` e `theme-color` do shell | `fonts.test.ts:75,84` | comprovado |
| 2 | "Página \| TJ Platform" | `fonts.test.ts:84`; smoke T14 | comprovado |
| 3 | ícone derivado do escudo | `brand.test.ts:70,74` | comprovado |
| 4 | dependência de asset reduzido | `STATE.md:180` | comprovado |

### DSS-06 — Página inicial

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | atalhos da lista da navegação | `app/(app)/inicio/atalhos.test.tsx:53,69` | comprovado |
| 2 | links com ícone, nome e descrição | `atalhos.test.tsx:53` | comprovado |
| 3 | sem indicadores | `atalhos.test.tsx:87` | comprovado |
| 4 | < 768 uma coluna | `atalhos.test.tsx:76` | comprovado |

### DSM-01 — Migração sem regressão

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | ACs da Web V1 válidos | 1794/1794 testes em `023558b`; `evidence/testes-alterados.md` | comprovado (mudou nesta rodada) |
| 2 | nada fora de `web/`, `.specs/`, `docs/development/` e `.gitignore` | esta rodada só mudou `web/` e `.specs/features/web-design/` | comprovado |
| 3 | sem dependência nova | `git diff --name-only feature/web-design -- web/package.json web/pnpm-lock.yaml` vazio | comprovado |
| 4 | alteração de teste cita decisão; lista no relatório | `evidence/testes-alterados.md` (inclui `fb62b99` e as asserções desta rodada) | comprovado (mudou nesta rodada) |
| 5 | nenhum teste removido ou enfraquecido | 110 arquivos de teste antes e depois; só acréscimos nesta rodada | comprovado |
| 6 | espécime aprovado antes | `STATE.md:147-148` | comprovado |

### DSM-02 — Identity

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | cabeçalho, filtros, lista responsiva | `users-page.layout.test.tsx:21` | comprovado |
| 2 | filtro na URL e `active`/`role` na API | `users-page.layout.test.tsx:65-82` | comprovado |
| 3 | menu nomeado pelo usuário | `users-page.layout.test.tsx:176`; `evidence/smoke-visual.md:31` | comprovado |
| 4 | troca de senha no layout de acesso | `app/(app)/conta/senha/page.test.tsx:41` | comprovado |
| 5 | USR-01 a 05, ACS-01 a 03 | testes da V1 passando | comprovado |

### DSM-03 — Financeiro

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | padrões nas telas | `contas-page.test.tsx:399`; `lancamento-detail.test.tsx:237`; `saldo-page.test.tsx:79`; `comprovantes-panel.test.tsx:389` | comprovado |
| 2 | MoneyValue tabular | `lancamento-detail.test.tsx:74,237`; `saldo-page.test.tsx:79` | comprovado |
| 3 | quatro status distintos | `lancamentos-page.test.tsx:417-436`; mutante M14 morto | comprovado |
| 4 | tipo, status, conta e período na URL | `lancamentos-page.test.tsx:340-410` | comprovado |
| 5 | < 768 data, tipo, conta, líquido e status | `lancamentos-page.test.tsx:493-506` | comprovado |
| 6 | link de detalhe nomeado | `lancamentos-page.test.tsx:511-522` | comprovado |
| 7 | hierarquia por recuo | `contas-page.test.tsx:418-419` | comprovado |
| 8 | FWB-01 a 05 | testes da V1 passando; smoke T14 | comprovado |

### DSM-04 — Estoque

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | padrões nas telas | `produtos-page.test.tsx:307,391`; `produto-detail.test.tsx:166` | comprovado |
| 2 | busca na URL | `produtos-page.test.tsx:264` | comprovado |
| 3 | < 768 histórico empilhado | `movimentacoes-panel.test.tsx:429` | comprovado |
| 4 | "Devolver" nomeia a origem | `movimentacoes-panel.test.tsx:417-425` | comprovado |
| 5 | diálogos em tela inteira < 640 | `overlays.test.tsx:289,314`; `evidence/smoke-visual.md:179` | comprovado |
| 6 | tabular; negativo com ícone e texto | `ajuste-dialog.test.tsx:197`; `produto-detail.test.tsx:166` | comprovado |
| 7 | EWB-01 a 04 | testes da V1 passando | comprovado |

### DSM-05 — Limpeza

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | sem contêiner, largura ou título próprios | `global.guard.test.ts:183,188,193,208` | comprovado |
| 2 | sem `<select>` fora do kit | `global.guard.test.ts:177` | comprovado |
| 3 | `select-field` e filtro de data substituídos | `git ls-files \| grep select-field` vazio | comprovado |
| 4 | guarda em todo o front | `global.guard.test.ts:163,168`; mutante M02 morto | comprovado |
| 5 | sem variantes `dark:` | regra `dark-variant` (`guard.ts:87`) | comprovado |

### DSM-06 — Verificação

| AC | Resultado definido na spec (resumo) | Evidência | Situação |
| --- | --- | --- | --- |
| 1 | gates passam | `pnpm lint`, `next typegen`, `tsc --noEmit`, `pnpm test` (1794/1794), `pnpm build`, `lint:api`, `gen:api:check` com saída 0 em `023558b` | comprovado (mudou nesta rodada) |
| 2 | `web-design-guidelines`: cada achado corrigido ou registrado | F1/D2 em `131a5c4`, F2/D3 em `9e95fb9`, F3/D4 em `023558b`, cada um com teste que falhava antes; A1 a A3 e O1 da iteração 2 já corrigidos. A reexecução da revisão cabe à 4ª verificação (WDS-D-035) | comprovado (mudou nesta rodada) |
| 3 | navegador real, produção, cinco larguras, todas as telas | `evidence/smoke-visual.md:66-85,152-181` | comprovado |
| 4 | percurso só por teclado | `evidence/smoke-visual.md:87-97`; `2903f32` | comprovado |
| 5 | 200% e 320 px | `evidence/smoke-visual.md:101-117` | comprovado |
| 6 | verificador independente com `validation.md` | `validation.md` (iteração 3); 4ª verificação autorizada em WDS-D-035 | comprovado |
| 7 | divergência vira correção | N1 a N3 (`evidence/smoke-visual.md:229-231`); D1 a D4 nesta rodada | comprovado (mudou nesta rodada) |
