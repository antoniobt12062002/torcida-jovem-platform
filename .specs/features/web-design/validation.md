# Web Design — Validation (T14)

**Veredito: FAIL ❌**

**Date**: 2026-10-13
**Spec**: `.specs/features/web-design/spec/01` a `05` (DSF, DSC, DSP, DSS, DSM); Web V1 em `.specs/features/web/spec/01` a `06`
**Diff range**: `7c30c6a..f9de714` (base documental `7c30c6a` sobre a `develop` em `cebbdf2`; 72 commits; HEAD `f9de7140e523f13f806bde40acf6696026f1e513`)
**Verifier**: sub-agente independente (autor ≠ verificador). Relatórios dos implementadores não foram usados como prova; cada afirmação abaixo vem de `arquivo:linha`, saída de comando ou medição no navegador.

Motivo do FAIL, em uma linha: a revisão `web-design-guidelines` (gate obrigatório de DSM-06 AC2 e T14) tem achados sem correção e sem conflito registrado, dois deles também violando ACs da spec (DSP-04 AC6 e AC7); e um mutante do sensor sobreviveu (M13).

---

## Task Completion

| Task | Status | Notes |
| --- | --- | --- |
| Rodadas 1 a 6 (DS-F, DS-T, DS-K, DS-L, DS-A, DS-E, DS-S, DS-ID, DS-FA, DS-FB, DS-ES) | ✅ | Registro de execução em `STATE.md:261-276` |
| T11 (remoção de alias, guarda global) | ✅ | `663febe`; `web/test/design/global.guard.test.ts` |
| T12 (guia e revisão estrutural) | ✅ | `2cc62da`; `evidence/revisao-estrutural.md` |
| T13 (smoke completo) | ✅ | `evidence/smoke-visual.md`; correções N1 (`ad67ac5`), N2 e N3 (`2903f32`) |
| T14 (esta verificação) | ❌ | Gate `web-design-guidelines` reprova; mutante M13 sobrevive |

---

## Gate Check (repositório principal, `web/`, `API_URL=http://localhost:8080`)

| Gate | Saída |
| --- | --- |
| `pnpm lint` | exit 0 |
| `pnpm exec next typegen` | exit 0, "Types generated successfully" |
| `pnpm exec tsc --noEmit` | exit 0 |
| `pnpm test` | exit 0 — **110 arquivos, 1774 testes, 1774 passaram, 0 falharam, 0 pulados** |
| `pnpm build` | exit 0 (Next 16.4.0, 20 páginas; `/design-system` pré-renderizada e respondendo 404 em produção). Rodado no worktree temporário em `f9de714` (árvore idêntica ao HEAD), para não sobrescrever o `.next` do `next start` que servia o smoke em `:3000` |
| `pnpm lint:api` | exit 0 |
| `pnpm gen:api:check` | exit 0 |
| `pnpm audit --audit-level high` | exit 0 — "1 vulnerabilities found, Severity: 1 high (1 ignored)": `GHSA-vfj7-8cjw-p6xm` ignorada em `web/pnpm-workspace.yaml:10-12`, arquivo sem alteração no ciclo (já vinha da `develop`) |
| `git diff --name-only develop -- api/ web/package.json web/pnpm-lock.yaml` | vazio |

- **Testes antes do ciclo**: 69 arquivos / 851 testes (`STATE.md:265`; contagem de arquivos reconferida: `git ls-tree -r --name-only 7c30c6a web` dá 69 `*.test.ts(x)`).
- **Depois**: 110 arquivos / 1774 testes. Delta: +41 arquivos, +923 testes.
- **Arquivos de teste removidos**: nenhum (`comm -23` da lista de `7c30c6a` contra HEAD vazio).

---

## Spec-Anchored Acceptance Criteria

Legenda: ✅ o teste/evidência afirma o valor que a spec define · ❌ lacuna (sem evidência, ou o código contradiz a spec) · ⚠️ lacuna de precisão da spec (a spec não define valor verificável, ou spec e design aprovado divergem). "Smoke T14" = medição deste verificador no build de produção (seção Smoke).

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
| 4 | caixa alta só em título de página, rótulo de grupo e sobretítulo | `components/ui/data.test.tsx:390-394` afirma `uppercase` também no `DisplayTitle`, usado no **nome da marca** (`components/app/brand.tsx:67`, `shell-frame.tsx:101`). O design aprovado manda "TORCIDA JOVEM… caixa alta" (`design/01-fundacao-visual.md:192,335`); a spec não lista nome da marca nem destaque | ⚠️ spec × design |
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
| 4 | sem gradiente/brilho | `tokens.test.ts:375`; grep de `gradient` vazio no código | ✅ |
| 5 | durações dos tokens, sem `transition: all` | `tokens.test.ts:353`; regra `transition-all` | ✅ |
| 6 | redução de movimento | `tokens.test.ts:364`; `evidence/smoke-visual.md:134-140` | ✅ |
| 7 | três larguras nomeadas | `components/app/page.test.tsx:91,103`; `global.guard.test.ts:176` | ✅ |
| 8 | dois pontos de quebra, 768 e 1024 | `lib/responsive/use-media-query.test.tsx:41`; `tokens.test.ts:307`. Mas DSC-04 AC4 exige 640 px (Dialog) e o shell usa `min-[400px]` (`shell-frame.tsx:101`) | ⚠️ spec contraditória (DSF-04 AC8 × DSC-04 AC4) |

