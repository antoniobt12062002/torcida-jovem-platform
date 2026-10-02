# Workflow e Saldo — Especificação

Status: Aprovada em 2026-09-29. Sub-spec 3 de 5 da feature `financeiro`. Decisões compartilhadas em `.specs/features/financeiro/STATE.md` (`FIN-D-NNN`), não repetidas aqui. Narrativa em português; critérios de aceite em inglês (EARS).

## Problem Statement

`financeiro` precisa das regras de ciclo de vida de um lançamento — quando ele passa a representar dinheiro que efetivamente mudou de mãos — e do cálculo do saldo derivado disso. Esta spec não possui entidade própria: define regras sobre a coluna `status` de `LancamentoFinanceiro` (dona: `02-lancamentos`).

## Goals

- Liquidar um lançamento: `RECEBIDA` para receita, `PAGA` para despesa (`FIN-D-001`).
- Cancelar um lançamento, de qualquer estado não-terminal, com motivo obrigatório (`FIN-D-006`).
- Calcular o saldo em regime de caixa: só `RECEBIDA`/`PAGA` contam (`FIN-D-001`).

## Out of Scope

| Item | Razão |
|---|---|
| Criação/edição de lançamento | Pertence a `02-lancamentos` |
| "Com comprovante" como estado | `FIN-D-005` — é condição derivada, não gerenciada aqui |
| Pagamento/recebimento parcial | `FIN-D-002` |
| Vínculo obrigatório entre cancelamento e lançamento substituto (`substituido_por_id`) | Rejeitado na revisão arquitetural — a janela de saldo descrita abaixo é aceita como está |
| Conciliação bancária | `FIN-D-004` |
| Regime de competência | `FIN-D-001` — só caixa no V1 |

## Assumptions & Open Questions

| Question | Chosen default | Rationale | Confirmed |
|---|---|---|---|
| Regime | Caixa — só `RECEBIDA`/`PAGA` contam para o saldo | `FIN-D-001` | y |
| Permissões de liquidação | `financeiro:lancamento:receive` e `financeiro:lancamento:pay`, distintas, sem genérica | `FIN-D-011` | y |
| Cancelamento de lançamento liquidado | Permitido, de `RECEBIDA` ou `PAGA`, sempre com motivo; pode alterar o saldo calculado temporariamente até o lançamento correto ser criado e liquidado — consequência aceita, sem mitigação automática no V1 | `FIN-D-006` | y |
| Cancelamento × devolução | Nunca a mesma coisa: cancelar uma receita `RECEBIDA` não é como registrar uma devolução (`02-lancamentos`, `FIN-D-003`) | `FIN-D-006` | y |
| Saldo | Consulta com permissão própria (`financeiro:saldo:read`), não reaproveita `financeiro:lancamento:read` | `FIN-D-013` | y |

Open questions: none.

## User Stories

### WKF-01: Liquidação

**User Story**: Como Tesouraria, quero marcar uma receita como recebida ou uma despesa como paga, para refletir que o dinheiro efetivamente mudou de mãos.

**Acceptance Criteria**:
1. WHEN a Tesouraria marks a RECEITA lançamento whose status is `CRIADA` as received THEN the system SHALL set its status to `RECEBIDA`.
2. IF a RECEITA lançamento's status is not `CRIADA` THEN the system SHALL reject the receiving attempt with the error `lancamento_nao_pode_ser_recebido`.
3. WHEN a Tesouraria marks a DESPESA lançamento whose status is `CRIADA` as paid THEN the system SHALL set its status to `PAGA`.
4. IF a DESPESA lançamento's status is not `CRIADA` THEN the system SHALL reject the paying attempt with the error `lancamento_nao_pode_ser_pago`.
5. IF the actor lacks `financeiro:lancamento:receive` (for a receita) or `financeiro:lancamento:pay` (for a despesa) THEN the system SHALL return a forbidden error and SHALL NOT change the status.
6. WHEN a liquidação succeeds THEN the system SHALL record an audit entry with a distinct action per type (`lancamento.receive` or `lancamento.pay`).

