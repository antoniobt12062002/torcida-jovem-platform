# Plano de Contas — Especificação

Status: Aprovada em 2026-09-29. Sub-spec 1 de 5 da feature `financeiro`. Decisões compartilhadas em `.specs/features/financeiro/STATE.md` — esta spec não as repete, só referencia (`FIN-D-NNN`). Narrativa em português; critérios de aceite em inglês (padrão EARS, exigido por `validate_spec.py`).

## Problem Statement

`financeiro` precisa de uma estrutura hierárquica de contas para classificar lançamentos (receita/despesa) de forma consistente ao longo do tempo, sem que uma classificação já usada historicamente mude de significado silenciosamente.

## Goals

- Criar contas contábeis, organizadas em hierarquia (categoria → subcategoria).
- Permitir renomear uma conta só enquanto ela nunca tiver sido usada em nenhum lançamento (`FIN-D-008`).
- Desativar contas sem nunca excluí-las.
- Consultar/listar o plano de contas.

## Out of Scope

| Item | Razão |
|---|---|
| Código contábil formal (ex. "1.1.01") | `FIN-D-009`: a hierarquia por `parent_id` é suficiente no V1 |
| Qualquer regra sobre `LancamentoFinanceiro` | Pertence a `02-lancamentos` |
| Exclusão definitiva de conta | Nunca prevista — só desativação |
| Snapshot do nome da conta no lançamento | `FIN-D-008`: como o nome é imutável após o primeiro uso, o nome exibido num lançamento histórico é sempre o nome atual e sempre correto — não há necessidade de copiar o nome no momento da criação |

## Assumptions & Open Questions

| Question | Chosen default | Rationale | Confirmed |
|---|---|---|---|
| Regra de renomeação | Permitida só antes do primeiro uso em um lançamento; depois, imutável | `FIN-D-008`, decisão do mantenedor | y |
| Como a regra de renomeação verifica "já foi usada" | Uma porta de leitura (`LancamentoExistenceChecker`) definida por esta spec, implementada consultando `LancamentoFinanceiro` (dono: `02-lancamentos`), sem escrita cruzada em nenhum sentido | `FIN-D-008`, Alternativa 1 aprovada na revisão arquitetural | y |
| Código contábil formal | Fora do V1 | `FIN-D-009` | y |
| Efeito da desativação sobre lançamentos históricos | Nenhum — a conta desativada continua referenciável por linhas antigas, só não pode ser escolhida num lançamento novo | Consequência mecânica de "desativar ≠ excluir"; sem controvérsia | y |
| Quem pode criar/renomear/desativar | `TESOURARIA` (e `PRESIDENTE`, automático) | `FIN-001` §24, matriz consolidada em `financeiro/STATE.md` | y |

Open questions: none.

## User Stories

### PC-01: Criação de conta

**User Story**: Como Tesouraria, quero criar contas contábeis organizadas em hierarquia, para classificar lançamentos por categoria e subcategoria.

**Why**: sem plano de contas, nenhum lançamento pode ser classificado — é a base de todo o resto da feature.

**Acceptance Criteria**:
1. WHEN a Tesouraria cria uma conta com tipo (RECEITA ou DESPESA) e nome THEN the system SHALL persist it with `ativo = true` and no parent, or with the given `parent_id` when provided.
2. IF a `parent_id` is given THEN the system SHALL require the parent account to exist and to have the same `tipo` as the new account.
3. IF the actor lacks `financeiro:conta:create` THEN the system SHALL return a forbidden error and SHALL NOT create the account.
4. WHEN a conta é criada THEN the system SHALL record an audit entry naming its id, tipo and nome.

**Independent Test**: criar uma conta raiz "Produtos" (RECEITA) e uma subconta "Camisetas" apontando para ela; ver a hierarquia refletida na consulta.

---

### PC-02: Renomeação restrita ao não-uso

**User Story**: Como Tesouraria, quero poder corrigir o nome de uma conta enquanto ela ainda não foi usada, e ser impedida de renomear uma conta já usada, para que relatórios históricos nunca mudem de significado.