### DSF-05 Marca

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | traçados e cores idênticos, `viewBox` no círculo | `test/design/brand.test.ts:39,44,49,54`; sensor M03 morto | ✅ |
| 2 | 1:1, sem efeito | `components/app/page.test.tsx:40` (largura = altura) | ✅ |
| 3 | ≥ 32 px e respiro | `page.test.tsx:47` | ✅ |
| 4 | alt vazio com nome; "Torcida Jovem" sozinho | `page.test.tsx:25,32` | ✅ |
| 5 | largura e altura | `page.test.tsx:40` | ✅ |
| 6 | sem substituto; dependência em `STATE.md` | `brand.test.ts:89`; `STATE.md:174-177` | ✅ |
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
| 6 | migração só após aprovação | `STATE.md:147-148` (WDS-D-033, 2026-10-12) antes das integrações da rodada 6 (`STATE.md:276`) | ✅ |

### DSC-01 Estados, foco e alvos

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | repouso, hover, pressionado, foco, desabilitado distintos | `actions.test.tsx:39,107,120` (classes); aparência só no espécime aprovado (`evidence/especime.md`) — "distinta" não tem valor verificável | ⚠️ |
| 2 | foco ≥ 2 px, ≥ 3:1 no claro e no shell | `actions.test.tsx:107`; `shell-frame.test.tsx:102`; par `ring` em `contrast.test.ts:55` | ✅ |
| 3 | foco sem vermelho nem erro | `actions.test.tsx:107`; `components/ui/field.test.tsx:374` | ✅ |
| 4 | toque ≥ 44 × 44 | `tokens.test.ts:294`; `data.test.tsx:269`; `evidence/smoke-visual.md:121-130` | ✅ |
| 5 | ≥ 24 × 24 | `tokens.test.ts:262` (alturas 32/36/44); `field.test.tsx:422` | ✅ |
| 6 | ícones decorativos ocultos | `actions.test.tsx:127`; `status-badge.test.tsx:40` | ✅ |
| 7 | estado não só por cor | `display.test.tsx:104,163,177` | ✅ |

### DSC-02 Ações

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | variantes e tamanhos | `actions.test.tsx:22,52,66,74,89` | ✅ |
| 2 | destrutiva sólida, texto ≥ 4,5:1 | `actions.test.tsx:59`; pares em `contrast.test.ts:34` | ✅ |
| 3 | IconButton com nome obrigatório | `actions.test.tsx:204,220` (`@ts-expect-error` conferido pelo `tsc`); sensor M04 morto (34 testes) | ✅ |
| 4 | desabilitado, `aria-busy`, rótulo "…" | `actions.test.tsx:140,168`; `confirm-dialog.test.tsx:198,222` | ✅ |
| 5 | navegação por link | `app-shell.test.tsx:140`; `overlays.test.tsx:488`; `components.test.tsx:57` | ✅ |

### DSC-03 Campos

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | rótulo associado e clicável | `field.test.tsx:32,196` | ✅ |
| 2 | `aria-invalid`, `aria-describedby`, ícone e texto | `field.test.tsx:42,85,233,264`; sensor M06 morto | ✅ |
| 3 | mesma altura, borda e estados | `field.test.tsx:345,353,380` | ✅ |
| 4 | select nativo no kit com cores explícitas | `field.test.tsx:325`; `global.guard.test.ts:160` | ✅ |
| 5 | Select rico só com busca/conteúdo rico | `global.guard.test.ts:160`; `components.test.tsx:34`; nenhum uso do Select rico em `features/` | ✅ |
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
| 5 | sem diálogo sobre diálogo | Nenhum teste. Inspeção: cada tela mantém um único diálogo aberto por estado (`features/identity/usuarios/users-page.tsx:55`) | ⚠️ sem teste |
| 6 | região viva; quatro tons com ícone | `display.test.tsx:252,266`; D5 (toast `role=dialog`) registrado em `STATE.md:293` | ✅ |
| 7 | dispensável; pausa em foco/hover | `display.test.tsx:279,293,349,361` | ✅ |
| 8 | Tooltip | não implementado (WDS-D-022) | ✅ |
| 9 | AlertDialog para destrutivas | `overlays.test.tsx:388`; `confirm-dialog.test.tsx:17` | ✅ |

### DSC-05 Dados e navegação interna

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | cabeçalhos, nome, numérico à direita | `components/ui/data.test.tsx:44,55`; `data-list.test.tsx:81,122` | ✅ |
| 2 | tons neutro, sucesso, aviso, erro **e informação** | `display.test.tsx:95-104` e `status-badge.tsx:9` têm 4 tons; informação foi fundida no neutro (`badge.tsx:6`, WDS-D-018 "informação em neutro") | ⚠️ spec × decisão |
| 3 | status distintos por rótulo e ícone | `status-badge.test.tsx:59`; `lancamentos-page.test.tsx:417-436`; sensor M14 morto | ✅ |
| 4 | kit sem tipos de módulo | `status-badge.test.tsx:67` | ✅ |
| 5 | uso do Card | regra de uso, sem valor verificável; `data.test.tsx:183` só cobre a estrutura | ⚠️ |
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
| 1 | contêiner com largura nomeada | `page.test.tsx:81,91`; `app/(app)/financeiro/saldo/page.test.tsx:49`; `global.guard.test.ts:176` | ✅ |
| 2 | exatamente um `h1`, descrição, ações | `page.test.tsx:110,126`; `global.guard.test.ts:166` | ✅ |
| 3 | no máximo uma primária | Só `contas-page.test.tsx:411`; nenhuma guarda geral. Inspeção de `lancamento-detail.tsx:118-122,263,303` confere uma primária | ⚠️ cobertura parcial |
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
| 6 | estado não desloca o cabeçalho | Sem valor verificável na spec; nenhum teste mede posição | ⚠️ |

