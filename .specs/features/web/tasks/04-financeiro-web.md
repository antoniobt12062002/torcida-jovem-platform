# Financeiro (Web) — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos ACs; um commit atômico por task. **Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/web/spec/04-financeiro-web.md`
**Design**: `.specs/features/web/design/04-financeiro-web.md`
**Status**: Draft (2026-10-04), aguardando autorização de execução.

**Pré-requisito**: sub-spec 01 completa (F5 integrada).

**Unidades de execução**:

| Unidade | Tasks | Arquivos exclusivos | Paralelismo |
| --- | --- | --- | --- |
| FIN-a | T1, T2 | `web/features/financeiro/nav.ts`, `web/features/financeiro/contas/`, `web/features/financeiro/saldo/`, `web/app/(app)/financeiro/contas/`, `web/app/(app)/financeiro/saldo/` | paralela a FIN-b, ID-ADM, EST |
| FIN-b | T3, T4, T5 | `web/features/financeiro/lancamentos/`, `web/features/financeiro/comprovantes/`, `web/app/(app)/financeiro/lancamentos/` | paralela a FIN-a, ID-ADM, EST |

FIN-a não altera arquivos de FIN-b, e vice-versa. As duas frentes não dependem de estado produzido pela outra: a consulta de contas e a invalidação do saldo passam só por `lib/api/query-keys.ts` (F1).

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| Funções puras (`tree.ts`, `filters.ts`, `actions.ts`) | unit | todos os ramos | `web/features/financeiro/**/*.test.ts` | `pnpm test` |
| Telas e diálogos | component com MSW | ACs de FWB-01 a FWB-05, incluindo visibilidade por permissão | `web/features/financeiro/**/*.test.tsx` | `pnpm test` |

## Gate Check Commands

A partir de `web/`, com código de saída capturado de verdade.

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Audit | T2 e T5 (fim de cada unidade) | `pnpm audit --audit-level high` |
| Contract | toda task | `pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Ownership | toda task | diff só nos caminhos da unidade (tabela acima) |

---

## Execution Plan

```
T1 → T2
T3 → T4
T4 → T5
```

T1-T2 (FIN-a) e T3-T5 (FIN-b) rodam em paralelo, em agentes e worktrees separados.

---

## Task Breakdown

### T1: Plano de contas e navegação do módulo

**What**: `nav.ts` do financeiro, hooks, `tree.ts`, catálogo de erros, página de contas com criar, criar subconta, renomear e desativar.
**Where**: `web/features/financeiro/nav.ts`, `web/features/financeiro/contas/`, `web/app/(app)/financeiro/contas/`
**Depends on**: None (pré-requisito externo: sub-spec 01 completa)
**Reuses**: `components/app`, `lib/api/query-keys.ts`.
**Requirement**: FWB-01

**Done when**:
- [x] Árvore com contas inativas marcadas; subconta herda o tipo do pai
- [x] `conta_ja_utilizada`, `conta_nao_encontrada` e `conta_tipo_incompativel` com mensagem
- [x] Ações só com a permissão correspondente
- [x] Gates passam

**Tests**: unit (`tree.ts`) + component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): plano de contas do financeiro`

---

### T2: Saldo

**What**: página de saldo.
**Where**: `web/features/financeiro/saldo/`, `web/app/(app)/financeiro/saldo/`
**Depends on**: T1 (mesma unidade; sequência do agente FIN-a)
**Reuses**: `format.ts`, `lib/money.ts`.
**Requirement**: FWB-02

**Done when**:
- [x] Positivo, zero e negativo formatados com `formatBRL`
- [x] Gates passam (incluindo Audit)

**Tests**: component com MSW
**Gate**: Web + Audit + Contract + Backend intacto + Ownership
**Commit**: `feat(web): saldo do financeiro`

---

### T3: Lançamentos: lista, filtros, criação, edição e detalhe

**What**: hooks (com `useContasOptions` próprio), `filters.ts`, `actions.ts` (ações de edição), formulário de lançamento com prévia do líquido, lista e detalhe.
**Where**: `web/features/financeiro/lancamentos/`, `web/app/(app)/financeiro/lancamentos/`
**Depends on**: None (pré-requisito externo: sub-spec 01 completa)
**Reuses**: `MoneyField`, `StatusBadge`, `format.ts` (`authorLabel`).
**Requirement**: FWB-03

**Done when**:
- [x] Valores enviados em centavos inteiros
- [x] Editar só em `CRIADA`; `lancamento_imutavel` explicado
- [x] Filtros e ordenação no cliente cobertos por teste unitário
- [x] Detalhe de id inexistente mostra "Lançamento não encontrado"
- [x] Gates passam

**Tests**: unit (`filters.ts`, `actions.ts`) + component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): lancamentos do financeiro`

---

### T4: Workflow, cancelamento e devolução

**What**: receber, pagar, cancelar com motivo, devolver; `actions.ts` completo; invalidação de lista e saldo.
**Where**: `web/features/financeiro/lancamentos/`
**Depends on**: T3
**Reuses**: `ConfirmDialog` com motivo obrigatório.
**Requirement**: FWB-04

**Done when**:
- [x] Cada ação aparece só no status e com a permissão certos (`actions.ts` testado em todas as combinações)
- [x] Cancelamento com motivo vazio ou só espaços não chama a API
- [x] Mudança de status invalida `financeiro.lancamentos` e `financeiro.saldo`
- [x] Gates passam

**Tests**: unit (`actions.ts`) + component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): workflow, cancelamento e devolucao de lancamentos`

---

### T5: Comprovantes

**What**: painel de comprovantes no detalhe: listar, anexar (recusa acima de 10.420.224 bytes antes do envio), baixar pela URL assinada.
**Where**: `web/features/financeiro/comprovantes/`, `web/app/(app)/financeiro/lancamentos/[id]/`
**Depends on**: T4 (o painel entra na página de detalhe, já alterada pelas tasks anteriores da unidade)
**Reuses**: `lib/api/client.ts` (`bodySerializer` com `FormData`).
**Requirement**: FWB-05

**Done when**:
- [x] Upload como `multipart/form-data` no campo `file`, com `X-CSRF-Token`
- [x] Limite efetivo do cliente (10.420.224 bytes = 10 MiB − 64 KiB) testado: abaixo aceita, exatamente no limite aceita, um byte acima recusa sem requisição (MSW confirma zero chamadas)
- [x] `413 document_too_large` devolvido pela API continua mostrado com a mensagem correta (a API segue protegendo o limite, coberto no backend por `api/internal/platform/documents`)
- [x] URL assinada buscada no clique e nunca no cache
- [x] Erros `413` e `422` de documento com mensagem
- [x] Gates passam (incluindo Audit)

**Tests**: component com MSW
**Gate**: Web + Audit + Contract + Backend intacto + Ownership
**Commit**: `feat(web): comprovantes de lancamentos`

---

## Phase Execution Map

```
T1 → T2
T3 → T4
T4 → T5
```

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | 1 recurso (contas) + item de navegação | ✅ Granular |
| T2 | 1 página | ✅ Granular |
| T3 | leitura e escrita básica de lançamentos | ✅ Aceitável |
| T4 | transições de status | ✅ Granular |
| T5 | 1 painel | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | None | None | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | T4 | T4 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | feature | unit + component | unit + component | ✅ OK |
| T2 | feature | component | component | ✅ OK |
| T3 | feature | unit + component | unit + component | ✅ OK |
| T4 | feature | unit + component | unit + component | ✅ OK |
| T5 | feature | component | component | ✅ OK |
