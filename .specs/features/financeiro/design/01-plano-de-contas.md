# Plano de Contas — Design

**Spec**: `.specs/features/financeiro/spec/01-plano-de-contas.md`
**Status**: Approved (2026-09-29).

## Architecture Overview

`01-plano-de-contas` é a única dona da tabela `contas_contabeis`/entidade `Conta`. O único ponto de acoplamento com `02-lancamentos` é a checagem de uso na renomeação (`FIN-D-008`), resolvida por uma porta de leitura, nunca por escrita cruzada. Ambas as sub-specs vivem no mesmo módulo Go `financeiro` (`domain/app/infra/http` compartilhados, não pacotes por sub-spec — ver Components).

```mermaid
graph TD
    UC[caso de uso: renomear conta<br/>financeiro/app] --> Port["LancamentoExistenceChecker<br/>(porta definida em financeiro/app)"]
    Port -.implementada por.-> Infra["financeiro/infra: consulta read-only<br/>a lancamentos.conta_id"]
    UC --> AZ[authz.Require]
    UC --> TX[database.WithTx]
    TX --> DB[(contas)]
    TX --> AUD[audit.Recorder]
```

## Code Reuse Analysis

| Component | Location | How to Use |
|---|---|---|
| Permissões | `platform/authz` | `Require(ctx, actor, financeiro:conta:*)` — nomes definidos em `05-permissoes` |
| Auditoria | `platform/audit` | `Record` na mesma transação de cada operação |
| Transação | `platform/database` | `WithTx` |

## Components

`financeiro` segue o esqueleto já documentado em `docs/architecture/architecture-overview.md` (mesmo padrão de `identity`): um único módulo Go, com `domain/app/infra/http` compartilhados por todas as 5 sub-specs — nenhuma sub-spec vira um pacote Go próprio. A independência das 5 sub-specs é documental/de execução (spec, design, tasks, testes, PR), não uma árvore de pacotes paralela. Plano de contas contribui:

- **`domain/conta.go`**: entidade `Conta` (tipo, nome, `parent_id`, `ativo`) e os erros de domínio (`ErrContaJaUtilizada`, `ErrContaInativa`, `ErrContaTipoIncompativel`).
- **`app/criar_conta.go`**: caso de uso `CriarConta(ctx, CriarContaInput) (Conta, error)`.
- **`app/listar_contas.go`**: caso de uso `ListarContas(ctx, ListarContasInput) ([]Conta, error)`.
- **`app/desativar_conta.go`**: caso de uso `DesativarConta(ctx, DesativarContaInput) error`.
- **`app/renomear_conta.go`** (T4, não implementado ainda): caso de uso `RenomearConta(ctx, RenomearContaInput) (Conta, error)` — chama `LancamentoExistenceChecker.TemLancamento(ctx, contaID)` antes de permitir.
- **`app/usecase.go`** (compartilhado por toda `financeiro`, criado pela primeira sub-spec que precisar dele): portas comuns — `Authorizer`, `Auditor`, `TxFunc` — mesmo padrão de `identity/app/usecase.go`.
- **`infra/conta_repository.go`**: `ContaRepository`, implementando as portas que `app` define para persistir/consultar `Conta`.
- **Porta definida por esta sub-spec**: `type LancamentoExistenceChecker interface { TemLancamento(ctx context.Context, contaID uuid.UUID) (bool, error) }`, declarada em `app/renomear_conta.go` (T4). A implementação concreta fica em `infra` (`02-lancamentos`, T6) — mesmo pacote Go `financeiro/infra`, então não há import entre módulos de negócio, só duas implementações dentro do mesmo `infra`.

## Data Models

```sql
-- financeiro: plano de contas
CREATE TABLE contas_contabeis (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tipo       text NOT NULL CHECK (tipo IN ('RECEITA', 'DESPESA')),
  nome       text NOT NULL,
  parent_id  uuid REFERENCES contas_contabeis,
  ativo      boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX contas_contabeis_parent_idx ON contas_contabeis (parent_id);
-- GRANT INSERT, SELECT, UPDATE ON contas_contabeis TO tj_app (sem DELETE)
```

`UPDATE` é concedido (nome e `ativo` mudam), mas nunca `DELETE` — mesmo padrão de imutabilidade parcial já usado em `documents` (que nem `UPDATE` concede, porque nada nele muda depois de criado; aqui, dois campos mudam sob regra restrita, então `UPDATE` é necessário, mas sem `DELETE`).

**Relationships**: auto-referência via `parent_id`, mesma `tipo` do pai (validado em aplicação, não em constraint de banco — uma `CHECK` entre linhas da mesma tabela exigiria um gatilho, desnecessário para uma regra que a aplicação já garante antes do INSERT). Referenciada por `lancamentos.conta_id` (dono: `02`), sem FK reversa desta tabela para lá.

## Tech Decisions (only non-obvious ones)

| Decision | Choice | Rationale |
|---|---|---|
| Checagem de uso | Porta de leitura, não flag desnormalizada nem gatilho | `FIN-D-008` — evita escrita cruzada (pior) e mantém a regra visível no código Go de `01`, não escondida em SQL |
| Mesma `tipo` do pai | Validado em aplicação | Regra simples, não justifica um `CHECK` entre linhas via gatilho |
| `UPDATE` concedido a `tj_app` | Sim, restrito à regra de negócio (aplicação decide quando aceitar) | Diferente de `documents`, aqui há campos que legitimamente mudam (nome antes do uso, `ativo`) |