### DSP-03 Feedback e confirmação

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | toast de sucesso e mudança sem recarga | `workflow.test.tsx` (Web V1, inalterado na asserção de toast); Smoke T14 ("Lançamento criado.", "Recebimento registrado.", "Comprovante anexado.") | ✅ |
| 2 | nomeia o registro e a consequência | `confirm-dialog.test.tsx:102`; `dialog-patterns.test.tsx:17`; `workflow.test.tsx:372` | ✅ |
| 3 | verbo da ação | `confirm-dialog.test.tsx:242`; `workflow.test.tsx:66` ("Receber receita", "Pagar despesa") | ✅ |
| 4 | motivo obrigatório (FND-05 AC6) | `confirm-dialog.test.tsx:75,82,90` | ✅ |
| 5 | falha mantém aberto com erro dentro | `confirm-dialog.test.tsx:151,169` | ✅ |
| 6 | indisponível desabilitado com motivo | `ajuste-dialog.test.tsx:197` | ✅ |
| 7 | mensagens FND-06 sem mudança | `git diff 7c30c6a f9de714 -- web/features/*/errors.ts`: só `comprovantes/errors.ts:3,7` (troca `formatBytes` → `formatLimit`, motivada por DSP-07 AC6) | ✅ |

### DSP-04 Formulários

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | coluna única, rótulo acima | `form-layout.test.tsx:33`; `numeric-text-field.test.tsx:55` | ✅ |
| 2 | erro ao sair do campo | `numeric-text-field.test.tsx:103` cobre só os campos numéricos; nenhum teste para os campos de texto | ⚠️ cobertura parcial |
| 3 | foco no primeiro inválido ao enviar | `access-screens.test.tsx:65,81,98`; `produtos-page.test.tsx:185`; `lib/forms/use-focus-invalid-after-submit.test.tsx:40` | ✅ |
| 4 | erro sem campo em alerta no topo | `form-layout.test.tsx:69,85`; `form-error-alert.test.tsx:26` | ✅ |
| 5 | envio desabilita e mostra progresso | `form-layout.test.tsx:137` | ✅ |
| 6 | **todo campo declara `autocomplete`** e tipo/modo | **Nenhum teste.** Sem `autoComplete`: `components/app/confirm-dialog.tsx:149`, `features/estoque/ajustes/ajuste-dialog.tsx:107`, `features/identity/usuarios/dialogs/admin-dialogs.tsx:139`, `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:116` (textareas de motivo, também sem `name`), `features/estoque/produtos/produtos-page.tsx:93` (busca), `features/financeiro/lancamentos/lancamentos-page.tsx:209,220,234,245,249` e `features/identity/usuarios/users-page.tsx:156,170` (filtros) | ❌ |
| 7 | **e-mail e código** sem correção ortográfica | E-mail coberto (`access-screens.test.tsx:55,73`). **Código do produto** sem `spellCheck={false}`: `features/estoque/produtos/create-produto-dialog.tsx:114` (únicos `spellCheck` do código: `login-form.tsx:112`, `request-reset-form.tsx:91`, `create-user-dialog.tsx:127`) | ❌ |
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
| 7 | < 768 empilhado | `data-list.test.tsx:215,231`; `lancamentos-page.test.tsx:490` | ✅ |
| 8 | tabela pequena mantida | `data-list.test.tsx:302` | ✅ |
| 9 | rolagem com nome, teclado, 1ª coluna fixa | `components/ui/data.test.tsx:149,164` | ✅ |
| 10 | ação de linha identifica o registro | `data-list.test.tsx:319,337`; `lancamentos-page.test.tsx:508-519`; `movimentacoes-panel.test.tsx:417-425`; sensores M05 e M05b mortos | ✅ |
| 11 | uma visível, demais em menu | `data-list.test.tsx:344,356` | ✅ |
| 12 | 50 e +50 | `lancamentos-page.test.tsx:453,462` | ✅ |
| 13 | filtro sobre todos | `lancamentos-page.test.tsx:477` | ✅ |
| 14 | cursor | `users-page.test.tsx` ("Mostrar mais" com `cursor`, `evidence/testes-alterados.md:37`) | ✅ |
| Borda: exatamente 50 | controle não aparece | `lancamentos-page.test.tsx:471-475` só afirma a ausência do botão; **mutante M13 sobreviveu** (ver Sensor) | ⚠️ teste fraco |

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
| 1 | pt-BR, frase comum | Sem teste nem guarda; só leitura das capturas | ⚠️ |
| 2 | rótulo começa por verbo | Sem teste; o link de linha é "Detalhes" (`lancamentos-page.tsx:270`), sem verbo | ⚠️ |
| 3 | erro do front diz o que houve e o próximo passo | Sem teste geral (catálogos da V1 inalterados) | ⚠️ |
| 4 | "…" e espaço não separável | `comprovantes-panel.test.tsx:373` | ✅ |
| 5 | progresso termina em "…" | `actions.test.tsx:168` (tipo `${string}…`) | ✅ |
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
| 6 | áreas seguras | `shell-frame.test.tsx:209`; `viewportFit` sem `cover` registrado (`STATE.md:282`) | ✅ |
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
| 4 | dependência de asset reduzido | `STATE.md:174` | ✅ |

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
| 2 | nada fora de `web/`, `.specs/features/web-design/` e doc de dev da Web | `git diff --name-only 7c30c6a f9de714` fora de `web/` e `.specs/features/web-design/`: só `docs/development/web.md` e **`.gitignore`** (`138d094`, uma linha `.playwright-mcp/`). `.gitignore` não está na lista do AC | ⚠️ desvio menor |
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
| 5 | < 768 data, tipo, conta, líquido, status | `lancamentos-page.test.tsx:490-503` | ✅ |
| 6 | link de detalhe nomeado | `lancamentos-page.test.tsx:508-519` | ✅ |
| 7 | hierarquia por recuo | `contas-page.test.tsx:418-419` | ✅ |
| 8 | FWB-01 a 05 | testes da V1 passando; Smoke T14 (criar, receber, anexar) | ✅ |