**Independent Test**: marcar uma receita `CRIADA` como recebida; tentar marcar de novo e ver `lancamento_nao_pode_ser_recebido`.

---

### WKF-02: Cancelamento

**User Story**: Como Tesouraria, quero cancelar um lançamento, mesmo já liquidado, com motivo registrado, para corrigir um erro sem apagar o histórico.

**Acceptance Criteria**:
1. WHEN a Tesouraria cancels a lançamento whose status is `CRIADA`, `RECEBIDA` or `PAGA`, providing a non-empty motivo THEN the system SHALL set its status to `CANCELADA`.
2. IF the motivo is empty THEN the system SHALL reject the cancellation.
3. IF the lançamento's status is already `CANCELADA` THEN the system SHALL reject the cancellation attempt.
4. WHEN a `RECEBIDA` or `PAGA` lançamento is cancelled THEN the system SHALL NOT treat this as a devolução and SHALL NOT write to any other lançamento row.
5. IF the actor lacks `financeiro:lancamento:cancel` THEN the system SHALL return a forbidden error and SHALL NOT cancel it.
6. WHEN a cancellation succeeds THEN the system SHALL record an audit entry with the motivo.

**Independent Test**: cancelar um lançamento `RECEBIDA` com motivo; ver `status = CANCELADA` e o saldo (WKF-03) refletir a mudança imediatamente.

---

### WKF-03: Saldo

**User Story**: Como Tesouraria, Diretoria, Presidente ou Conselho Fiscal, quero consultar o saldo calculado, para saber quanto dinheiro a instituição efetivamente tem.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:saldo:read` requests the saldo THEN the system SHALL return the sum of `valor_liquido_cents` of every `RECEBIDA` lançamento minus the sum of `valor_liquido_cents` of every `PAGA` lançamento.
2. The system SHALL NOT include `CRIADA` or `CANCELADA` lançamentos in the saldo calculation, regardless of how recently they changed status.
3. IF the actor lacks `financeiro:saldo:read` THEN the system SHALL return a forbidden error and SHALL NOT return a value.

**Independent Test**: receber uma receita de R$100 e pagar uma despesa de R$40; ver saldo R$60. Cancelar a despesa; ver saldo subir para R$100.

## Edge Cases

- **Cancelamento de lançamento liquidado e a janela de saldo** (`FIN-D-006`, destacado explicitamente conforme a revisão arquitetural): cancelar uma receita `RECEBIDA` ou despesa `PAGA` altera o saldo calculado imediatamente, mesmo que o dinheiro já tenha fisicamente entrado ou saído do caixa. Entre esse cancelamento e a criação (e, quando aplicável, liquidação) de um lançamento corrigido, o saldo do sistema pode divergir do caixa físico. Esta é uma consequência **aceita** no V1 — nenhuma tarefa desta spec implementa mitigação (sem `substituido_por_id`, sem transação vinculando as duas operações).
- Cancelar uma despesa que é uma devolução (`devolucao_de_id` preenchido): segue exatamente a mesma regra — a receita original apontada por ela continua intocada, porque o cancelamento só afeta a própria despesa.
- Tentar receber uma DESPESA ou pagar uma RECEITA: recusado estruturalmente — WKF-01 AC1/AC3 só se aplicam ao tipo correspondente.

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| WKF-01 | Liquidação | In Tasks | Done (T1, T2) |
| WKF-02 | Cancelamento | In Tasks | Done (T3) |
| WKF-03 | Saldo | In Tasks | Not Started |

**Coverage:** 3 total, 3 mapped to tasks, 0 unmapped.

## Success Criteria

- Nenhuma transição além de `CRIADA→RECEBIDA/PAGA` e `{CRIADA,RECEBIDA,PAGA}→CANCELADA` é possível.
- O saldo nunca conta `CRIADA` nem `CANCELADA`.
- A janela de inconsistência descrita acima é conhecida e documentada, não escondida.
