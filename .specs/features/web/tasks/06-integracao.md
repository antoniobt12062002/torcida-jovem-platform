# Integração e Verificação da Web V1 — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos ACs; um commit atômico por task. **Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/web/spec/06-integracao.md`
**Design**: `.specs/features/web/design/06-integracao.md`
**Status**: Draft (2026-10-04), aguardando autorização de execução.

**Pré-requisito**: unidades F3, ID-ADM, FIN-a, FIN-b e EST integradas (rodada 4 completa, com todos os gates verdes).

**Unidades de execução**: INT (T1-T3), sequencial e única; VERIFY (T4), agente independente, depois da INT.

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `nav-items.ts`, `/inicio`, `/` | unit + component com MSW | INT-01 com cinco perfis de sessão | `web/components/app/nav-items.test.ts`, `web/app/**/inicio/*.test.tsx` | `pnpm test` |
| Guia | revisão manual | INT-02 | `docs/development/web.md` | — |
| Smoke | aceite manual contra a stack real | INT-03 | `.specs/features/web/smoke.md` | — |
| Verificação | spec-anchored + sensor | INT-04 | `.specs/features/web/validation.md` | `validate_state.py web` |

## Gate Check Commands

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `cd web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Audit | T3 e T4 | `cd web && pnpm audit --audit-level high` |
| Contract | toda task | `cd web && pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Ownership | T1-T3 | diff só em `web/components/app/nav-items.ts`, `web/app/(app)/layout.tsx`, `web/app/(app)/inicio/`, `web/app/page.tsx`, `docs/development/web.md`, `.specs/features/web/smoke.md` |
| State | T4 | `python .claude/skills/tlc-spec-driven/scripts/validate_state.py web` |

---

## Execution Plan

```
T1 → T3
T2 → T3
T3 → T4
```

---

## Task Breakdown

### T1: Navegação global, página inicial e raiz

**What**: `nav-items.ts`, `AppShell` recebendo a lista real em `app/(app)/layout.tsx`, `/inicio`, redirecionamento de `/` (sem o ping de `/healthz`).
**Where**: `web/components/app/nav-items.ts`, `web/app/(app)/layout.tsx`, `web/app/(app)/inicio/`, `web/app/page.tsx`
**Depends on**: None (pré-requisito externo: rodada 4 integrada)
**Reuses**: `features/*/nav.ts`, `AppShell`, `visibleItems`.
**Requirement**: INT-01

**Done when**:
- [x] Os cinco perfis de sessão mostram exatamente os itens esperados
- [x] `/` leva a `/inicio`; conta sem área vê a mensagem
- [x] Gates passam

**Tests**: unit + component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): navegacao global, pagina inicial e redirecionamento da raiz`

---

### T2: Guia de desenvolvimento da Web

**What**: `docs/development/web.md`.
**Where**: `docs/development/web.md`
**Depends on**: None (usa a evidência de F1 já integrada)
**Reuses**: `.specs/features/web/evidence/f1-rewrite.md`, `CLAUDE.md`, `.env.example`.
**Requirement**: INT-02

**Done when**:
- [x] Variáveis, ordem de subida (incluindo o S3 local avulso), rewrite e gates documentados; comportamento de `API_URL` no build registrado
- [x] Nenhum segredo ou dado pessoal
- [x] Gates passam

**Tests**: revisão manual
**Gate**: Contract + Backend intacto + Ownership
**Commit**: `docs(web): guia de desenvolvimento da web`

---

### T3: Smoke manual contra a stack real

**What**: executar o checklist de INT-03 com banco, API, S3 local e front; registrar em `smoke.md`; abrir correção na frente dona para cada falha e repetir.
**Where**: `.specs/features/web/smoke.md`
**Depends on**: T1, T2
**Reuses**: guia de T2.
**Requirement**: INT-03

**Done when**:
- [x] Todos os itens com resultado, data e observação
- [x] Nenhuma falha pendente (falhas corrigidas pela frente dona e repetidas)
- [x] Gates passam

**Tests**: aceite manual contra a stack real
**Gate**: Web + Audit + Contract + Backend intacto + Ownership
**Commit**: `docs(web): resultado do smoke manual da web v1`

---

### T4: Verificação independente

**What**: agente VERIFY executa a estratégia do design e escreve `validation.md`; lacunas viram tasks de correção (no máximo 3 ciclos).
**Where**: `.specs/features/web/validation.md`
**Depends on**: T3
**Reuses**: `tlc-spec-driven` (`validate.md`, `lessons.py`).
**Requirement**: INT-04

**Done when**:
- [ ] Evidência `arquivo:linha` para cada AC das sub-specs 01 a 05 e INT-01 a INT-03
- [ ] As dez falhas mínimas do sensor mortas; árvore real igual à de antes do sensor
- [ ] `validate_state.py web` passa com veredito PASS

**Tests**: spec-anchored + sensor de discriminação
**Gate**: Web + Audit + Contract + Backend intacto + State
**Commit**: `docs(web): relatorio de verificacao independente da web v1`

---

## Phase Execution Map

```
T1 → T3
T2 → T3
T3 → T4
```

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | ligação das frentes, 4 arquivos pequenos | ✅ Aceitável |
| T2 | 1 documento | ✅ Granular |
| T3 | 1 checklist | ✅ Granular |
| T4 | 1 relatório | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | None | None | ✅ Match |
| T3 | T1, T2 | T1, T2 | ✅ Match |
| T4 | T3 | T3 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | navegação e rotas | unit + component | unit + component | ✅ OK |
| T2 | documentação | revisão manual | revisão manual | ✅ OK |
| T3 | evidência | aceite manual | aceite manual | ✅ OK |
| T4 | relatório | spec-anchored + sensor | spec-anchored + sensor | ✅ OK |