### DSM-04 Estoque

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | padrões nas telas | `produtos-page.test.tsx:306,368`; `produto-detail.test.tsx:166` | ✅ |
| 2 | busca na URL | `produtos-page.test.tsx:264` | ✅ |
| 3 | < 768 histórico empilhado | `movimentacoes-panel.test.tsx:429` | ✅ |
| 4 | "Devolver" nomeia a origem | `movimentacoes-panel.test.tsx:417-425` | ✅ |
| 5 | diálogos em tela inteira < 640 | `overlays.test.tsx:289,314`; `evidence/smoke-visual.md:179` | ✅ |
| 6 | tabular; negativo com ícone e texto | `ajuste-dialog.test.tsx:197`; `produto-detail.test.tsx:166` | ✅ |
| 7 | EWB-01 a 04 | testes da V1 passando | ✅ |

### DSM-05 Limpeza

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | sem contêiner, largura ou título próprios | `global.guard.test.ts:166,171,176,191` | ✅ |
| 2 | sem `<select>` fora do kit | `global.guard.test.ts:160` | ✅ |
| 3 | `select-field` e filtro de data substituídos | `git ls-files | grep select-field` vazio | ✅ |
| 4 | guarda em todo o front | `global.guard.test.ts:146,151`; sensor M02 morto | ✅ |
| 5 | sem variantes `dark:` | regra `dark-variant` (`guard.ts:87`) | ✅ |

### DSM-06 Verificação

| AC | Outcome | Evidência | Result |
| --- | --- | --- | --- |
| 1 | gates passam | seção Gate Check | ✅ |
| 2 | `web-design-guidelines` com cada achado corrigido ou registrado como conflito | **Achados sem correção nem conflito** (seção `web-design-guidelines`) | ❌ |
| 3 | navegador real, produção, 5 larguras, todas as telas | `evidence/smoke-visual.md:66-85,152-181` | ✅ |
| 4 | percurso só por teclado | `evidence/smoke-visual.md:87-97`; N2 e N3 corrigidos em `2903f32` | ✅ |
| 5 | 200% e 320 px | `evidence/smoke-visual.md:101-117` | ✅ |
| 6 | verificador independente com `validation.md` | este arquivo | ✅ |
| 7 | divergência vira correção | N1, N2, N3 corrigidos (`evidence/smoke-visual.md:229-231`) | ✅ |

**Status**: 219 ACs — **203 ✅, 3 ❌ (DSP-04 AC6, DSP-04 AC7, DSM-06 AC2), 13 ⚠️** (DSF-03 AC4, DSF-04 AC8, DSC-01 AC1, DSC-04 AC5, DSC-05 AC2, DSC-05 AC5, DSP-01 AC3, DSP-02 AC6, DSP-04 AC2, DSP-07 AC1, AC2, AC3, DSM-01 AC2), mais a borda "exatamente 50" de DSP-05 com teste fraco.

---

## Regressão da Web V1

- **Testes da V1 passam**: os 69 arquivos da base estão entre os 110 que passam (gate `pnpm test`, 1774/1774). Nenhum foi removido nem renomeado (`comm -23` vazio).
- **Arquivos da base com linhas removidas** (fora `import` e comentários), por `git diff 7c30c6a f9de714`: exatamente os 15 de `evidence/testes-alterados.md:24-40` — `app-shell`, `nav-items`, `status-badge`, `display`, `movimentacoes-panel`, os três `nav.test.ts`, `comprovantes-panel`, `lancamento-detail`, `lancamentos-page`, `workflow`, `saldo-page`, `users-page.permissions`, `users-page`. Depois de `663febe` (data do relatório), nenhum arquivo da base perdeu linha: `2903f32` só acrescentou casos em `confirm-dialog.test.tsx` e `ajuste-dialog.test.tsx`.
- **Julgamento de cada alteração**: todas trocam o valor afirmado pelo valor que uma decisão ou requisito novo define (link em vez de `router.push`, DSS-01 AC7; "+R$", WDS-D-024; "−" U+2212, DSC-06 AC2; "Receber receita", DSP-03 AC3; "9 MB", DSP-07 AC6; nomes de linha, DSP-05 AC10), mantendo as mesmas requisições e o mesmo comportamento. `movimentacoes-panel` passou de nome exato "Devolver" a `/^Devolver /`, mas o nome completo é afirmado em `movimentacoes-panel.test.tsx:423-424`, então não há enfraquecimento.
- **Casos removidos na remoção de alias** (`components/app/status-badge.test.tsx`, 23 → 11; `actions`, `display`, `data`): o comportamento ainda exigido continua coberto em outro lugar:
  - Rótulos "Criada", "Recebida", "Paga", "Cancelada": `lancamentos-page.test.tsx:431` (`toEqual(["Criada","Recebida","Paga","Cancelada"])`).
  - `CANCELADA` neutra (DD-3): `lancamentos-page.test.tsx:422,432` (`tone` `neutral`; `["neutral","success","success","neutral"]`).
  - Ícones distintos: `lancamentos-page.test.tsx:421,435` (`new Set(icons).size === 4`); sensor M14 morto.
  - "Ativo"/"Inativo": `status-badge.test.tsx:15-21,76-80`.
  - Variantes novas do Button, Badge e Alert continuam com teste próprio (`actions.test.tsx:22-89`, `display.test.tsx:104,163`).

---

## Discrimination Sensor

