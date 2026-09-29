# Plano de Contas — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Não buscar arquivos da skill por caminho de sistema de arquivos.

**Se a skill não puder ser ativada, PARE e avise — não prossiga sem ela.**

---

**Spec**: `.specs/features/financeiro/spec/01-plano-de-contas.md`
**Design**: `.specs/features/financeiro/design/01-plano-de-contas.md`
**Status**: In Progress (T1 concluída; `tests-first`, um commit atômico por tarefa).

**Pré-requisito**: nenhum código de `financeiro` existe ainda — este é o primeiro conjunto de tarefas do módulo.

**Nota de ordem cruzada entre sub-specs** (não é uma tarefa desta spec, registrar aqui para não se perder): `T4` (renomear conta) só pode ser testada de ponta a ponta depois que a migração de `02-lancamentos` (tabela `lancamentos`) existir no banco de testes, porque a checagem de uso consulta essa tabela. Isso não muda a ordem de **especificação** (`01` continua especificada e revisada antes de `02`), só a ordem real de **execução de código**: `T4` desta spec deve ser codificada depois que a migração de `02-lancamentos` (a primeira tarefa dela) já existir. **Isso também significa que `01` é implementada em duas PRs**: uma com `T1`-`T3` (sem dependência, PR 1), outra só com `T4` (PR 3, depois de `02` mesclada) — exceção documentada de "uma sub-spec, uma PR", não uma mistura de sub-specs. Ver `financeiro/STATE.md`, seção "Ordem real de execução entre sub-specs".

## Test Coverage Matrix

| Layer | Tests | Notes |
|---|---|---|
| Migração | integration | mesmo padrão de `documents_migration_test.go`: grants, constraints |
| Casos de uso (`app`) | integration | Postgres real via testcontainers |

## Gate Check Commands

| Level | When | Command |
|---|---|---|
| Full | fim de cada tarefa | `cd api && go vet ./... && go test -tags=integration ./...` |

## Execution Plan

```
T1 → T2 → T3 → T4
```

Sequencial — cada tarefa depende só da anterior.

## Task Breakdown

### T1: Migração `contas_contabeis`

**What**: Criar a tabela do plano de contas, com `down` e concessões (`INSERT`/`SELECT`/`UPDATE`, sem `DELETE`).
**Where**: `api/migrations/000006_financeiro_contas.up.sql`
**Depends on**: None
**Reuses**: Migração `000002_audit_log`/`000005_documents` como padrão (guarda de papel `tj_app`, grants explícitos)
**Requirement**: PC-01 (AC1, AC2), PC-03 (AC1, AC4)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Tabela com `id`, `tipo` (CHECK RECEITA|DESPESA), `nome`, `parent_id` (auto-referência), `ativo` (default true)
- [x] `tj_app` recebe `INSERT`, `SELECT`, `UPDATE` — nunca `DELETE`
- [x] Migração falha com mensagem clara se `tj_app` não existir
- [x] `down` desfaz sem erro
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): cria tabela do plano de contas`

---

### T2: `CriarConta` e `ListarContas`

**What**: Casos de uso de criação e listagem, com permissão parametrizada pelo catálogo de `05-permissoes`.
**Where**: `api/internal/financeiro/planocontas/service.go`
**Depends on**: T1
**Reuses**: `platform/authz`, `platform/audit`, `platform/database.WithTx`
**Requirement**: PC-01 (AC1-AC4), PC-04 (AC1, AC2)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cria conta raiz e subconta (mesma `tipo` do pai, validado em aplicação)
- [ ] Sem a permissão `financeiro:conta:create`/`financeiro:conta:read`, recusa sem escrever
- [ ] Toda criação audita
- [ ] Listagem devolve hierarquia e `ativo`
- [ ] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona criacao e consulta do plano de contas`

---

### T3: `DesativarConta`

**What**: Desativação sem exclusão; impede escolher conta inativa em novo lançamento (checagem que `02` vai consumir).
**Where**: `api/internal/financeiro/planocontas/service.go`
**Depends on**: T2
**Reuses**: idem T2
**Requirement**: PC-03 (AC1-AC4)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `ativo = false`, nunca `DELETE`
- [ ] Idempotente (desativar já-inativa não é erro)
- [ ] Sem `financeiro:conta:deactivate`, recusa sem escrever
- [ ] Audita
- [ ] Nenhuma operação de exclusão existe no pacote
- [ ] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona desativacao de conta contabil`

---

### T4: `RenomearConta` e a porta `LancamentoExistenceChecker`

**What**: Renomeação restrita ao não-uso (`FIN-D-008`); define a porta de leitura e sua implementação concreta contra `lancamentos`.
**Where**: `api/internal/financeiro/planocontas/service.go`
**Depends on**: T3
**Reuses**: idem T2; consulta read-only à tabela de `02-lancamentos` (ver nota de ordem cruzada no topo deste arquivo)
**Requirement**: PC-02 (AC1-AC4)

**Tools**: MCP: NONE · Skill: `security-best-practices` (consulta cruzada entre entidades, checar que é read-only)

**Done when**:
- [ ] Conta nunca usada: renomeia com sucesso
- [ ] Conta já usada (mesmo com o único lançamento `CANCELADA`): recusa com `conta_ja_utilizada`
- [ ] A checagem nunca escreve em `lancamentos`; `02` nunca escreve em `contas_contabeis`
- [ ] Renomeação audita nome anterior e novo
- [ ] Sem `financeiro:conta:update`, recusa mesmo para uma conta nunca usada
- [ ] Cada AC listado em Requirement tem ao menos um teste
- [ ] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona renomeacao restrita de conta contabil`

---

## Task Granularity Check

| Task | Scope | Status |
|---|---|---|
| T1 | 1 arquivo (migração) | ✅ Granular |
| T2 | 1 arquivo, 2 casos de uso relacionados | ✅ Granular |
| T3 | 1 arquivo, 1 caso de uso | ✅ Granular |
| T4 | 1 arquivo, 1 caso de uso + 1 porta | ✅ Granular |
