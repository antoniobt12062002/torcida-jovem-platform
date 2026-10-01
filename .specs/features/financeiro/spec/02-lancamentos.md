# Lançamentos — Especificação

Status: Aprovada em 2026-09-29. Sub-spec 2 de 5 da feature `financeiro`. Decisões compartilhadas em `.specs/features/financeiro/STATE.md` — referenciadas por `FIN-D-NNN`, não repetidas. Narrativa em português; critérios de aceite em inglês (EARS).

## Problem Statement

`financeiro` precisa da entidade central que registra receitas e despesas: seus valores, sua classificação e, quando aplicável, sua relação com uma devolução de dinheiro já recebido.

## Goals

- Criar lançamentos (receita ou despesa), com valor bruto, taxa, valor líquido e forma de pagamento.
- Garantir a consistência entre o tipo do lançamento e o tipo da conta usada: um lançamento `RECEITA` só pode usar conta `RECEITA`; um lançamento `DESPESA` só pode usar conta `DESPESA` — regra de negócio do lançamento, não um detalhe técnico (`FIN-D-015`).
- Permitir editar um lançamento só enquanto ele ainda não foi liquidado (`FIN-D-007`).
- Registrar devolução de dinheiro já recebido como uma nova despesa vinculada à receita original (`FIN-D-003`).
- Consultar/listar lançamentos.

## Out of Scope

| Item | Razão |
|---|---|
| Transições de estado (`receber`/`pagar`/`cancelar`) e cálculo de saldo | Pertence a `03-workflow-e-saldo` — esta spec só define a entidade e sua criação/edição |
| Anexar/consultar comprovante | Pertence a `04-comprovantes` |
| Pagamento parcial, parcelamento | `FIN-D-002` |
| Mecanismo de estorno formal | `FIN-D-003` — devolução é uma nova despesa, não um evento próprio |
| Código contábil formal na conta referenciada | `FIN-D-009`, pertence a `01-plano-de-contas` |

## Assumptions & Open Questions

| Question | Chosen default | Rationale | Confirmed |
|---|---|---|---|
| Campos do lançamento | tipo, conta_id, valor_bruto_cents, taxa_cents (default 0), valor_liquido_cents (calculado e armazenado), forma_pagamento, status, devolucao_de_id (nullable) | `FIN-001` §7-8, `financeiro/STATE.md` modelo de dados | y |
| `tipo` do lançamento deve bater com `tipo` da conta | Sim — uma receita só pode apontar para uma conta RECEITA, uma despesa só para DESPESA | Regra de consistência do lançamento, aprovada explicitamente pelo mantenedor (`FIN-D-015`) | y |
| Edição permitida | Só enquanto `status = CRIADA` | `FIN-D-007` | y |
| Devolução | Nova despesa com `devolucao_de_id`; exige que o lançamento referenciado seja RECEITA e esteja RECEBIDA; receita original nunca é escrita | `FIN-D-003` | y |
| Permissão de devolução | Mesma `financeiro:lancamento:create` — não é uma operação nova, é um lançamento com um campo extra | Consequência direta de FIN-D-003 tratar devolução como "uma nova despesa", não como operação distinta | y |
| Verificação de uso da conta (para `01`) | Este pacote expõe a implementação concreta da porta `LancamentoExistenceChecker` definida por `01-plano-de-contas`, consultando esta mesma tabela, sem nunca escrever em `contas_contabeis` | `FIN-D-008` | y |

Open questions: none.

## User Stories

### LAN-01: Criação de lançamento

**User Story**: Como Tesouraria, quero registrar uma receita ou despesa com seus valores e forma de pagamento, para manter o histórico financeiro da instituição.

**Acceptance Criteria**:
1. WHEN a Tesouraria creates a lançamento with tipo, conta_id, valor_bruto_cents, taxa_cents (default 0) and forma_pagamento THEN the system SHALL persist it with `status = CRIADA` and `valor_liquido_cents = valor_bruto_cents - taxa_cents`, computed once at creation.
2. IF conta_id refers to an inactive or nonexistent account THEN the system SHALL reject the lançamento with the error `conta_invalida` and SHALL NOT create it.
3. The system SHALL enforce that a `RECEITA` lançamento only uses a `RECEITA` conta, and a `DESPESA` lançamento only uses a `DESPESA` conta; IF the lançamento's tipo does not match the referenced conta's tipo THEN it SHALL reject the lançamento with the error `conta_tipo_incompativel` and SHALL NOT create it.
4. IF the actor lacks `financeiro:lancamento:create` THEN the system SHALL return a forbidden error and SHALL NOT create the lançamento.
5. WHEN a lançamento is created THEN the system SHALL record an audit entry with tipo, conta_id, valor_bruto_cents and forma_pagamento.

