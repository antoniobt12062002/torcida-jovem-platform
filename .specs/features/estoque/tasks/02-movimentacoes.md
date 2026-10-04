# Movimentações de Estoque — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/estoque/spec/02-movimentacoes.md`
**Design**: `.specs/features/estoque/design/02-movimentacoes.md`
**Status**: Approved (2026-10-03); tests-first, um commit atômico por tarefa.

**Pré-requisito**: `01-produtos` completa (precisa de `produtos_estoque` e de `ProdutoReader`).

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `domain` | none | build gate only | `api/internal/estoque/domain/*.go` | gate `Full` |
| `infra` (incl. `SaldoComLock`) | integration | caminhos de consulta + **teste de concorrência real** (2 goroutines, Postgres real, disputando o último saldo) | `api/internal/estoque/infra/*_test.go` | `go test -tags=integration ./...` |
| `app` | integration (Postgres real via `testutil.NewTestDB` + `infra` real; `Authz` via `simpleAuthz`) | 1:1 com os ACs de MOV-01 a MOV-04, incluindo todo edge case | `api/internal/estoque/app/*_test.go` (`//go:build integration`) | `go test -tags=integration ./...` |

## Gate Check Commands

| Level | When | Command |
| --- | --- | --- |
| Full | toda tarefa | `cd api && go vet ./... && go vet -tags=integration ./... && go build ./... && go test ./... && go test -tags=integration ./...` |
| Lint | toda tarefa | `golangci-lint run --path-mode=abs --build-tags=integration` |
| Architecture | toda tarefa | `go test ./internal/... -run TestProductionCodeRespectsTheModuleBoundaries` |

---

## Execution Plan

```
T1 → T3
T2 → T3
T3 → T4
T3 → T5
```

---

## Task Breakdown

### T1: Migration `movimentacoes_estoque`

**What**: `000009_estoque_movimentacoes.up.sql`/`.down.sql` — tabela com `tipo`/`origem` via `CHECK`, `quantidade bigint CHECK (quantidade <> 0)`, FKs para `produtos_estoque`, `users` e auto-referência (`movimentacao_de_id`), índice em `produto_id`.
**Where**: `api/migrations/000009_estoque_movimentacoes.up.sql`, `api/migrations/000009_estoque_movimentacoes.down.sql`
**Depends on**: `01-produtos` completa (FK para `produtos_estoque`)
**Reuses**: estrutura de `000007_financeiro_lancamentos` como modelo (FK + auto-referência para devolução).
**Requirement**: infraestrutura para MOV-01 a MOV-04.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `up`/`down` corretos, incluindo ajuste do teste de down-migration de `01-produtos` (desfazer `000009` antes de `000008`, mesmo cuidado que `financeiro/01/T1` teve com `identity`)
- [ ] `CHECK` de `tipo`/`origem`/`quantidade <> 0` confirmados por teste
- [ ] Gate check passes

**Tests**: integration (teste de migração)
**Gate**: Full
**Commit**: `feat(estoque): migration de movimentacoes de estoque`

---

### T2: `domain.Movimentacao`, enums e sentinels

**What**: `TipoMovimentacao`, `OrigemMovimentacao`, `Movimentacao`, `ErrMovimentacaoNaoEncontrada`, `ErrDevolucaoInvalida`, `ErrSaldoInsuficiente`, `ErrQuantidadeInvalida`.
**Where**: `api/internal/estoque/domain/movimentacao.go`
**Depends on**: None
**Reuses**: formato de `financeiro/domain.Lancamento`.
**Requirement**: MOV-01 a MOV-04.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Enums e struct declarados, 4 sentinels
- [ ] Gate check passes

**Tests**: none (build gate only)
**Gate**: Full
**Commit**: `feat(estoque): entidade de movimentacao, enums e sentinels`

---

### T3: `infra.MovimentacaoRepository`

