# Lançamentos — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/financeiro/spec/02-lancamentos.md`
**Design**: `.specs/features/financeiro/design/02-lancamentos.md`
**Status**: Concluída (T1-T6); `tests-first`, um commit atômico por tarefa.

**Pré-requisito cruzado**: a migração `T1` desta spec precisa existir antes de `01-plano-de-contas/T4` (renomear conta) poder ser testada de ponta a ponta — ver `financeiro/STATE.md`, "Ordem real de execução entre sub-specs". Dentro desta spec, nenhuma tarefa depende de `01` além da FK `conta_id` (a migração de `01` precisa existir antes de `T1` aqui).

**Nota sobre `T4` (devolução)**: o teste desta tarefa precisa de uma receita em `status = RECEBIDA` como pré-condição — como `03-workflow-e-saldo` (dona da transição `receber`) ainda não existe neste ponto da execução, o teste cria essa pré-condição inserindo a linha diretamente via SQL de teste (fixture), não через o caso de uso de `03`. Isso não é uma dependência de código entre `02` e `03`, só uma montagem de cenário de teste.

## Test Coverage Matrix

| Layer | Tests | Notes |
|---|---|---|
| Migração | integration | grants, constraints, índices |
| Casos de uso (`app`) | integration | Postgres real via testcontainers |

## Gate Check Commands

| Level | When | Command |
|---|---|---|
| Full | fim de cada tarefa | `cd api && go vet ./... && go test -tags=integration ./...` |

## Execution Plan

```
T1 → T2 → T3
T2 → T4
T2 → T5
T1 → T6
```

## Task Breakdown

### T1: Migração `lancamentos`

**What**: Criar a tabela de lançamentos, com `down` e concessões (`INSERT`/`SELECT`/`UPDATE`, sem `DELETE`).
**Where**: `api/migrations/000007_financeiro_lancamentos.up.sql`
**Depends on**: None (nesta spec; depende da migração de `01-plano-de-contas` já existir, por causa da FK)
**Reuses**: mesmo padrão de `000006_financeiro_contas`
**Requirement**: LAN-01 (AC1)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Colunas conforme o modelo de dados do design; `status` nasce `CRIADA`, restrito por `CHECK` aos 4 valores válidos
- [x] `devolucao_de_id` auto-referência, nullable
- [x] `tj_app` recebe `INSERT`, `SELECT`, `UPDATE` — nunca `DELETE`
- [x] Migração falha com mensagem clara se `tj_app` ou `contas_contabeis` não existirem
- [x] `down` desfaz sem erro
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): cria tabela de lancamentos`

---

### T2: `CriarLancamento`

**What**: Criação de receita/despesa, com validação de conta (existe, ativa, mesma `tipo`) e cálculo de `valor_liquido_cents`.
**Where**: `api/internal/financeiro/app/criar_lancamento.go`
**Depends on**: T1
**Reuses**: `platform/authz`, `platform/audit`, `platform/money`, `platform/database.WithTx`
**Requirement**: LAN-01 (AC1-AC5)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Cria com `status = CRIADA`, `valor_liquido_cents` correto
- [x] Recusa conta inativa/inexistente (`conta_invalida`) e tipo incompatível (`conta_tipo_incompativel`)
- [x] Sem `financeiro:lancamento:create`, recusa sem escrever
- [x] Audita
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona criacao de lancamento`

---

### T3: `EditarLancamento`

**What**: Edição restrita a `status = CRIADA`.
**Where**: `api/internal/financeiro/app/editar_lancamento.go`
**Depends on**: T2
**Reuses**: idem T2
**Requirement**: LAN-02 (AC1-AC4)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Edita campos permitidos só em `CRIADA`, recalcula `valor_liquido_cents`
- [x] Fora de `CRIADA`: recusa com `lancamento_imutavel`
- [x] Sem `financeiro:lancamento:update`, recusa
- [x] Audita antes/depois de cada campo alterado
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona edicao de lancamento em CRIADA`

---

### T4: `CriarDevolucao`

**What**: Criação de despesa com `devolucao_de_id`, validando o lançamento referenciado.
**Where**: `api/internal/financeiro/app/criar_devolucao.go`
**Depends on**: T2
**Reuses**: idem T2
**Requirement**: LAN-03 (AC1-AC4)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Exige referenciado `RECEITA` e `RECEBIDA` (`devolucao_invalida` caso contrário)
- [x] Receita original nunca é escrita
- [x] Usa a mesma permissão de `T2` (`financeiro:lancamento:create`)
- [x] Cada AC listado em Requirement tem ao menos um teste
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona devolucao vinculada ao lancamento original`

---

### T5: `ListarLancamentos`

**What**: Consulta/listagem com todos os campos.
**Where**: `api/internal/financeiro/app/listar_lancamentos.go`
**Depends on**: T2
**Reuses**: idem T2
**Requirement**: LAN-04 (AC1, AC2)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Lista com todos os campos, incluindo cancelados
- [x] Sem `financeiro:lancamento:read`, nenhuma linha
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona consulta de lancamentos`

---

### T6: Implementação de `LancamentoExistenceChecker`

**What**: Implementação concreta da porta que `01-plano-de-contas` vai definir (`FIN-D-008`), consultando esta tabela. Nenhuma interface Go existe ainda para essa porta: `01-plano-de-contas/T4` (`RenomearConta`, ainda não implementada) é quem vai declará-la, do lado do consumidor — este tipo concreto a satisfaz estruturalmente quando isso acontecer, mesmo padrão AD-015 já usado no resto do módulo.
**Where**: `api/internal/financeiro/infra/lancamento_existence_checker.go`
**Depends on**: T1
**Reuses**: nenhuma dependência de código de `01` — implementa uma interface que `01` vai definir
**Requirement**: PC-02 (AC3, de `01-plano-de-contas`, consumida aqui)

**Tools**: MCP: NONE · Skill: `security-best-practices` (confirmar que é read-only)

**Done when**:
- [ ] `TemLancamento(ctx, contaID)` retorna verdadeiro se existe ao menos uma linha para a conta, inclusive se `CANCELADA`
- [ ] Nunca escreve em `contas_contabeis`
- [ ] Gate check passes

**Pendência registrada (decisão do mantenedor)**: nenhum `financeiro/module.go` (composition root, equivalente a `identity/module.go`) existe ainda — não foi entrega de nenhuma tarefa aprovada de `01` ou `02`. O wiring real de `LancamentoExistenceChecker` (e de todo caso de uso de `01`/`02` até aqui) fica para o primeiro ponto de composição aprovado (provavelmente junto de `01-plano-de-contas/T4` ou da camada HTTP) — não é responsabilidade desta tarefa.

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): implementa verificacao de uso de conta para renomeacao`

---

## Task Granularity Check

| Task | Scope | Status |
|---|---|---|
| T1 | 1 arquivo (migração) | ✅ Granular |
| T2-T5 | 1 arquivo, 1 caso de uso cada | ✅ Granular |
| T6 | 1 arquivo, 1 implementação de porta | ✅ Granular |
