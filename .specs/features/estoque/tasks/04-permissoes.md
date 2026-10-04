# Permissões — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/estoque/spec/04-permissoes.md`
**Design**: `.specs/features/estoque/design/04-permissoes.md`
**Status**: Approved (2026-10-03); tests-first, um commit atômico por tarefa.

**Pré-requisito**: `01-produtos`, `02-movimentacoes`, `03-ajustes-e-saldo` completas (as 6 constantes `Perm*` já declaradas pelos casos de uso).

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `module.go` (`Contribution`) | unit | matriz combinada completa (`FoundationContributions`+`financeiro.Contribution`+`estoque.Contribution`) | `api/internal/estoque/module_test.go` | `go test ./...` |
| `cmd/api`, `cmd/bootstrap-admin` (wiring) | integration | confirma que as 6 permissões de `estoque` existem no banco depois do sync | `api/cmd/bootstrap-admin/*_test.go` | `go test -tags=integration ./...` |

## Gate Check Commands

| Level | When | Command |
| --- | --- | --- |
| Full | toda tarefa | `cd api && go vet ./... && go vet -tags=integration ./... && go build ./... && go test ./... && go test -tags=integration ./...` |
| Lint | toda tarefa | `golangci-lint run --path-mode=abs --build-tags=integration` |
| Architecture | toda tarefa | `go test ./internal/... -run TestProductionCodeRespectsTheModuleBoundaries` |

---

## Execution Plan

```
T1 → T2
```

---

## Task Breakdown

### T1: `estoque.Contribution()`

**What**: `api/internal/estoque/module.go` (primeira vez que o arquivo existe) — `Contribution()` com a matriz aprovada (`ESTOQUE_LOJA`: 6; `DIRETORIA`/`CONSELHO_FISCAL`: 3 de leitura; demais: nenhuma), referenciando as constantes `Perm*` de `estoque/app` (nunca strings literais, mesma decisão de `FIN-D-019`).
**Where**: `api/internal/estoque/module.go`
**Depends on**: None (consome só o que `01`-`03` já publicaram)
**Reuses**: `financeiro.Contribution()` como modelo direto.
**Requirement**: PERM-01.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] `estoque.Contribution()` devolve exatamente a matriz aprovada — teste unitário confere `Grants` por papel
- [x] Nenhuma permissão usa string literal fora de `estoque/app`'s constantes `Perm*`
- [x] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(estoque): contribuicao de rbac do modulo`

---

### T2: Wiring em `cmd/api` e `cmd/bootstrap-admin`

**What**: agrega `estoque.Contribution()` junto de `FoundationContributions()` e `financeiro.Contribution()` nos dois composition roots.
**Where**: `api/cmd/api/main.go`, `api/cmd/bootstrap-admin/main.go`
**Depends on**: T1
**Reuses**: wiring já existente de `financeiro.Contribution()` como modelo direto (mesma linha, só mais um argumento).
**Requirement**: PERM-01.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] `app.BuildMatrix(append(app.FoundationContributions(), financeiro.Contribution(), estoque.Contribution())...)` (ou equivalente) nos dois arquivos
- [x] Teste de integração confirma 6 novas permissões `estoque:%` no banco depois do sync (mesmo padrão do teste que `FIN-D-020` adicionou para `financeiro:%`, contando 16 — aqui o número esperado é 6)
- [x] Nenhum teste existente de `identity`/`httpapi` quebra por `PRESIDENTE`/`ADMIN_SISTEMA` ganharem/perderem cobertura inesperada — revisar os mesmos ambientes de teste que `FIN-D-020` listou (`identity/app`, `identity/http`, `httpapi`), já que a matriz de produção muda de novo
- [x] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(estoque): wiring da contribuicao de rbac nos composition roots`

**Riscos/ambiguidades**: mesma classe de achado mecânico que `FIN-D-020` encontrou para `financeiro` — qualquer ambiente de teste que construa a matriz sem `estoque.Contribution()` pode divergir sutilmente da produção (ex. `PRESIDENTE` ganhando uma vantagem a mais sobre outro papel só por essas 6 permissões). Corrigir como consequência técnica direta, registrando o achado em `STATE.md`, sem pedir nova aprovação — mesma regra já estabelecida.

---

## Phase Execution Map

```
T1 ──→ T2
```

---

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1: Contribution | 1 arquivo | ✅ Granular |
| T2: Wiring | 2 arquivos, mesma mudança mecânica nos dois | ✅ Aceitável (mesmo padrão de `financeiro/05-permissoes/T3`) |

---

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |

---

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1: Contribution | `module.go` | unit | unit | ✅ OK |
| T2: Wiring | `cmd/*` | integration | integration | ✅ OK |