Worktree temporário `…/scratchpad/verify/wt` em `f9de714 --detach`, `pnpm install --frozen-lockfile`. Uma falha por vez, testes relevantes com `pnpm exec vitest run`, `git checkout -- .` entre uma e outra (status do worktree vazio depois de cada reversão). Quatro falhas (M04, M06, M08, M11b) falharam ao aplicar na primeira passada (CRLF) e foram reaplicadas com um auxiliar que respeita a quebra de linha; o resultado abaixo é o da aplicação efetiva.

| # | Falha do design | Mutação (`arquivo:linha`) | Testes rodados | Resultado |
| --- | --- | --- | --- | --- |
| M01 | token quebra contraste | `app/globals.css:178` `--muted-foreground: var(--tj-neutral-600)` → `--tj-neutral-400` | `test/design/contrast`, `tokens` | **MORTO** (5 falhas, ex.: "muted-foreground sobre background (text) atinge o mínimo") |
| M02 | cor literal numa tela | `features/financeiro/lancamentos/lancamentos-page.tsx:205` + `text-[#D81E28]` | `global.guard`, guarda da frente | **MORTO** (3) |
| M03 | traçado do escudo | `public/brand/torcida-jovem-logo.svg`, um dígito do primeiro `d=` | `test/design/brand` | **MORTO** (3) |
| M04 | IconButton sem nome | `components/ui/icon-button.tsx:32` remove `aria-label={label}` | `components/ui`, `components/app` | **MORTO** (34) |
| M05 | `rowLabel` removido | `lancamentos-page.tsx:271` `rowLabel=""` | `features/financeiro/lancamentos` | **MORTO** (3) |
| M05b | idem, no componente | `components/app/row-actions.tsx:60` `name = label` | `components/app`, `features/financeiro`, `features/estoque` | **MORTO** (32) |
| M06 | Field sem `aria-invalid`/`describedby` | `components/ui/field.tsx:79-83` força `invalid={false}` | `components/ui/field.test.tsx` | **MORTO** (6) |
| M07 | `useQueryState` aceita inválido | `lib/url-state/codecs.ts:34` `oneOf.parse` → `raw ?? fallback` | `lib/url-state`, lançamentos, usuários | **MORTO** (5) |
| M08 | parâmetro não limpo no padrão | `lib/url-state/use-query-state.ts:69-70` → `set(key, serialized ?? "")` | idem | **MORTO** (8) |
| M09 | ponto de quebra do shell em `md` | `components/app/shell-frame.tsx` todos os `lg:` → `md:` | `shell-frame`, `app-shell` | **MORTO** (3) |
| M10 | link de pular removido | `shell-frame.tsx:109` `{showNav ? (` → `{false ? (` | idem | **MORTO** (7) |
| M11a | MoneyValue sem sinal | `components/app/money-value.tsx:33` → `return amount` | `money-value`, `features/financeiro/saldo` | **MORTO** (3) |
| M11b | MoneyValue sem ícone do negativo | `money-value.tsx:50` `{negativeBalance ? (` → `{false ? (` | idem | **MORTO** (2) |
| M12 | espécime em produção | `app/(dev)/design-system/dev-only.ts:8` `=== "production"` → `=== "never"` | `app/(dev)` | **MORTO** (5) |
| M13 | "Mostrar mais" com exatamente 50 | `lancamentos-page.tsx:277` `visible.length > PAGE_SIZE` → `>=` | `lancamentos-page.test.tsx`, depois **suíte inteira** | **SOBREVIVEU** (36/36; e 1774/1774 na suíte inteira) |
| M13b | idem, no componente | `components/ui/show-more.tsx:33` `shown < total` → `<=` | kit, compostos, financeiro, identity | **MORTO** (2) |
| M14 | dois status com rótulo e ícone iguais | `features/financeiro/lancamentos/labels.ts:45` `PAGA` com rótulo e ícone de `RECEBIDA` | `features/financeiro` | **MORTO** (5) |

**Análise de M13**: com exatamente 50 registros o mutante renderiza o `ShowMore`; o botão não aparece porque o próprio componente tem uma segunda guarda (`show-more.tsx:33`, defendida por M13b), mas a contagem "Mostrando 50 de 50 lançamentos" passa a aparecer. `lancamentos-page.test.tsx:471-475` só afirma `queryByRole("button", { name: "Mostrar mais" })` nulo, então não distingue. A regra da borda de DSP-05 ("o controle 'Mostrar mais' SHALL não aparecer") fica defendida só por uma das duas guardas.

**Sensor depth**: P0 ampliado (16 injeções sobre as 14 falhas do design). **Result**: 16 injetadas, 15 mortas, 1 sobreviveu → **FAIL**.

**Isolamento**: `git status --porcelain` do repositório principal antes do sensor: vazio (0 bytes). Depois de `git worktree remove --force` (a remoção do diretório esbarrou em caminho longo do `node_modules` e foi concluída por `Remove-Item -LiteralPath \\?\…`; `git worktree list` não lista mais o temporário): vazio (0 bytes). Idênticos. HEAD continua `f9de714`.

---

## Revisão `web-design-guidelines` (gate obrigatório)

Aplicada a `web/app`, `web/components` e `web/features` (156 arquivos não-teste, `git ls-files … | grep -vE '\.test\.'`), regra a regra do `guideline.md`, com buscas mecânicas (rótulos, `outline-none`, `onClick` em `div/span`, `img`, `autoFocus`, `user-scalable`, `onPaste`, placeholders, `autocomplete`/`name`, `spellCheck`, `transition`, `tap-highlight`, `aria-live`) e leitura dos trechos apontados.

### Achados sem correção e sem conflito registrado (reprovam)

