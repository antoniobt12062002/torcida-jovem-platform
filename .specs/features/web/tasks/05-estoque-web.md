# Estoque (Web) — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos ACs; um commit atômico por task. **Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/web/spec/05-estoque-web.md`
**Design**: `.specs/features/web/design/05-estoque-web.md`
**Status**: Draft (2026-10-04), aguardando autorização de execução.

**Pré-requisito**: sub-spec 01 completa (F5 integrada).

**Unidade de execução**: EST (T1-T3), paralela a ID-ADM, FIN-a e FIN-b. Só escreve em `web/features/estoque/` e `web/app/(app)/estoque/`. Não altera arquivos de Financeiro nem de Identity.

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| Funções puras (`search.ts`, `payload.ts`) | unit | todos os ramos | `web/features/estoque/**/*.test.ts` | `pnpm test` |
| Telas e diálogos | component com MSW | ACs de EWB-01 a EWB-04 | `web/features/estoque/**/*.test.tsx` | `pnpm test` |

## Gate Check Commands

A partir de `web/`, com código de saída capturado de verdade.

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Audit | T3 (fim da unidade) | `pnpm audit --audit-level high` |
| Contract | toda task | `pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Ownership | toda task | diff só em `web/features/estoque/` e `web/app/(app)/estoque/` |

---

## Execution Plan

```
T1 → T2
T2 → T3
```

---

## Task Breakdown

### T1: Produtos, detalhe e saldo

**What**: `nav.ts`, catálogo de erros, lista com busca, criação, detalhe com saldo e seções por permissão.
**Where**: `web/features/estoque/`, `web/app/(app)/estoque/`
**Depends on**: None (pré-requisito externo: sub-spec 01 completa)
**Reuses**: `components/app`, `lib/api/query-keys.ts`.
**Requirement**: EWB-01, EWB-04

**Done when**:
- [x] `codigo_duplicado` no campo de código; `validation_failed` nos campos
- [x] Saldo negativo destacado; seção omitida sem a permissão (sem chamar a API)
- [x] Gates passam

**Tests**: unit (`search.ts`) + component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): produtos e saldo do estoque`

---

### T2: Movimentações

**What**: histórico, entrada, saída e devolução a partir do histórico; `payload.ts` com `INVENTARIO`.
**Where**: `web/features/estoque/movimentacoes/`
**Depends on**: T1
**Reuses**: `QuantityField` (`positive`).
**Requirement**: EWB-02

**Done when**:
- [x] Corpo enviado sempre com `origem = INVENTARIO`; nenhuma origem reservada oferecida
- [x] "Devolver" só em `ENTRADA` e `SAIDA`
- [x] Histórico e saldo invalidados depois do registro
- [x] Gates passam

**Tests**: unit (`payload.ts`) + component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): movimentacoes de estoque`

---

### T3: Ajuste

**What**: diálogo de ajuste com quantidade com sinal, motivo obrigatório e prévia do saldo resultante.
**Where**: `web/features/estoque/ajustes/`
**Depends on**: T2 (entra no mesmo detalhe e reusa a invalidação de T2)
**Reuses**: `QuantityField` (`nonZero`), `ConfirmDialog` com motivo.
**Requirement**: EWB-03

**Done when**:
- [ ] Zero ou motivo vazio mantém a confirmação desabilitada
- [ ] Saldo resultante negativo avisado sem impedir
- [ ] Gates passam (incluindo Audit)

**Tests**: component com MSW
**Gate**: Web + Audit + Contract + Backend intacto + Ownership
**Commit**: `feat(web): ajuste de estoque`

---

## Phase Execution Map

```
T1 → T2
T2 → T3
```

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | produtos + detalhe | ✅ Aceitável |
| T2 | movimentações | ✅ Granular |
| T3 | 1 diálogo | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | T2 | T2 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | feature | unit + component | unit + component | ✅ OK |
| T2 | feature | unit + component | unit + component | ✅ OK |
| T3 | feature | component | component | ✅ OK |