**Why**: é a decisão de negócio central desta sub-spec (`FIN-D-008`) — sem ela, o histórico financeiro perderia integridade.

**Acceptance Criteria**:
1. IF a conta has never been referenced by any `LancamentoFinanceiro` THEN the system SHALL allow changing its `nome`.
2. IF a conta has been referenced by at least one `LancamentoFinanceiro`, regardless of that lançamento's status THEN the system SHALL reject a rename attempt with the error `conta_ja_utilizada`, mesmo se o único lançamento estiver `CANCELADA`.
3. WHEN the system checks whether a conta was ever used THEN it SHALL do so through a read-only port owned by this sub-spec, never by writing to `ContaContabil` from `02-lancamentos` nor by writing to `LancamentoFinanceiro` from this sub-spec.
4. WHEN a rename succeeds THEN the system SHALL record an audit entry with the previous and new `nome`.
5. IF the actor lacks `financeiro:conta:update` THEN the system SHALL return a forbidden error and SHALL NOT rename the conta, regardless of whether it was ever used.

**Independent Test**: criar uma conta, renomeá-la com sucesso; criar um lançamento apontando para ela; tentar renomear de novo e ver `conta_ja_utilizada`.

---

### PC-03: Desativação

**User Story**: Como Tesouraria, quero desativar uma conta que não deve mais ser usada, sem apagar seu histórico, para manter o plano de contas organizado sem perder rastreabilidade.

**Why**: `FIN-001` §6 proíbe exclusão definitiva de categoria oficial.

**Acceptance Criteria**:
1. WHEN a Tesouraria desativa uma conta THEN the system SHALL set `ativo = false` and SHALL NOT delete the row.
2. IF a conta is inactive THEN the system SHALL NOT allow it to be chosen as `conta_id` for a new `LancamentoFinanceiro`.
3. WHEN a conta is deactivated THEN the system SHALL record an audit entry.
4. The system SHALL NOT provide any operation that deletes a `ContaContabil` row.
5. IF the actor lacks `financeiro:conta:deactivate` THEN the system SHALL return a forbidden error and SHALL NOT deactivate the conta.

**Independent Test**: desativar uma conta usada; confirmar que lançamentos antigos continuam mostrando-a normalmente e que ela não aparece na lista de contas disponíveis para um novo lançamento.

---

### PC-04: Consulta e listagem

**User Story**: Como qualquer papel com a permissão de leitura, quero consultar o plano de contas, para escolher a conta certa ao lançar ou para entender a estrutura vigente.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:conta:read` lists accounts THEN the system SHALL return them with their hierarchy (`parent_id`) and `ativo` status.
2. IF the actor lacks `financeiro:conta:read` THEN the system SHALL return a forbidden error and SHALL NOT return any row.

**Independent Test**: listar contas e ver ativas e inativas, com a árvore corretamente representada.

## Edge Cases

- Renomear uma conta cujo único lançamento está `CANCELADA`: ainda recusado (PC-02 AC2) — o critério é "já foi referenciada alguma vez", não "tem lançamento ativo".
- Criar uma subconta com `tipo` diferente do pai: recusado (PC-01 AC2).
- Desativar uma conta que já está inativa: idempotente, sem erro (repetir `ativo = false` não é uma falha).

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| PC-01 | Criação de conta | In Tasks | Done (T2) |
| PC-02 | Renomeação restrita | In Design | Not Started |
| PC-03 | Desativação | In Tasks | Done (T3) |
| PC-04 | Consulta e listagem | In Tasks | Done (T2) |

**Coverage:** 4 total, 4 mapped to tasks, 0 unmapped.

## Success Criteria

- Uma conta nunca usada pode ser renomeada; uma conta já usada não pode, sob nenhuma circunstância.
- Nenhuma conta é fisicamente excluída em nenhum fluxo.
- A verificação de uso nunca envolve escrita cruzada entre `01` e `02`.