```text
## features/estoque/produtos/create-produto-dialog.tsx
features/estoque/produtos/create-produto-dialog.tsx:114 - campo "Código" sem spellCheck={false} (regra "Disable spellcheck on … codes"; também DSP-04 AC7)

## components/app/confirm-dialog.tsx
components/app/confirm-dialog.tsx:149 - textarea de motivo sem name nem autocomplete

## features/estoque/ajustes/ajuste-dialog.tsx
features/estoque/ajustes/ajuste-dialog.tsx:107 - textarea de motivo sem name nem autocomplete
features/estoque/ajustes/ajuste-dialog.tsx:118 - prévia "Saldo resultante" fora de região aria-live (a do valor líquido tem: lancamento-form.tsx:257)

## features/identity/usuarios/dialogs/admin-dialogs.tsx
features/identity/usuarios/dialogs/admin-dialogs.tsx:139 - textarea de motivo sem name nem autocomplete

## features/identity/usuarios/dialogs/temporary-password-dialog.tsx
features/identity/usuarios/dialogs/temporary-password-dialog.tsx:116 - textarea de motivo sem name nem autocomplete

## features/estoque/produtos/produtos-page.tsx
features/estoque/produtos/produtos-page.tsx:93 - busca sem name nem autocomplete="off"

## features/financeiro/lancamentos/lancamentos-page.tsx
features/financeiro/lancamentos/lancamentos-page.tsx:209 - filtro (select) sem name nem autocomplete
features/financeiro/lancamentos/lancamentos-page.tsx:220 - idem
features/financeiro/lancamentos/lancamentos-page.tsx:234 - idem
features/financeiro/lancamentos/lancamentos-page.tsx:245 - data "Criado de" sem name nem autocomplete
features/financeiro/lancamentos/lancamentos-page.tsx:249 - data "Criado até" sem name nem autocomplete

## features/identity/usuarios/users-page.tsx
features/identity/usuarios/users-page.tsx:156 - filtro (select) sem name nem autocomplete
features/identity/usuarios/users-page.tsx:170 - idem

## components/app/numeric-text-field.tsx
components/app/numeric-text-field.tsx:79-86 - controle composto (prefixo "R$" + campo) sem :focus-within; o anel fica só no input

## app/globals.css
app/globals.css - nenhum -webkit-tap-highlight-color definido (regra "set intentionally")

## components/ui/button.tsx
components/ui/button.tsx:12 - whitespace-nowrap: rótulo longo não quebra (texto longo sem tratamento)

## app/(dev)/design-system
app/(dev)/design-system/acesso/example-login-form.tsx:52 - placeholder sem "…"
app/(dev)/design-system/example-listing.tsx:58 - placeholder sem "…"
app/(dev)/design-system/fields-section.tsx:101 - placeholder sem "…"
app/(dev)/design-system/fields-section.tsx:138 - placeholder sem "…"
```

Vários deles (motivo sem `name`, `focus-within` do campo de dinheiro, `tap-highlight`, rótulo longo do botão) estão em `STATE.md:288` como **pendências** "que não bloqueiam", não como conflito com decisão do projeto. O critério de T14 (`tasks/05-migracao-web-v1.md:450`) e DSM-06 AC2 exigem correção ou conflito registrado; pendência não basta.

### Achados casados com conflito ou decisão registrada (não reprovam)

| Achado | Onde | Registro |
| --- | --- | --- |
| Title Case | todo o front | `STATE.md:187` (WEB-D-012) |
| Botão de envio desabilitado sem motivo | `confirm-dialog.tsx` | `STATE.md:188` (FND-05 AC6) |
| Estado na URL sem `nuqs` | `lib/url-state` | `STATE.md:189` (WDS-D-011) |
| Listas > 50 sem virtualização | `lancamentos-page.tsx:103`, `produtos-page.tsx:155` | `STATE.md:190` (WDS-D-012) |
| Regras de dark mode | — | `STATE.md:191` (WDS-D-003) |
| `env(safe-area-inset-*)` sem `viewportFit: cover` | `app/layout.tsx` | `STATE.md:282` |
| Placeholder `0,00` sem "…" | `components/app/money-field.tsx:41` | `STATE.md:283` |
| Toast anima `height` | `components/ui/toast.tsx:73-77` | `STATE.md:284` |
| Data de calendário sem `Intl` | `components/app/format.ts` | `STATE.md:285` |
| Anel de foco também no hover do item de menu | `components/ui/dropdown-menu.tsx` | `STATE.md:286` |
| Toast exposto como `role="dialog"` (D5) | `components/ui/toast.tsx` | `STATE.md:293` |
| Toast de erro `polite` | `components/ui/toast.tsx` | `STATE.md:298` |
| Sem aviso de alterações não salvas | formulários em diálogo | `STATE.md:256` (fora do ciclo) |

### Conferido e sem achado

`outline-none` só em alvos de foco programático com `tabIndex={-1}` (`app/error-heading.tsx:26`, `components/app/page-container.tsx:24`, `access-layout.tsx:31`, `shell-frame.tsx:200`), não interativos; nenhum `onClick` em `div`/`span`/`li`/`p`; a única imagem (`components/app/brand.tsx:56`) declara largura, altura e prioridade; nenhum `autoFocus`, `user-scalable`, `maximum-scale` ou `onPaste`; nenhum `...` literal; `transition-all` barrado pela guarda; `<select>` nativo com cores explícitas (`field.test.tsx:325`); `touch-manipulation` em botão, link da trilha, marca e "Sem acesso".

---

## SSR e hidratação

