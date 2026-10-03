# Permissões — Especificação

Status: Aprovada em 2026-09-29. Sub-spec 5 de 5 da feature `financeiro`. Decisões compartilhadas em `.specs/features/financeiro/STATE.md` (`FIN-D-NNN`), matriz consolidada lá. Narrativa em português; critérios de aceite em inglês (EARS).

## Problem Statement

`financeiro` precisa publicar seu catálogo de permissões e a matriz de concessão por papel para `identity.app.BuildMatrix` (mesmo mecanismo já usado por `identity` e `audit`), preservando as 3 permissões institucionais já reservadas ao Conselho Fiscal desde a fundação.

## Goals

- Publicar a `Contribution{Module: "financeiro"}` com as 13 permissões operacionais do V1 (`01`-`04`).
- Preservar, sem alteração, as 3 permissões institucionais já existentes.
- Distribuir as permissões pelos papéis conforme a matriz de `financeiro/STATE.md`.

## Out of Scope

| Item | Razão |
|---|---|
| Qualquer caso de uso que verifique `financeiro:prestacao_contas:*` ou `financeiro:parecer:opine` | `FIN-D-014` — preparação para uma sub-spec futura de prestação de contas |
| A lógica de cada operação (`create`, `receive`, etc.) | Pertence a `01`-`04` — esta spec só declara e distribui os nomes |
| Alterar a matriz de `identity`/`audit` | Fora do escopo — só adiciona uma nova `Contribution` |

## Assumptions & Open Questions

| Question | Chosen default | Rationale | Confirmed |
|---|---|---|---|
| Formato | `financeiro:recurso:ação`, mesma regex do catálogo (`^[a-z_]+:[a-z_]+:[a-z_]+$`) | `platform/authz`, já existente | y |
| `CommonRead` | Nunca, em nenhuma permissão de `financeiro` | Comentário em `authz/permission.go`: "administration, security, RBAC, financial and institutional permissions are always audited" | y |
| Conteúdo do catálogo | Preenchido incrementalmente por `01`-`04`, cada uma na sua própria fase de tarefas — esta spec define a regra, não a lista fechada de antemão | Decisão da revisão de ordem de implementação | y |
| Permissões institucionais | Preservadas exatamente como estão hoje em `identity/app/roles_matrix.go` (`FoundationContributions`), sem nenhum caso de uso novo | `FIN-D-014` | y |
| Distribuição por papel | Conforme a matriz consolidada em `financeiro/STATE.md` | `FIN-001` §24, decisões da rodada de revisão (`FIN-D-010` a `FIN-D-013`) | y |

Open questions: none.

## User Stories

### PERM-01: Catálogo e regras do módulo

**User Story**: Como mantenedor da plataforma, quero que o catálogo de permissões de `financeiro` siga as mesmas regras já aplicadas aos demais módulos, para que a matriz agregada continue consistente.

**Acceptance Criteria**:
1. WHEN `financeiro` contributes a permission THEN its name SHALL match the format `financeiro:recurso:ação`.
2. The system SHALL NOT declare any `financeiro` permission as `CommonRead`.
3. The system SHALL NOT grant `CONSELHO_FISCAL` any `financeiro` permission whose action is `create`, `update`, `deactivate`, `receive`, `pay` or `cancel`. `BuildMatrix` already refuses such a grant for all six (`forbiddenToConselhoFiscal`, `api/internal/identity/app/roles_matrix.go` — hardened for `deactivate`/`receive`/`pay` by `fix(identity): reforca a invariante do Conselho Fiscal no RBAC`, PR #40, mesclado em `develop` antes desta sub-spec ser implementada). Esta sub-spec's own test still asserts it directly against `financeiro`'s real `Contribution` — belt-and-suspenders, not reliance on the central mechanism alone.
4. WHEN `financeiro`'s `Contribution` is aggregated by `BuildMatrix` THEN `PRESIDENTE` SHALL receive every `financeiro` permission automatically, as an explicit grant created by the synchronization.

**Independent Test**: montar a `Contribution` de `financeiro` com as 13 permissões do V1, agregar com `BuildMatrix` junto das demais, e ver a matriz resultante sem erro.

---

### PERM-02: Preservação das permissões institucionais

**User Story**: Como Conselho Fiscal, quero continuar com as permissões de prestação de contas que já possuo, mesmo depois que `financeiro` substituir o placeholder da fundação, para não perder acesso que já tinha.

**Acceptance Criteria**:
1. WHEN `financeiro` publishes its own `Contribution`, replacing the foundation's placeholder THEN it SHALL still declare `financeiro:prestacao_contas:read`, `financeiro:prestacao_contas:approve` and `financeiro:parecer:opine`, granted to `CONSELHO_FISCAL`, unchanged.
2. The system SHALL NOT introduce any use case or HTTP endpoint that checks these three permissions in this round.
3. WHEN the matrix is rebuilt with `financeiro`'s real `Contribution` in place of the placeholder THEN `CONSELHO_FISCAL`'s effective permissions for these three names SHALL be identical to before the replacement.

**Independent Test**: comparar a matriz agregada antes (placeholder da fundação) e depois (Contribution real de `financeiro`) e ver essas 3 permissões idênticas para `CONSELHO_FISCAL`.

---

### PERM-03: Distribuição por papel

**User Story**: Como mantenedor, quero que cada papel receba exatamente as permissões operacionais que `FIN-001` e as decisões fechadas preveem, para que ninguém tenha acesso a mais ou a menos do que deveria.

**Acceptance Criteria**:
1. WHEN the matrix is built THEN `TESOURARIA` SHALL hold all 13 operational permissions of the V1 (conta create/update/deactivate/read; lançamento create/update/read/receive/pay/cancel; saldo read; comprovante create/read).
2. WHEN the matrix is built THEN `DIRETORIA` SHALL hold only the read permissions: `financeiro:conta:read`, `financeiro:lancamento:read`, `financeiro:saldo:read`, `financeiro:comprovante:read`.
3. WHEN the matrix is built THEN `CONSELHO_FISCAL` SHALL hold the same four read permissions as `DIRETORIA`, plus the three institutional ones (PERM-02), and none of the write/transition permissions.
4. WHEN the matrix is built THEN `ASSOCIADO`, `ESTOQUE_LOJA`, `EVENTOS` and `ADMIN_SISTEMA` SHALL hold no `financeiro` permission.

**Independent Test**: montar a matriz completa e verificar, papel a papel, exatamente as permissões da tabela consolidada em `financeiro/STATE.md`.

## Edge Cases

- Tentar declarar uma permissão fora do prefixo `financeiro:`: recusado por `BuildMatrix` (`belongsTo`), já garantido pelo mecanismo existente, não uma regra nova desta spec.
- Tentar conceder `CONSELHO_FISCAL` `create`, `update`, `deactivate`, `receive`, `pay` ou `cancel`: recusado por `BuildMatrix` (`forbiddenToConselhoFiscal`), já garantido pelo mecanismo existente — as três últimas passaram a ser cobertas pelo PR #40, mesclado especificamente para fechar esta lacuna antes da implementação desta sub-spec.

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| PERM-01 | Catálogo e regras do módulo | Done | Done (T2) |
| PERM-02 | Preservação das permissões institucionais | Done | Done (T1, T3) |
| PERM-03 | Distribuição por papel | Done | Done (T2) |

**Coverage:** 3 total, 3 mapped to tasks, 0 unmapped.

## Success Criteria

- A matriz final não perde nenhuma permissão do Conselho Fiscal já existente.
- Nenhum papel além dos listados na seção PERM-03 recebe qualquer permissão de `financeiro`.
