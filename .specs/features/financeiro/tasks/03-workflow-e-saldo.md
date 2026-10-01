# Workflow e Saldo — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/financeiro/spec/03-workflow-e-saldo.md`
**Design**: `.specs/features/financeiro/design/03-workflow-e-saldo.md`
**Status**: In Progress (T1 e T2 concluídas); `tests-first`, um commit atômico por tarefa.

**Pré-requisito cruzado**: todas as tarefas aqui operam sobre a tabela `lancamentos` — a migração `02-lancamentos/T1` precisa existir primeiro. Nenhuma tarefa desta spec cria ou altera schema.

## Test Coverage Matrix

| Layer | Tests | Notes |
|---|---|---|
| Casos de uso (`app`) | integration | Postgres real via testcontainers |

## Gate Check Commands

| Level | When | Command |
|---|---|---|
| Full | fim de cada tarefa | `cd api && go vet ./... && go test -tags=integration ./...` |

## Execution Plan

```
T1 → T3
T2 → T3
T1 → T4
T2 → T4
T3 → T4
```

## Task Breakdown

### T1: `Receber`

**What**: Transição `CRIADA→RECEBIDA` para lançamento `RECEITA`.
**Where**: `api/internal/financeiro/app/receber_lancamento.go`
**Depends on**: None
**Reuses**: `platform/authz`, `platform/audit`, `platform/database.WithTx`
**Requirement**: WKF-01 (AC1, AC2, AC5, AC6 — parte de receita)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] `CRIADA→RECEBIDA` só para `RECEITA`
- [x] Fora de `CRIADA`: `lancamento_nao_pode_ser_recebido`
- [x] Sem `financeiro:lancamento:receive`, recusa
- [x] Audita `lancamento.receive`
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona transicao de recebimento`

---

### T2: `Pagar`

**What**: Transição `CRIADA→PAGA` para lançamento `DESPESA`.
**Where**: `api/internal/financeiro/app/pagar_lancamento.go`
**Depends on**: None
**Reuses**: idem T1
**Requirement**: WKF-01 (AC3, AC4, AC5, AC6 — parte de despesa)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] `CRIADA→PAGA` só para `DESPESA`
- [x] Fora de `CRIADA`: `lancamento_nao_pode_ser_pago`
- [x] Sem `financeiro:lancamento:pay`, recusa
- [x] Audita `lancamento.pay`
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona transicao de pagamento`

---

### T3: `Cancelar`

**What**: Cancelamento de qualquer estado não-terminal, motivo obrigatório, inclusive de lançamento já liquidado.
**Where**: `api/internal/financeiro/app/cancelar_lancamento.go`
**Depends on**: T1, T2
**Reuses**: idem T1
**Requirement**: WKF-02 (AC1-AC6)

**Tools**: MCP: NONE · Skill: `security-best-practices` (motivo obrigatório, garantir que não vira devolução acidentalmente)

**Done when**:
- [ ] Cancela de `CRIADA`, `RECEBIDA` ou `PAGA`, nunca de `CANCELADA` de novo
- [ ] Motivo vazio é recusado
- [ ] Nunca escreve em outro lançamento (não é devolução)
- [ ] Sem `financeiro:lancamento:cancel`, recusa
- [ ] Audita com o motivo
- [ ] Cada AC listado em Requirement tem ao menos um teste
- [ ] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona cancelamento de lancamento`

---

### T4: `ConsultarSaldo`

**What**: Saldo em regime de caixa (`RECEBIDA` soma, `PAGA` subtrai, `CRIADA`/`CANCELADA` nunca contam).
**Where**: `api/internal/financeiro/app/consultar_saldo.go`
**Depends on**: T1, T2, T3
**Reuses**: `platform/money`
**Requirement**: WKF-03 (AC1-AC3)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Soma `RECEBIDA` menos soma `PAGA`, em centavos
- [ ] `CRIADA` e `CANCELADA` nunca contam — inclusive um lançamento cancelado depois de `RECEBIDA`/`PAGA` (prova a janela de saldo documentada no edge case da spec)
- [ ] Sem `financeiro:saldo:read`, recusa
- [ ] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona consulta de saldo`

---

## Task Granularity Check

| Task | Scope | Status |
|---|---|---|
| T1-T4 | 1 arquivo, 1 caso de uso cada | ✅ Granular |