**Independent Test**: criar uma despesa de R$100 com taxa de R$4 e ver `valor_liquido_cents = 9600`, `status = CRIADA`.

---

### LAN-02: Edição restrita a `CRIADA`

**User Story**: Como Tesouraria, quero corrigir um lançamento enquanto ele ainda não foi liquidado, para consertar um erro de digitação sem precisar cancelar.

**Acceptance Criteria**:
1. WHILE a lançamento's status is `CRIADA` THE system SHALL allow editing its conta_id, valor_bruto_cents, taxa_cents and forma_pagamento, recalculating valor_liquido_cents.
2. IF a lançamento's status is not `CRIADA` THEN the system SHALL reject any edit attempt with the error `lancamento_imutavel`.
3. IF the actor lacks `financeiro:lancamento:update` THEN the system SHALL return a forbidden error and SHALL NOT edit it.
4. WHEN an edit succeeds THEN the system SHALL record an audit entry with the previous and new values of every changed field.
5. The system SHALL reapply LAN-01 AC2/AC3's conta validation (exists, ativo, same tipo as the lançamento) to the edited conta_id, rejecting with `conta_invalida` or `conta_tipo_incompativel` accordingly — the invariant holds for the lançamento's whole lifecycle, not only at creation.

**Independent Test**: editar o valor de um lançamento `CRIADA`; tentar editar um `RECEBIDA` e ver `lancamento_imutavel`.

---

### LAN-03: Devolução

**User Story**: Como Tesouraria, quero registrar a devolução de uma receita já recebida como uma nova despesa vinculada a ela, para preservar o histórico da receita original intacto.

**Acceptance Criteria**:
1. WHEN a Tesouraria creates a despesa with `devolucao_de_id` pointing to another lançamento THEN the system SHALL require that referenced lançamento to have `tipo = RECEITA` and `status = RECEBIDA`.
2. IF the referenced lançamento is not RECEITA or is not RECEBIDA THEN the system SHALL reject the devolução with the error `devolucao_invalida` and SHALL NOT create it.
3. WHEN a devolução is created THEN the system SHALL NOT modify the referenced receita in any way — it SHALL remain exactly as it was, still RECEBIDA.
4. The system SHALL NOT store any field on the original receita referencing the devolução (unidirectional relationship only).

**Independent Test**: criar uma receita, marcá-la recebida (fora de escopo desta spec, mas necessário como pré-condição de teste), criar uma devolução apontando para ela; ver a receita original inalterada.

---

### LAN-04: Consulta e listagem

**User Story**: Como qualquer papel com a permissão de leitura, quero consultar lançamentos, para acompanhar as movimentações financeiras.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:lancamento:read` lists or queries lançamentos THEN the system SHALL return them with every field, including status and devolucao_de_id when present.
2. IF the actor lacks `financeiro:lancamento:read` THEN the system SHALL return a forbidden error and SHALL NOT return any row.

**Independent Test**: listar lançamentos e ver os campos completos, incluindo os cancelados (nenhum é ocultado por padrão).

## Edge Cases

- Editar um lançamento que é uma devolução: permitido enquanto `CRIADA`, como qualquer outro (LAN-02 não distingue devolução de lançamento comum).
- Criar uma devolução apontando para outra devolução: **não restringido por esta spec** — o único requisito é que o lançamento referenciado seja `RECEITA` e `RECEBIDA`; uma devolução é sempre `DESPESA`, então encadear devolução-de-devolução é estruturalmente impossível (uma despesa nunca satisfaz "tipo = RECEITA").
- Criar lançamento apontando para conta desativada depois de já ter sido usada: recusado (LAN-01 AC2) — desativação sempre impede novo uso, independente de uso passado.

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| LAN-01 | Criação de lançamento | In Tasks | Done (T2) |
| LAN-02 | Edição restrita a CRIADA | In Tasks | Done (T3) |
| LAN-03 | Devolução | In Tasks | Done (T4) |
| LAN-04 | Consulta e listagem | In Tasks | Done (T5) |

**Coverage:** 4 total, 4 mapped to tasks, 0 unmapped.

## Success Criteria

- `valor_liquido_cents` sempre reflete `valor_bruto_cents - taxa_cents`, calculado uma vez.
- Nenhum lançamento fora de `CRIADA` pode ser editado.
- Uma devolução nunca escreve na receita original.