- `useSearchParams` só em componentes cliente e sempre sob `<Suspense>`: `app/(app)/admin/usuarios/page.tsx`, `conta/senha/page.tsx`, `estoque/produtos/page.tsx`, `financeiro/lancamentos/page.tsx` (ex.: `financeiro/lancamentos/page.tsx:18-20`), `app/(public)/entrar/page.tsx`; `lib/url-state/use-query-state.ts:46` documenta a exigência.
- `useMediaQuery` (`lib/responsive/use-media-query.ts:45-63`) usa `useSyncExternalStore` com `getServerSnapshot` fixo no desktop: HTML do servidor e primeira renderização do cliente coincidem. Testes: `use-media-query.test.tsx:49,53,59,89,107`, `use-media-query.ssr.test.tsx:15`, `data-list.test.tsx:426,433,457`.
- `window`, `Date.now`, `Math.random`, `new Date()` em renderização: nenhum. Leituras de `window` só em efeitos e manipuladores (`use-query-state.ts:27,66-77`, `reset-password-form.tsx:52-54`, `comprovantes-panel.tsx:55`, `session-provider.tsx:86,150`).
- `useQueryState` chama `setDraft(null)` durante a renderização (`use-query-state.ts:51-52`), padrão de estado derivado aceito pelo React; sem efeito colateral externo.
- No navegador: `evidence/smoke-visual.md:142-150` e Smoke T14 sem nenhum aviso de hidratação (busca por `hydrat|#418|#423|#425|did not match`: 0), 0 exceções, 0 avisos.

---

## Propriedade e escopo

- `git diff --name-only 7c30c6a f9de714` fora de `web/` e `.specs/`: `.gitignore` e `docs/development/web.md`. Em `.specs/`, só `.specs/features/web-design/`.
- Sem mudança: `api/`, `web/package.json`, `web/pnpm-lock.yaml`, `web/lib/api/`, `web/lib/session/session-provider.tsx`, `features/*/hooks.ts` (8 arquivos; diff vazio).
- `web/components/app/require-permission.tsx`: só a aparência de `NoAccess` mudou (linhas 35-56 do diff); `RequirePermission`, `useCan` e a decisão por permissão estão intactos.
- `web/lib/session/navigation-mock.ts`: arquivo de teste de DS-T, permitido (`STATE.md:133`).
- **Desvio não registrado**: `web/lib/forms/use-focus-invalid-after-submit.ts` (e o teste) foi criado em `2903f32`. `design/05-migracao-web-v1.md:212` põe `lib/forms/` "sem dono e sem alteração neste ciclo"; `tasks/05-migracao-web-v1.md:37` repete; não há registro em `STATE.md`. É um arquivo novo, sem mudar `applyProblemToForm` nem os catálogos, mas fere a regra de propriedade.
- `features/financeiro/comprovantes/errors.ts:3,7`: só troca o formatador do limite (DSP-07 AC6); mensagens da API iguais.

---

## Smoke (amostra do verificador, build de produção em `:3000`, API em `:8080`)

Edge headless dirigido por CDP, scripts próprios em `…/scratchpad/verify/smoke/` (sobre `t13/lib.mjs`), perfil temporário apagado com o processo. Senhas lidas do arquivo fora do repositório só em memória; nada foi impresso nem gravado. Sem capturas novas no repositório.

| Passo | Resultado |
| --- | --- |
| Login presidente | `/inicio`, título "Início \| TJ Platform" |
| Criar lançamento (receita R$ 31,00, Mensalidades, Pix) | aviso "Lançamento criado."; primeira linha "13/10/2026 Receita Mensalidades R$ 31,00 R$ 0,00 R$ 31,00 Pix Criada" |
| Receber | `alertdialog` "Receber esta receita?"; aviso "Recebimento registrado."; status "Recebida"; foco no `h1` (correção N3) |
| Anexar comprovante (PDF de 191 bytes) | aviso "Comprovante anexado."; linha `t14-comprovante.pdf PDF 191 bytes 1 você` |
| Navegação 1440 / 1024 | "Navegação principal" visível, botão de menu oculto, "Lançamentos" com `aria-current` |
| Navegação 820 / 768 / 390 | sem navegação fixa; "Abrir menu" 44 × 44 |
| Gaveta 390 | `dialog` "Menu" com 6 links; escolher "Saldo" fecha a gaveta |
| Gaveta 768, Esc | foco volta a "Abrir menu" |
| Rolagem horizontal 390, 1024, 1440 em lançamentos, usuários e produtos | `scrollWidth === clientWidth` nas 9 medições (ex.: 375/375 em 390 com barra vertical; 1009/1009 em 1024) |
| Conselho Fiscal | login → `/inicio`; navegação sem "Usuários"; início, lançamentos, contas, saldo, produtos e detalhe do lançamento com 0 botões, 0 campos de arquivo, 0 formulários em `main`; `/admin/usuarios` → "Sem acesso" com "Voltar ao início" |
| `/design-system` em produção | 404, "Página não encontrada" |
| Console | 0 avisos de hidratação, 0 exceções, 0 avisos; erros só 401 de `/auth/me` sem sessão e o 404 esperado |

Observação: a primeira rodada do script mediu a navegação em 1440 antes de a página carregar (lista vazia) e não achou "Receber" porque abriu o primeiro link da tabela; a segunda rodada (`v2.mjs`), esperando a carga e abrindo a linha de R$ 31,00, deu os resultados acima. Os dados criados (receita de R$ 31,00 recebida, com `t14-comprovante.pdf`) ficam no banco local.

---

## Code Quality

| Principle | Status |
| --- | --- |
| Minimum code / surgical changes | ✅ (alias removidos, guardas com exceções justificadas) |
| No scope creep | ⚠️ `lib/forms` novo fora da propriedade |
| Matches patterns | ✅ |
| Spec-anchored outcome check | ⚠️ 13 lacunas de precisão, 3 ACs ❌ |
| Per-layer coverage | ✅ kit e compostos 1:1; ❌ DSP-04 AC6/AC7 sem teste |
| Every test maps to a requirement | ✅ (testes citam o AC no cabeçalho) |
| Documented guidelines followed | `CLAUDE.md`, `web/AGENTS.md`, `docs/development/web.md`, `web-design-guidelines` — ❌ achados acima |

