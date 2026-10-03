# API HTTP — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/financeiro/spec/06-api-http.md`
**Design**: `.specs/features/financeiro/design/06-api-http.md`
**Status**: Approved (2026-10-03); `tests-first`, um commit atômico por tarefa.

**Pré-requisito**: `01`-`05` completas em `develop` (confirmado). Nenhuma decisão de negócio aberta nesta sub-spec — as únicas pendências (`FIN-D-021` a `FIN-D-024`) são técnicas/de escopo, resolvidas ou explicitamente marcadas como fora do escopo na spec/design. **Antes de iniciar T1, confirme que `FIN-D-021` a `FIN-D-024` foram revisadas pelo mantenedor** (ver relatório de entrega) — nenhuma delas bloqueia o início, mas `FIN-D-021` (paginação fora do escopo) afeta o contrato de T2.

## Test Coverage Matrix

| Layer | Tests | Notes |
|---|---|---|
| Composition root (`financeiro.New`) | unit | sem I/O, só confere que nenhum campo fica nil |
| `cmd/api` (wiring) | integration | primeira suíte de teste de `cmd/api` — não existe nenhuma hoje |
| Contrato OpenAPI | `redocly lint` + `go generate` | confirma validade sem precisar de handler nenhum |
| Handlers | integration | mesmo padrão de `identity/http`, requisição HTTP real contra casos de uso reais |

## Gate Check Commands

| Level | When | Command |
|---|---|---|
| Full | T1, T8 | `cd api && go vet ./... && go vet -tags=integration ./... && go build ./... && go test ./... && go test -tags=integration ./...` |
| Full (com exceção documentada) | T2-T7 | mesmo comando acima — **exceto** `TestEveryRegisteredRouteHasAContractOperationAndViceVersa` e `TestRoutesAndContractFilesAreTheSameList`, em `internal/httpapi`, que ficam vermelhos por desenho até `T8` (`FIN-D-026`); todo o resto, incluindo os outros ~27 testes de `httpapi`, continua exigido verde |
| Contract | T2 e T8 | `cd web && pnpm lint:api && pnpm gen:api:check` |
| Lint | toda tarefa | `golangci-lint run --path-mode=abs --build-tags=integration` |
| Architecture | toda tarefa | `go test ./internal/... -run TestProductionCodeRespectsTheModuleBoundaries` |

## Execution Plan

```
T1 → T7
T2 → T3
T3 → T4
T3 → T5
T3 → T6
T3 → T7
T4 → T8
T5 → T8
T6 → T8
T7 → T8
```

`T1` (composition root) e `T2` (contrato/geração) são independentes entre si. `T3` (infraestrutura do handler) só precisa do `StrictServerInterface` gerado por `T2`. `T4`-`T6` partem de `T3` e são independentes entre si (grupos de recurso distintos); `T7` parte de `T3` e também de `T1` (precisa de `Documents` já injetado em `financeiro.Module`). `financeiro/http.Handler` só satisfaz `StrictServerInterface` por completo depois de `T4`-`T7`; `Register` (`T8`) é o único ponto que exige a interface inteira, por isso fica por último.

## Task Breakdown

### T1: Composition root (`financeiro.New`)

**What**: `financeiro.New(Deps) *Module`, mesmo template de `identity.New`/`identity.Module`. Instancia `platform/storage.NewS3` e `platform/documents.Service` em `cmd/api/main.go` pela primeira vez (já existem desde `fundacao-documentos`, nunca antes chamados) e os injeta como `financeiro.Deps.Documents`.
**Where**: `api/internal/financeiro/module.go` (estende, não recria), `api/cmd/api/main.go` (extraído para uma função `boot` testável, mesmo padrão `run`/`deps` de `cmd/bootstrap-admin`), `api/internal/httpapi/router.go` (consequência técnica direta, não planejada originalmente: `Deps` ganha o campo `Financeiro *financeiro.Module`, inerte — `fin` precisa de algum destino para compilar; nenhuma rota é registrada, nenhum comportamento do roteador muda — ver `FIN-D-025`).
**Depends on**: None (consome só o que `01`-`05` já publicaram).
**Reuses**: `storage.NewS3`, `documents.Service`, `identity.New` como modelo.
**Requirement**: nenhum AC de `spec/06-api-http.md` diretamente — é infraestrutura para todos eles.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] `financeiro.Module` expõe os 2 repositórios e os 15 casos de uso como campos públicos, todos não-nil depois de `New`
- [x] `cmd/api/main.go` falha ao iniciar se `STORAGE_ENABLED` não for `true` quando `financeiro` é composto (`FIN-D-023`)
- [x] Nenhuma rota HTTP, nenhum contrato, nenhum handler ainda — só o composition root em Go
- [x] Gate check passes

