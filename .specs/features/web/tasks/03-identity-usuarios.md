# Identity — Administração de Usuários (Web) — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos ACs; um commit atômico por task. **Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/web/spec/03-identity-usuarios.md`
**Design**: `.specs/features/web/design/03-identity-usuarios.md`
**Status**: Draft (2026-10-04), aguardando autorização de execução.

**Pré-requisito**: sub-spec 01 completa (F5 integrada).

**Unidade de execução**: ID-ADM (T1-T3), paralela a FIN-a, FIN-b e EST. Só escreve em `web/features/identity/nav.ts`, `web/features/identity/usuarios/` e `web/app/(app)/admin/usuarios/`. Não é executada pelos agentes de Financeiro ou Estoque.

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `features/identity/usuarios/` | component com MSW | todos os ACs de USR-01 a USR-05, incluindo visibilidade por permissão | `web/features/identity/usuarios/**/*.test.tsx` | `pnpm test` |
| `features/identity/nav.ts` | unit | item e permissão | `web/features/identity/nav.test.ts` | `pnpm test` |

## Gate Check Commands

A partir de `web/`, com código de saída capturado de verdade.

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Audit | T3 (fim da unidade) | `pnpm audit --audit-level high` |
| Contract | toda task | `pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Ownership | toda task | diff só nos três caminhos da unidade |

---

## Execution Plan

```
T1 → T2
T2 → T3
```

---

## Task Breakdown

### T1: Listagem, filtros, paginação e item de navegação

**What**: `nav.ts`, hooks de consulta, `roles.ts`, `errors.ts`, página e tabela (sem ações de escrita ainda).
**Where**: `web/features/identity/nav.ts`, `web/features/identity/usuarios/`, `web/app/(app)/admin/usuarios/`
**Depends on**: None (pré-requisito externo: sub-spec 01 completa)
**Reuses**: `components/app`, `components/ui`, `lib/api/query-keys.ts`.
**Requirement**: USR-01

**Done when**:
- [x] Filtros `active` e `role` enviados à API; "Carregar mais" com `next_cursor`
- [x] Sem `identity:user:read`: "Sem acesso"
- [x] Gates passam

**Tests**: component com MSW + unit (`nav.ts`)
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): listagem de usuarios com filtros e paginacao`

---

### T2: Criar, desativar e reativar

**What**: diálogo de criação, diálogos de desativação e reativação, menu de ações da linha.
**Where**: `web/features/identity/usuarios/`
**Depends on**: T1
**Reuses**: `ConfirmDialog`, `lib/forms`.
**Requirement**: USR-02, USR-03

**Done when**:
- [x] Ações visíveis só com a permissão certa e nunca na própria linha
- [x] `email_taken` no campo de e-mail; `last_admin` e `self_change_forbidden` com mensagem
- [x] Gates passam

**Tests**: component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): criacao, desativacao e reativacao de usuarios`

---

### T3: Acesso administrativo, papéis e senha temporária

**What**: promover, editar papéis (com aviso de lista vazia), retirar acesso, senha temporária.
**Where**: `web/features/identity/usuarios/dialogs/`
**Depends on**: T2
**Reuses**: `roles.ts`, `errors.ts`, `ConfirmDialog` com motivo.
**Requirement**: USR-04, USR-05

**Done when**:
- [ ] Só a ação compatível com o estado do usuário aparece
- [ ] Motivo com menos de 10 caracteres úteis mantém o envio desabilitado na promoção e na senha temporária
- [ ] Senha temporária some do DOM ao fechar e nunca está no `QueryClient` nem em armazenamento
- [ ] Gates passam (incluindo Audit)

**Tests**: component com MSW
**Gate**: Web + Audit + Contract + Backend intacto + Ownership
**Commit**: `feat(web): acesso administrativo, papeis e senha temporaria`

---

## Phase Execution Map

```
T1 → T2
T2 → T3
```

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | leitura da feature | ✅ Granular |
| T2 | 3 ações simples de usuário | ✅ Granular |
| T3 | 4 ações de acesso, mesmo conjunto de diálogos | ✅ Aceitável |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | T2 | T2 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | feature + nav | component + unit | component + unit | ✅ OK |
| T2 | feature | component com MSW | component com MSW | ✅ OK |
| T3 | feature | component com MSW | component com MSW | ✅ OK |