---

## Lacunas ordenadas (com unidade dona)

1. **DSP-04 AC7 / `web-design-guidelines`** — campo "Código" do produto sem `spellCheck={false}` (`features/estoque/produtos/create-produto-dialog.tsx:114`), sem teste. Dono: DS-ES (via DS-INT). Correção: `spellCheck={false}` e teste em `create-produto-dialog`.
2. **DSP-04 AC6 / `web-design-guidelines`** — campos sem `autocomplete` (e textareas de motivo sem `name`): `confirm-dialog.tsx:149` (DS-A), `ajuste-dialog.tsx:107` (DS-ES), `admin-dialogs.tsx:139` e `temporary-password-dialog.tsx:116` (DS-ID), `produtos-page.tsx:93` (DS-ES), `lancamentos-page.tsx:209,220,234,245,249` (DS-FB), `users-page.tsx:156,170` (DS-ID). Nenhum teste cobre o AC. Correção: `autoComplete="off"` e `name` nos campos não de autenticação; teste por tela ou guarda.
3. **DSM-06 AC2 / gate `web-design-guidelines`** — demais achados sem correção nem conflito: `numeric-text-field.tsx:79-86` sem `focus-within` (DS-A); `-webkit-tap-highlight-color` ausente em `app/globals.css` (DS-F); `button.tsx:12` rótulo longo (DS-K); prévia "Saldo resultante" sem `aria-live` em `ajuste-dialog.tsx:118` (DS-ES); placeholders sem "…" no espécime (DS-E). Dono da decisão: coordenador (corrigir ou registrar cada um como conflito em `STATE.md`).
4. **Sensor M13 sobreviveu** — `lancamentos-page.test.tsx:471-475` não afirma a ausência do `ShowMore` inteiro com exatamente 50. Dono: DS-FB. Correção: afirmar que `[data-slot=show-more]` (ou o texto "Mostrando…") não existe com 50 registros; ou não renderizar o bloco pela guarda da tela.
5. **Propriedade** — `web/lib/forms/use-focus-invalid-after-submit.ts` criado em `2903f32` contra `design/05-migracao-web-v1.md:212`, sem registro. Dono: coordenador (registrar em `STATE.md` ou mover para `components/app`).
6. **DSM-01 AC2** — `.gitignore` alterado (`138d094`), fora da lista do AC. Dono: coordenador (registrar como exceção ou ajustar o AC).
7. **Lacunas de precisão da spec** (dono: autor da spec, via `STATE.md`):
   - DSF-03 AC4 não lista o nome da marca, mas o design aprovado e `data.test.tsx:390-394` põem `DisplayTitle` em caixa alta.
   - DSF-04 AC8 ("dois pontos de quebra") contradiz DSC-04 AC4 (640 px) e o uso de `min-[400px]` no shell.
   - DSC-05 AC2 pede o tom "informação"; WDS-D-018 o fundiu no neutro, e o kit tem 4 tons.
   - DSC-01 AC1, DSC-05 AC5, DSP-02 AC6 e DSP-07 AC1 a AC3 não têm valor verificável; DSP-07 AC2 é contrariado pelo link "Detalhes".
   - DSC-04 AC5 e DSP-01 AC3 só têm evidência por inspeção ou numa única tela; DSP-04 AC2 só tem teste nos campos numéricos.

---

## Requirement Traceability Update

| Requirement | New Status |
| --- | --- |
| DSF-01, DSF-02, DSF-05, DSF-06 | ✅ Verified |
| DSF-03, DSF-04 | ✅ Verified (⚠️ precisão: AC4, AC8) |
| DSC-01 a DSC-07 | ✅ Verified (⚠️ DSC-01 AC1, DSC-04 AC5, DSC-05 AC2/AC5) |
| DSP-01, DSP-02, DSP-03, DSP-06, DSP-07 | ✅ Verified (⚠️ indicados) |
| DSP-04 | ❌ Needs Fix (AC6, AC7) |
| DSP-05 | ⚠️ Needs Fix (teste da borda de 50, M13) |
| DSS-01 a DSS-06 | ✅ Verified (FND-05 AC2 revisado reconfirmado) |
| DSM-01 a DSM-05 | ✅ Verified (⚠️ DSM-01 AC2) |
| DSM-06 | ❌ Needs Fix (AC2) |

---

## Summary

**Overall**: ❌ Not Ready

**Spec-anchored check**: 203/219 ACs com o valor da spec; 3 ❌; 13 lacunas de precisão
**Sensor**: 16 injetadas, 15 mortas, 1 sobreviveu (M13)
**Gate**: lint, typegen, tsc, test (1774/1774), build, contrato e audit com saída 0; `api/` e dependências intactos

**What works**: fundação de tokens, contraste e marca com guardas fortes; kit e compostos com testes de papel, nome e teclado; estado na URL e SSR seguros; FND-05 AC2 revisado (gaveta abaixo de 1024, fixa a partir de 1024) confirmado por teste, mutante e navegador; permissões e regras de negócio intocadas; visão do Conselho Fiscal sem escrita.

**Next steps**: corrigir as lacunas 1 a 4 (ou registrar como conflito o que for decisão), registrar 5 e 6, e nova verificação (iteração 1 de 3).

---

## Fechamento determinístico

`python .claude/skills/tlc-spec-driven/scripts/validate_state.py web-design --root C:/Users/Michels/Desktop/torcida-jovem` → exit 1: "ERROR web-design: validation.md verdict is FAIL - route the ranked gaps to fix tasks, then re-verify (feature is not done)". Coerente com o veredito: a feature não está concluída.