**Tests**: unit (completude do wiring, 2 testes) + integration (primeira suíte de `cmd/api` — `boot_test.go` —, confirma que a API sobe com `financeiro` composto e `STORAGE_ENABLED=true`, e que falha claramente sem isso)
**Gate**: full
**Commit**: `feat(financeiro): composition root financeiro.New`

**Riscos/ambiguidades**: nenhuma decisão de negócio; `FIN-D-023` já resolvida na spec/design.

---

### T2: Contrato OpenAPI + geração de código

**What**: `api/openapi/financeiro.yaml` completo (as 15 operações, `API-01` a `API-05`), `api/openapi/codegen/financeiro.yaml`, `financeiro/http/generate.go`, `go generate` gerando `financeiro/http/api.gen.go` (`StrictServerInterface`, modelos, `GetSpec()`). Atualiza `web/redocly.yaml` com a entrada `financeiro@v1`.
**Where**: `api/openapi/financeiro.yaml`, `api/openapi/codegen/financeiro.yaml`, `api/internal/financeiro/http/generate.go`, `web/redocly.yaml`.
**Depends on**: None — independente de T1. Nenhum schema usa `x-go-type` fora de `Cents` (já em `common.yaml`), seguindo o padrão de `identity.yaml`; o contrato não precisa conhecer a forma de `financeiro.Module`.
**Reuses**: `common.yaml` (`Cents`, `Cursor`/`Limit` — não usados aqui, `FIN-D-021`, mas disponíveis se a decisão mudar —, `Problem`, `FieldError`, respostas comuns), `identity.yaml` como modelo de estilo.
**Requirement**: `API-01` a `API-05` (forma do contrato, não o comportamento ainda).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] As 15 operações estão no contrato, cada uma com `operationId`, `security` (nunca `[]`), parâmetros, `requestBody`/`responses` conforme a tabela de `design/06-api-http.md`
- [x] Toda resposta de erro usa o schema `Problem` de `common.yaml` — via `$ref: 'common.yaml#/components/responses/...'` nos casos genéricos, ou `$ref: '...#/components/schemas/Problem'` com descrição própria nos casos que precisam nomear o `code` específico (mesmo padrão já usado por `identity.yaml`, ex. `login`'s `401`) — nunca um `Problem` redefinido
- [x] `pnpm lint:api` passa sem erro
- [x] `go generate ./...` produz `financeiro/http/api.gen.go` sem erro, com `StrictServerInterface` contendo os 15 métodos
- [x] `pnpm gen:api` produz `web/lib/api/financeiro.d.ts`
- [x] Nenhum handler ainda — `api.gen.go` é só interface e modelos
- [x] Gate check passes (`go build ./...` compila o pacote `financeiro/http` mesmo sem nenhum `Handler` ainda, porque código gerado é autocontido) — exceto `httpapi`'s 2 testes de paridade rota×contrato, vermelhos por desenho até `T8` (`FIN-D-026`)

**Tests**: `redocly lint`, `openapi-typescript --check`, `go generate` sem erro
**Gate**: full + contract
**Commit**: `feat(financeiro): contrato openapi e geracao de codigo`

**Riscos/ambiguidades**: nenhuma — forma do contrato já fechada em `design/06-api-http.md`.

---

### T3: Infraestrutura do handler (`financeiro/http/handler.go`)

**What**: `Handler{M *financeiro.Module, Log *slog.Logger}`, `New(m, log)`, `writeError` (tabela completa de 21 mapeamentos de `FIN-D-022`), `Register` (esqueleto, sem nenhuma rota ainda — só a estrutura `ServerInterfaceWrapper`/`NewStrictHandlerWithOptions`).
**Where**: `api/internal/financeiro/http/handler.go`.
**Depends on**: T2 (precisa do `StrictServerInterface` gerado para a assinatura de `Register`, ainda que incompleto).
**Reuses**: `identity/http/handler.go` como modelo direto (`writeError`, `mapped`/`sentinels`).
**Requirement**: `API-06` (tratamento de erros uniforme).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Os 21 mapeamentos de erro de `FIN-D-022` estão implementados e testados individualmente
- [x] `conta_invalida` usa o status confirmado pelo mantenedor (`422`, salvo decisão em contrário — ver relatório de entrega)
- [x] Nenhum sentinel de `financeiro/domain` tem seu `.Error()` usado como `code` (todos vêm da tabela explícita); os de `platform/documents` continuam usando `.Error()` (já `snake_case`)
- [x] Gate check passes

**Tests**: unit — um teste por mapeamento de erro (21), confirmando status e code exatos
**Gate**: full
**Commit**: `feat(financeiro): infraestrutura do handler http e mapeamento de erros`

**Riscos/ambiguidades**: `conta_invalida` (404 vs. 422) é uma decisão técnica proposta, não travada — confirmar antes desta tarefa.

---

### T4: Handlers de contas (`API-01`)

**What**: `CreateConta`, `ListContas`, `RenameConta`, `DeactivateConta` — os 4 métodos de `StrictServerInterface` referentes a `financeiro:conta:*`.
**Where**: `api/internal/financeiro/http/contas_handler.go`.
**Depends on**: T3.
**Reuses**: `financeiro.Module.{CriarConta,ListarContas,RenomearConta,DesativarConta}`.
**Requirement**: `API-01` (AC1-AC5).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Os 4 ACs de `API-01` cobertos
- [x] Nenhuma verificação de permissão na camada HTTP — delegada ao caso de uso
- [x] `ListContas` devolve a lista completa, sem paginação (`FIN-D-021`)
- [x] Gate check passes

**Tests**: integration (requisição HTTP real contra o caso de uso real, mesmo padrão de `identity/http`)
**Gate**: full
**Commit**: `feat(financeiro): handlers http de contas`

---

### T5: Handlers de lançamentos e workflow (`API-02`, `API-03`)

**What**: `CreateLancamento`, `UpdateLancamento`, `CreateDevolucao`, `ListLancamentos`, `ReceiveLancamento`, `PayLancamento`, `CancelLancamento` — 7 métodos.
**Where**: `api/internal/financeiro/http/lancamentos_handler.go`.
**Depends on**: T3.
**Reuses**: `financeiro.Module.{CriarLancamento,EditarLancamento,CriarDevolucao,ListarLancamentos,ReceberLancamento,PagarLancamento,CancelarLancamento}`.
**Requirement**: `API-02` (AC1-AC5), `API-03` (AC1-AC5).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Os 10 ACs combinados de `API-02`+`API-03` cobertos
- [x] `UpdateLancamento` é `PUT`, substitui os 4 campos sempre (nunca parcial)
- [x] `CreateDevolucao` é endpoint próprio, nunca um campo opcional de `CreateLancamento`
- [x] `CancelLancamento` exige `reason` não vazio antes de qualquer chamada ao caso de uso (mesma ordem já garantida por `CancelarLancamento.Execute`)
- [x] `ListLancamentos` devolve a lista completa, sem paginação, incluindo `CANCELADA` (`FIN-D-021`)
- [x] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): handlers http de lancamentos e workflow`

---

### T6: Handler de saldo (`API-04`)

**What**: `GetSaldo` — 1 método.
**Where**: `api/internal/financeiro/http/saldo_handler.go`.
**Depends on**: T3.
**Reuses**: `financeiro.Module.ConsultarSaldo`.
**Requirement**: `API-04` (AC1).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `GET /financeiro/saldo` devolve `{saldo_cents}` correto, inclusive zero
- [ ] Nenhum parâmetro de filtro (nenhum existe no caso de uso)
- [ ] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): handler http de saldo`

