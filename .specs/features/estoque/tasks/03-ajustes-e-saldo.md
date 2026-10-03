# Ajustes e Saldo — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/estoque/spec/03-ajustes-e-saldo.md`
**Design**: `.specs/features/estoque/design/03-ajustes-e-saldo.md`
**Status**: Approved (2026-10-03); tests-first, um commit atômico por tarefa.

**Pré-requisito**: `02-movimentacoes` completa (reaproveita `movimentacoes_estoque`, `infra.MovimentacaoRepository`, `domain.ErrQuantidadeInvalida`/`ErrMotivoObrigatorio`).

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `infra` (extensão: `Saldo` sem lock) | integration | caminho de consulta + caso de SKU sem nenhuma movimentação (saldo zero) | `api/internal/estoque/infra/*_test.go` | `go test -tags=integration ./...` |
| `app` | unit (com doubles) | 1:1 com os ACs de AJS-01/AJS-02 | `api/internal/estoque/app/*_test.go` | `go test ./...` |

## Gate Check Commands

| Level | When | Command |
| --- | --- | --- |
| Full | toda tarefa | `cd api && go vet ./... && go vet -tags=integration ./... && go build ./... && go test ./... && go test -tags=integration ./...` |
| Lint | toda tarefa | `golangci-lint run --path-mode=abs --build-tags=integration` |
| Architecture | toda tarefa | `go test ./internal/... -run TestProductionCodeRespectsTheModuleBoundaries` |

---

## Execution Plan

```
T1

T2 → T3
```

(T1 é independente de T2/T3 — `AjustarEstoque` não precisa do método `Saldo` sem lock.)

---

## Task Breakdown

### T1: `app.AjustarEstoque`

**What**: implementa AJS-01 — permissão própria, motivo obrigatório checado antes de qualquer leitura/escrita, quantidade como delta assinado, sem checagem de saldo.
**Where**: `api/internal/estoque/app/ajustar_estoque.go`
**Depends on**: None (`02-movimentacoes` completa — reaproveita `infra.MovimentacaoRepository.Criar` via porta estreita própria, sem `SaldoComLock` na interface)
**Reuses**: template de `financeiro/app.CancelarLancamento` (ordem permissão→motivo→busca→transação); `motivoBlank` reaproveitado por cópia (pacotes distintos, mesma razão de `FIN-D-028`).
**Requirement**: AJS-01.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Os 5 ACs de AJS-01 cobertos (sucesso inclusive negativo, motivo vazio, quantidade zero, produto inexistente, sem permissão)
- [ ] A porta usada (`MovimentacaoCreator`) não declara `SaldoComLock` — confirmado lendo a interface, garantia estrutural de que este caso de uso não pode checar saldo
- [ ] Auditoria `movimentacao.adjust` com `Reason` preenchido
- [ ] Gate check passes

**Tests**: unit (doubles)
**Gate**: Full
**Commit**: `feat(estoque): caso de uso ajustar estoque`

---

### T2: `infra.MovimentacaoRepository.Saldo` (extensão, sem lock)

**What**: adiciona `Saldo(ctx, produtoID) (int64, error)` — mesma agregação de `SaldoComLock`, mas sem `pg_advisory_xact_lock` (consulta pura).
**Where**: `api/internal/estoque/infra/movimentacao_repository.go` (estende o arquivo de `02-movimentacoes/T3`)
**Depends on**: `02-movimentacoes` completa
**Reuses**: a própria query de `SaldoComLock`, sem a linha do lock.
**Requirement**: AJS-02.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `Saldo` devolve `0` para um SKU sem nenhuma movimentação (nunca erro)
- [ ] `Saldo` reflete corretamente entrada+saída+ajuste+devolução combinados (teste de integração com os 4 tipos)
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(estoque): consulta de saldo sem lock`

---

### T3: `app.ConsultarSaldo`

**What**: implementa AJS-02 — permissão própria, confirma que o produto existe, devolve o saldo exato.
**Where**: `api/internal/estoque/app/consultar_saldo.go`
**Depends on**: T2
**Reuses**: template de `financeiro/app.ConsultarSaldo`.
**Requirement**: AJS-02.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Os 3 ACs de AJS-02 cobertos (sucesso inclusive zero/negativo, produto inexistente, sem permissão)
- [ ] `estoque:saldo:read` é permissão distinta de `estoque:movimentacao:read` — teste confirma que um ator com só a segunda é recusado
- [ ] Gate check passes

**Tests**: unit (doubles)
**Gate**: Full
**Commit**: `feat(estoque): caso de uso consultar saldo`

---

## Phase Execution Map

```
T1

T2 ──→ T3
```

---

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1: AjustarEstoque | 1 arquivo | ✅ Granular |
| T2: Extensão do repositório | 1 arquivo (extensão) | ✅ Granular |
| T3: ConsultarSaldo | 1 arquivo | ✅ Granular |

---

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | None (fora de `02` já completa) | None | ✅ Match |
| T3 | T2 | T2 | ✅ Match |

---

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1: AjustarEstoque | `app` | unit | unit | ✅ OK |
| T2: Extensão do repositório | `infra` | integration | integration | ✅ OK |
| T3: ConsultarSaldo | `app` | unit | unit | ✅ OK |
