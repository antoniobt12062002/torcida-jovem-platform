# Workflow e Saldo — Design

**Spec**: `.specs/features/financeiro/spec/03-workflow-e-saldo.md`
**Status**: Approved (2026-09-29). Opera sobre `LancamentoFinanceiro`, dona: `02-lancamentos`.

## Architecture Overview

```mermaid
stateDiagram-v2
    [*] --> CRIADA
    CRIADA --> RECEBIDA: receber (só RECEITA)
    CRIADA --> PAGA: pagar (só DESPESA)
    CRIADA --> CANCELADA: cancelar (motivo)
    RECEBIDA --> CANCELADA: cancelar (motivo)
    PAGA --> CANCELADA: cancelar (motivo)
```

Nenhum outro estado ou transição existe no V1. `04-comprovantes` não aparece neste diagrama — a existência de comprovante nunca é insumo de nenhuma transição (`FIN-D-005`).

Ordem em cada caso de uso: permissão recebida → checar `status` atual permite a transição → transação (`UPDATE status` + campos de cancelamento quando aplicável + auditoria).

## Code Reuse Analysis

| Component | Location | How to Use |
|---|---|---|
| Permissões | `platform/authz` | `financeiro:lancamento:receive`/`pay`/`cancel`, `financeiro:saldo:read` |
| Auditoria | `platform/audit` | Ações distintas `lancamento.receive`/`lancamento.pay`/`lancamento.cancel` (formato `dominio.verbo` exigido por `platform/audit.Register`); `lancamento.cancel` com `Reason` obrigatório, já reforçado por `platform/audit`'s `requiresReason()` no sufixo `.cancel` |
| Transação | `platform/database` | `WithTx` |
| Dinheiro | `platform/money` | soma de `Cents` para o saldo |

## Components

Mesmo módulo Go `financeiro` de `01`/`02` (`domain/app/infra/http` compartilhados; nenhuma sub-spec vira pacote próprio). `03-workflow-e-saldo` não tem entidade nem repositório próprios — opera sobre `Lancamento`/`infra/lancamento_repository.go`, ambos de `02-lancamentos`. Contribui:

- **`app/receber_lancamento.go`**: caso de uso `Receber(ctx, ReceberInput) (Lancamento, error)`.
- **`app/pagar_lancamento.go`**: caso de uso `Pagar(ctx, PagarInput) (Lancamento, error)`.
- **`app/cancelar_lancamento.go`**: caso de uso `Cancelar(ctx, CancelarInput) (Lancamento, error)` — `CancelarInput` inclui `Motivo string`, validado não-vazio antes de qualquer escrita.
- **`app/consultar_saldo.go`**: caso de uso `ConsultarSaldo(ctx, ConsultarSaldoInput) (SaldoCents money.Cents, error)`.

## Tech Decisions (only non-obvious ones)

| Decision | Choice | Rationale |
|---|---|---|
| Saldo | `SELECT COALESCE(SUM(...) FILTER (WHERE status='RECEBIDA'), 0) - COALESCE(SUM(...) FILTER (WHERE status='PAGA'), 0)` numa única query | Simples, correto por definição em regime de caixa (`FIN-D-001`); sem necessidade de materializar um saldo em cache no V1, dado o volume esperado |
| Cancelamento de liquidado | Sem vínculo obrigatório a um lançamento substituto | Decisão fechada na revisão arquitetural (`FIN-D-006`) — a janela de saldo é aceita, não mitigada nesta versão |
| Ações de auditoria por tipo de liquidação | `lancamento.receive` e `lancamento.pay` distintas | `FIN-D-011`, mesmo padrão já usado em `identity` (`user.reactivate`, `admin.promote`, nunca um genérico `status_change`) |