---

### T7: Handlers de comprovantes (`API-05`)

**What**: `CreateComprovante` (multipart), `GetComprovanteUrl`, `ListComprovantes` — 3 métodos. Adiciona a chain `authenticatedUpload` em `httpapi/router.go` (`FIN-D-024`), só para a rota de anexar.
**Where**: `api/internal/financeiro/http/comprovantes_handler.go`, `api/internal/httpapi/router.go` (chain nova).
**Depends on**: T3, T1 (precisa de `financeiro.Module.Documents` já injetado).
**Reuses**: `financeiro.Module.{AnexarComprovante,ConsultarComprovante,ListarComprovantes}`.
**Requirement**: `API-05` (AC1-AC5).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `CreateComprovante` lê `multipart/form-data`, nunca valida extensão/tipo/tamanho por conta própria
- [ ] `GetComprovanteUrl` responde com `Cache-Control: no-store`
- [ ] `ListComprovantes` não gera URL, não audita (herdado de `ListarComprovantes`, nada novo na camada HTTP)
- [ ] A chain de upload usa um limite de corpo maior que o padrão de 1 MiB, acomodando os 10 MiB de `platform/documents` (`FIN-D-024`, valor confirmado antes desta tarefa)
- [ ] Todos os 5 erros de `platform/documents` (seção de erros de `FIN-D-022`) mapeados corretamente
- [ ] Gate check passes

