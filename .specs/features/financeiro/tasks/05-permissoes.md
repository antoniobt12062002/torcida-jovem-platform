# Permissões — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/financeiro/spec/05-permissoes.md`
**Design**: `.specs/features/financeiro/design/05-permissoes.md`
**Status**: Concluída (T1-T3); `tests-first`, um commit atômico por tarefa (T1 teve 2 commits — estrutural + funcional, `FIN-D-018`; T2 e T3, 1 commit cada).

**Nota de ordem cruzada**: `T2` (completar com as 13 permissões operacionais) só pode ser finalizada depois que `01-plano-de-contas`, `02-lancamentos`, `03-workflow-e-saldo` e `04-comprovantes` já tiverem suas operações especificadas (não necessariamente implementadas em código — os nomes de permissão já estão fechados nas 4 specs). `T1` (esqueleto + preservação institucional) não tem essa dependência e pode ser feita a qualquer momento.

## Test Coverage Matrix

| Layer | Tests | Notes |
|---|---|---|
| `Contribution` + integração com `BuildMatrix` | unit (Go puro, mesmo padrão de `roles_matrix_test.go`) | não precisa de Postgres — `BuildMatrix` é uma função pura |
| Fiação em `main.go`/`bootstrap-admin` | integration | confirma que a API sobe com a matriz nova |

## Gate Check Commands

| Level | When | Command |
|---|---|---|
| Full | fim de cada tarefa | `cd api && go vet ./... && go test -tags=integration ./...` |

## Execution Plan

```
T1 → T2 → T3
```

## Task Breakdown

### T1: Esqueleto da `Contribution` e preservação institucional

**What**: Função `financeiro.Contribution()` com as 3 permissões institucionais preservadas exatamente como no placeholder atual. Inclui, como pré-requisito estrutural, reexportar `Role` e as 8 constantes de papel via `identity/app` (`FIN-D-018`) — `financeiro` não pode importar `identity/domain` diretamente (`architecture_test.go`).
**Where**: `api/internal/financeiro/module.go` (novo); `api/internal/identity/app/roles.go` (novo, re-export estrutural)
**Depends on**: None
**Reuses**: `identity/app.Contribution`, `identity/app.Role`/`RoleConselhoFiscal` (re-export de `identity/domain`, `FIN-D-018`)
**Requirement**: PERM-02 (AC1-AC3)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Declara `financeiro:prestacao_contas:read`, `financeiro:prestacao_contas:approve`, `financeiro:parecer:opine`, concedidas a `CONSELHO_FISCAL`, idênticas ao placeholder de `roles_matrix.go:132-144`
- [x] Comparação da matriz antes/depois (placeholder vs. `Contribution` real) mostra essas 3 permissões idênticas para `CONSELHO_FISCAL`
- [x] Nenhum caso de uso ou endpoint as verifica ainda
- [x] Gate check passes

**Tests**: unit
**Gate**: full
**Commit**: `feat(financeiro): cria esqueleto de permissoes preservando as institucionais`

---

### T2: Completar com as 13 permissões operacionais e a distribuição por papel

**What**: Adicionar as permissões de `01`-`04` (conforme suas specs) e a matriz de concessão de `TESOURARIA`/`DIRETORIA`/`CONSELHO_FISCAL`.
**Where**: `api/internal/financeiro/module.go`
**Depends on**: T1
**Reuses**: idem T1
**Requirement**: PERM-01 (AC1-AC4), PERM-03 (AC1-AC4)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Todas as 13 permissões declaradas com o formato correto, nenhuma `CommonRead`
- [x] `CONSELHO_FISCAL` nunca recebe `create`/`update`/`deactivate`/`receive`/`pay`/`cancel`
- [x] `TESOURARIA` recebe as 13; `DIRETORIA` e `CONSELHO_FISCAL` recebem só as 4 de leitura (`CONSELHO_FISCAL` também as 3 institucionais); `ASSOCIADO`/`ESTOQUE_LOJA`/`EVENTOS`/`ADMIN_SISTEMA` nenhuma
- [x] `BuildMatrix` aceita a `Contribution` sem erro, `PRESIDENTE` recebe tudo automaticamente
- [x] Cada AC listado em Requirement tem ao menos um teste
- [x] Gate check passes

**Tests**: unit
**Gate**: full
**Commit**: `feat(financeiro): completa catalogo de permissoes operacionais`

---

### T3: Substituir o placeholder da fundação

**What**: Remover `FoundationContributions()`'s bloco `financeiro` e usar `financeiro.Contribution()` em `cmd/api/main.go` e `cmd/bootstrap-admin/main.go`.
**Where**: `api/internal/identity/app/roles_matrix.go`, `cmd/api/main.go`, `cmd/bootstrap-admin/main.go`. Consequência técnica direta (achado durante a execução, `FIN-D-020`): mais 9 arquivos de teste precisaram de ajuste — `identity/app/roles_matrix_test.go` (2 testes cujo `want` incluía as 3 institucionais via `FoundationContributions()` isolada), `identity/infra/role_repository_test.go` (3 testes que usavam o placeholder como fixture "removível" + 1 teste de união de permissões + o helper `matrix(t)`), `httpapi/{e2e_test.go,router_integration_test.go}` e `identity/{app/session_service_test.go,app/usecase_env_test.go,http/env_test.go}` (ambiente de teste que precisa espelhar a composição real), `identity/app/{promote_admin_test.go,admin_reset_password_test.go}` (fixture `RoleTesouraria` → `RoleEventos`, já que `TESOURARIA` deixou de ser um papel sem conteúdo de RBAC), e `cmd/bootstrap-admin/bootstrap_integration_test.go` (nova asserção, lacuna de cobertura real encontrada por mutação).
**Depends on**: T2
**Reuses**: idem T1
**Requirement**: PERM-02 (AC3)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Placeholder removido de `roles_matrix.go`
- [x] `main.go` e `bootstrap-admin/main.go` passam `financeiro.Contribution()` junto das demais
- [x] API sobe normalmente, matriz sincroniza sem erro
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): substitui o placeholder de permissoes da fundacao`

---

## Task Granularity Check

| Task | Scope | Status |
|---|---|---|
| T1 | 2 arquivos (`identity/app/roles.go` + `financeiro/module.go`), 2 commits atômicos (estrutural + funcional, mesmo padrão de `04/T1`) — granularidade aceitável: o re-export é pré-requisito estrutural único para a `Contribution`, não uma tarefa independente (`FIN-D-018`) | ✅ Aceitável |
| T2 | 1 arquivo | ✅ Granular |
| T3 | 1 arquivo principal (`roles_matrix.go`) + 2 pontos de fiação (`main.go`, `bootstrap-admin/main.go`) + 10 arquivos de teste ajustados como consequência técnica direta (`FIN-D-020`) — granularidade aceitável: é uma única mudança lógica (trocar o placeholder pela Contribution real), não um agrupamento artificial de tarefas distintas; os ajustes de teste não introduzem nenhuma regra nova, só restauram a fidelidade do ambiente de teste à composição real | ✅ Aceitável |
