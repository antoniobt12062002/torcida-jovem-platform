# Identity — Acesso e Senha (Web) — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos ACs; um commit atômico por task. **Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/web/spec/02-identity-acesso.md`
**Design**: `.specs/features/web/design/02-identity-acesso.md`
**Status**: Draft (2026-10-04), aguardando autorização de execução.

**Pré-requisito**: unidade F2 (sub-spec 01, T5-T6) integrada.

**Unidade de execução**: F3 (T1, T2), paralela a F5 (sub-spec 01, T7-T8). As duas não compartilham arquivo: F3 só escreve em `features/identity/acesso/` e nas três rotas abaixo; F5 só em `components/app/` e `app/(app)/layout.tsx`. F3 não depende de F5: usa só `components/ui` e `lib/session`.

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `features/identity/acesso/` | component com MSW + unit | todos os ACs de ACS-01 a ACS-03 | `web/features/identity/acesso/*.test.ts(x)` | `pnpm test` |
| Rotas | component | página monta o formulário certo | `web/app/**/senha/*.test.tsx`, `web/app/(public)/**/*.test.tsx` | `pnpm test` |

## Gate Check Commands

A partir de `web/`, com código de saída capturado de verdade.

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Audit | T2 (fim da unidade) | `pnpm audit --audit-level high` |
| Contract | toda task | `pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Ownership | toda task | diff só em `web/features/identity/acesso/`, `web/app/(app)/conta/senha/`, `web/app/(public)/recuperar-acesso/`, `web/app/(public)/redefinir-senha/` |

---

## Execution Plan

```
T1 → T2
```

---

## Task Breakdown

### T1: Troca de senha (obrigatória e voluntária)

**What**: catálogo de erros da feature, formulário e mutação de troca, página `/conta/senha`.
**Where**: `web/features/identity/acesso/`, `web/app/(app)/conta/senha/`
**Depends on**: None (pré-requisito externo: F2 integrada)
**Reuses**: `lib/session`, `lib/forms`, `components/ui`.
**Requirement**: ACS-01

**Done when**:
- [x] Todos os ACs de ACS-01 cobertos por teste
- [x] Depois da troca obrigatória, `queryKeys.session.me` é invalidada e a pessoa segue para o destino
- [x] Gates passam

**Tests**: component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): troca de senha obrigatoria e voluntaria`

---

### T2: Recuperação de acesso e redefinição pelo link

**What**: pedido de recuperação, leitura do token do fragmento, redefinição; páginas `/recuperar-acesso` e `/redefinir-senha`.
**Where**: `web/features/identity/acesso/`, `web/app/(public)/recuperar-acesso/`, `web/app/(public)/redefinir-senha/`
**Depends on**: T1 (reusa o catálogo de erros criado em T1)
**Reuses**: `features/identity/acesso/errors.ts`.
**Requirement**: ACS-02, ACS-03

**Done when**:
- [ ] Mesma mensagem para qualquer e-mail
- [ ] Fragmento removido da barra de endereço; token nunca em query string nem em armazenamento (teste confere)
- [ ] `invalid_reset_token` oferece pedir outro link
- [ ] Gates passam (incluindo Audit)

**Tests**: unit (`read-reset-token`) + component com MSW
**Gate**: Web + Audit + Contract + Backend intacto + Ownership
**Commit**: `feat(web): recuperacao de acesso e redefinicao de senha pelo link`

---

## Phase Execution Map

```
T1 → T2
```

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | 1 fluxo, 1 rota | ✅ Granular |
| T2 | 2 fluxos encadeados (pedido → link), 2 rotas | ✅ Aceitável |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | feature + rota | component com MSW | component com MSW | ✅ OK |
| T2 | feature + rotas | unit + component | unit + component | ✅ OK |