**Tests**: integration, incluindo um upload real (Postgres+Garage/S3, mesmo padrão de `fundacao-documentos`/`04-comprovantes`)
**Gate**: full
**Commit**: `feat(financeiro): handlers http de comprovantes`

**Riscos/ambiguidades**: valor exato do limite de corpo (`FIN-D-024`) — confirmar antes desta tarefa.

---

### T8: Wiring final e testes de contrato

**What**: `financeirohttp.Register(...)` completo em `httpapi/router.go` (agora que `Handler` satisfaz `StrictServerInterface` por inteiro), `httpapi.Deps.Financeiro` conectado, verificação de paridade rota×contrato cobrindo `financeiro.yaml`. Testes de contrato/e2e cobrindo os 15 fluxos ponta a ponta.
**Where**: `api/internal/httpapi/router.go`, `api/cmd/api/main.go` (conclui o wiring iniciado em T1).
**Depends on**: T4, T5, T6, T7.
**Reuses**: `httpapi/router.go`'s `parity`/`routeKeys`/`operationKeys` (já existentes, cobrem qualquer spec adicionada à lista, sem alteração de código).
**Requirement**: todos os AC de `API-01` a `API-06`, agora ponta a ponta.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `httpapi.NewRouter` falha se rota e contrato de `financeiro` divergirem (mesma garantia de `AD-012`, agora também para `financeiro`)
- [ ] Um teste e2e por recurso (contas, lançamentos, workflow, saldo, comprovantes) — fluxo HTTP completo, sem acesso direto ao banco/casos de uso
- [ ] `pnpm lint:api`/`pnpm gen:api:check` seguem verdes com `financeiro.yaml` incluído
- [ ] Gate check passes (full + contract)

**Tests**: integration (e2e, mesmo padrão de `httpapi/e2e_test.go`/`router_integration_test.go`)
**Gate**: full + contract
**Commit**: `feat(financeiro): wiring final da api http e testes de contrato`

---

## Task Granularity Check

| Task | Scope | Status |
|---|---|---|
| T1 | 1 composition root + wiring de infraestrutura já pronta | ✅ Granular |
| T2 | 1 contrato + geração de código, unidade verificável por `lint:api`/`go generate` | ✅ Granular |
| T3 | 1 arquivo, infraestrutura compartilhada de erro — pré-requisito funcional de T4-T7, não uma tarefa de arquivo isolada | ✅ Aceitável |
| T4-T7 | 1 grupo de recurso cada (contas, lançamentos+workflow, saldo, comprovantes) — unidades arquiteturalmente verificáveis, não divisão por arquivo | ✅ Aceitável |
| T8 | 1 wiring final, inevitavelmente tocando `router.go`+`main.go` — única tarefa que pode ligar tudo, pois exige a interface completa | ✅ Aceitável |