**What**: `Criar`, `Buscar`, `ListarPorProduto`, `SaldoComLock` (`pg_advisory_xact_lock(hashtext(produto_id))` + `SUM`, na mesma transação do chamador).
**Where**: `api/internal/estoque/infra/movimentacao_repository.go`
**Depends on**: T1, T2
**Reuses**: estilo de `financeiro/infra.LancamentoRepository`; `pg_advisory_xact_lock`, estilo de `identity/infra/role_repository.go:60`.
**Requirement**: MOV-01 a MOV-04.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `SaldoComLock` serializa corretamente — **teste de integração real com 2 goroutines disputando a última unidade do mesmo produto: exatamente uma conclui com sucesso, a outra falha por saldo insuficiente, nunca as duas sucedem nem as duas falham**
- [ ] `Buscar` devolve `domain.ErrMovimentacaoNaoEncontrada` quando o id não existe
- [ ] Gate check passes

**Tests**: integration (Postgres real, incluindo o teste de concorrência obrigatório)
**Gate**: Full
**Commit**: `feat(estoque): repositorio de movimentacoes`

---

### T4: `app.RegistrarMovimentacao`

**What**: implementa MOV-01/MOV-02/MOV-03 — entrada, saída, devolução; sinal por tipo; checagem de saldo via lock só quando a quantidade assinada é negativa; auditoria `movimentacao.create` para os 3 tipos.
**Where**: `api/internal/estoque/app/registrar_movimentacao.go`
**Depends on**: T3
**Reuses**: template de `financeiro/app.CriarLancamento` (permissão→validação de referência→transação); decisão de reaproveitar a mesma ação de auditoria para os 3 tipos, mesma lógica de `FIN-D-003`.
**Requirement**: MOV-01, MOV-02, MOV-03.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Os 4+5+4 ACs combinados de MOV-01/02/03 cobertos
- [ ] `ENTRADA` nunca aciona o lock (só quando a quantidade assinada é negativa)
- [ ] `DEVOLUCAO` de uma `SAIDA` nunca é bloqueada por saldo; `DEVOLUCAO` de uma `ENTRADA` segue a mesma regra de `SAIDA`
- [ ] `DEVOLUCAO` referenciando `AJUSTE`, outra `DEVOLUCAO`, produto diferente, ou id inexistente, é recusada com `devolucao_invalida`
- [ ] Gate check passes

**Tests**: integration (Postgres real via `testutil.NewTestDB`; pelo menos um teste por tipo, confirmando o efeito real no saldo agregado)
**Gate**: Full
**Commit**: `feat(estoque): caso de uso registrar movimentacao`

---

### T5: `app.ListarMovimentacoes`

**What**: implementa MOV-04 — consulta pura, ordem de criação.
**Where**: `api/internal/estoque/app/listar_movimentacoes.go`
**Depends on**: T3
**Reuses**: template de `financeiro/app.ListarLancamentos`.
**Requirement**: MOV-04.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Os 2 ACs de MOV-04 cobertos, incluindo listar movimentações de todos os tipos (mesmo as que só `03-ajustes-e-saldo` criará — a listagem já deve devolver qualquer linha da tabela, sem filtrar por tipo)
- [ ] Gate check passes

**Tests**: integration (Postgres real via `testutil.NewTestDB`)
**Gate**: Full
**Commit**: `feat(estoque): caso de uso listar movimentacoes`

---

## Phase Execution Map

```
T1 → T3
T2 → T3
T3 → T4
T3 → T5
```

---

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1: Migration | 1 par up/down | ✅ Granular |
| T2: Entidade+enums+sentinels | 1 arquivo | ✅ Granular |
| T3: Repositório | 1 arquivo | ✅ Granular |
| T4: Caso de uso RegistrarMovimentacao | 1 arquivo | ✅ Granular (cobre 3 tipos pela mesma lógica interna — mesma decisão de granularidade que `financeiro/app.CriarLancamento` já aplicou para RECEITA/DESPESA) |
| T5: Caso de uso ListarMovimentacoes | 1 arquivo | ✅ Granular |

---

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | `01-produtos` completa | — (fora do diagrama local) | ✅ Match |
| T2 | None | None | ✅ Match |
| T3 | T1, T2 | T1, T2 | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | T3 | T3 | ✅ Match |

---

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1: Migration | — | integration | integration | ✅ OK |
| T2: Entidade | `domain` | none | none | ✅ OK |
| T3: Repositório | `infra` | integration (+ concorrência) | integration | ✅ OK |
| T4: RegistrarMovimentacao | `app` | integration | integration | ✅ OK |
| T5: ListarMovimentacoes | `app` | integration | integration | ✅ OK |
