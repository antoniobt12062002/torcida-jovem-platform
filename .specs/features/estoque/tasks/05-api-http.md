# API HTTP — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/estoque/spec/05-api-http.md`
**Design**: `.specs/features/estoque/design/05-api-http.md`
**Status**: Approved (2026-10-03); tests-first, um commit atômico por tarefa.

**Pré-requisito**: `01`-`04` completas em `develop`.

**Nota de design, lida antes de iniciar T3**: diferente de `financeiro/06-api-http` (que precisou dos achados `FIN-D-026`/`FIN-D-027`/`FIN-D-029` porque os 15 métodos de `StrictServerInterface` foram implementados em 4 tarefas/commits separados), aqui os 6 métodos nascem **numa única tarefa** (`T4`). Isso muda a forma de `Register`:

- `T3` pode escrever `Register(r gin.IRoutes, server StrictServerInterface, h *Handler, chains Chains)` **já com as 6 rotas montadas no corpo** (formato literal de `identityhttp.Register`, nunca o formato dividido que `financeiro/FIN-D-027` precisou) — compila normalmente, porque nada chama `Register` com um `*Handler` concreto dentro de `T3` (só `T5` fará essa chamada real); o parâmetro `server StrictServerInterface` é só um tipo de interface, não exige nada do `*Handler` até o ponto de chamada.
- `T4`, ao terminar de implementar os 6 métodos, já deixa `*Handler` satisfazendo `StrictServerInterface` por completo **no mesmo commit** — por isso `T4` pode testar via um harness local que chama `estoquehttp.Register` de verdade (mesmo padrão de `identity/http/env_test.go`), **sem precisar do harness manual que `FIN-D-029` exigiu em `financeiro`** (lá, a interface só ficava completa depois de 4 tarefas).
- `httpapi/routes_test.go` compara contratos no disco × rotas registradas por `glob`, independente de como `Register` foi escrito — de `T2` (contrato criado) até `T4` (último antes da wiring real), `TestEveryRegisteredRouteHasAContractOperationAndViceVersa`/`TestRoutesAndContractFilesAreTheSameList` **vão falhar de verdade**, porque é `router.go`, não `estoque/http`, quem decide quando a rota é montada na aplicação real. Isso é uma **falha esperada de composição intermediária** (o servidor strict ainda está incompleto na composição real, mesma causa-raiz de `FIN-D-026`) — **nunca um gate verde**. Nenhuma task de `T2` a `T4` pode ser declarada concluída silenciando essa falha: o "Done when" de cada uma exige confirmar explicitamente que a falha observada é exatamente essa (e nenhuma outra) antes de seguir. A partir de `T5` (roteador completo), **todos os gates, sem exceção, devem estar verdes** — nenhuma falha é aceitável dali em diante.

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| Composition root (`estoque.New`) | unit | sem I/O, só confere que nenhum campo fica nil | `api/internal/estoque/module_new_test.go` | `go test ./...` |
| Contrato OpenAPI | `redocly lint` + `go generate` | confirma validade sem precisar de handler nenhum | — | `pnpm lint:api`, `go generate ./...` |
| `handler.go` (erro) | unit | 1:1 com a tabela de erro (8 mapeamentos) | `api/internal/estoque/http/handler_internal_test.go` | `go test ./...` |
| Handlers (6 métodos) | integration | requisição HTTP real contra o caso de uso real, via `estoquehttp.Register` de verdade (harness local, ver nota acima) | `api/internal/estoque/http/*_test.go` | `go test -tags=integration ./...` |
| Wiring final (`router.go`) | integration (e2e) | um teste por recurso (produtos, movimentações, ajustes+saldo), através do roteador real | `api/internal/httpapi/estoque_e2e_test.go` | `go test -tags=integration ./...` |

## Gate Check Commands

| Level | When | Command |
| --- | --- | --- |
| Full | T1, T5 | `cd api && go vet ./... && go vet -tags=integration ./... && go build ./... && go test ./... && go test -tags=integration ./...` — **sem nenhuma falha aceitável** |
| Full (com falha esperada de composição intermediária) | T2-T4 | mesmo comando — `TestEveryRegisteredRouteHasAContractOperationAndViceVersa`/`TestRoutesAndContractFilesAreTheSameList` (`internal/httpapi`) **vão falhar de verdade** nessas 3 tarefas; a task só pode ser concluída depois de confirmar que a falha é exatamente essas 2 e nenhuma outra. Qualquer outra falha bloqueia a tarefa. A partir de `T5`, esta linha deixa de existir — só a linha `Full` acima se aplica. |
| Contract | T2 e T5 | `cd web && pnpm lint:api && pnpm gen:api:check` |
| Lint | toda tarefa | `golangci-lint run --path-mode=abs --build-tags=integration` |
| Architecture | toda tarefa | `go test ./internal/... -run TestProductionCodeRespectsTheModuleBoundaries` |

---

## Execution Plan

```
T2 → T3
T3 → T4
T1 → T4
T4 → T5
```

---

## Task Breakdown

### T1: Composition root (`estoque.New`)

**What**: `estoque.New(Deps) *Module` — 2 repositórios + 6 casos de uso, mesmo template de `financeiro.New`. `cmd/api/main.go` instancia e passa para `httpapi.Deps`.
**Where**: `api/internal/estoque/module.go` (estende — já tem `Contribution()` de `04-permissoes`), `api/cmd/api/main.go`, `api/internal/httpapi/router.go` (`Deps` ganha `Estoque *estoque.Module`, inerte até `T5` — mesma razão mecânica de `FIN-D-025`, campo de struct nunca usado não é erro de compilação).
**Depends on**: None (consome só o que `01`-`04` já publicaram)
**Reuses**: `financeiro.New`/`cmd/api/main.go`'s wiring de `financeiro` como modelo direto — mais simples, pois `estoque` não usa `platform/documents`/`storage`.
**Requirement**: infraestrutura para API-01 a API-04.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `estoque.Module` expõe os 2 repositórios e os 6 casos de uso, todos não-nil depois de `New`
- [ ] Nenhuma rota HTTP, nenhum contrato, nenhum handler ainda
- [ ] Gate check passes

**Tests**: unit (completude do wiring)
**Gate**: Full
**Commit**: `feat(estoque): composition root estoque.New`

---

### T2: Contrato OpenAPI + geração de código

**What**: `api/openapi/estoque.yaml` completo (as 6 operações, API-01 a API-04), `api/openapi/codegen/estoque.yaml`, `estoque/http/generate.go`, `go generate` gerando `estoque/http/api.gen.go`. Atualiza `web/redocly.yaml` com `estoque@v1`.
**Where**: `api/openapi/estoque.yaml`, `api/openapi/codegen/estoque.yaml`, `api/internal/estoque/http/generate.go`, `web/redocly.yaml`
**Depends on**: None — independente de T1.
**Reuses**: `financeiro.yaml` como modelo de estilo (`Cents` não se aplica — `estoque` não tem campo monetário no V1; `Cursor`/`Limit`, `Problem`, respostas comuns de `common.yaml` via `$ref`).
**Requirement**: API-01 a API-04 (forma do contrato).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] As 6 operações no contrato, cada uma com `operationId`, `security` (nunca `[]`), parâmetros, `requestBody`/`responses`
- [ ] `movimentacao_de_id` é campo opcional no schema (nunca `oneOf`/discriminador) — conforme design
- [ ] `pnpm lint:api` passa sem erro
- [ ] `go generate ./...` produz `estoque/http/api.gen.go` com `StrictServerInterface` contendo os 6 métodos
- [ ] `pnpm gen:api` produz `web/lib/api/estoque.d.ts`
- [ ] Gate check `Full` executado por completo; confirmado que a única falha é `TestEveryRegisteredRouteHasAContractOperationAndViceVersa`/`TestRoutesAndContractFilesAreTheSameList` (falha esperada de composição intermediária, não um gate verde) — qualquer outra falha bloqueia esta tarefa

**Tests**: `redocly lint`, `openapi-typescript --check`, `go generate` sem erro
**Gate**: Full + Contract
**Commit**: `feat(estoque): contrato openapi e geracao de codigo`

---

### T3: Infraestrutura do handler (`estoque/http/handler.go`)

**What**: `Handler{M *estoque.Module, Log *slog.Logger}`, `New(m, log)`, `writeError` (tabela completa de 8 mapeamentos), `sessionOf`/`ginContext`/`requestContext` (copiados, mesma razão de `FIN-D-028`), `Register(r gin.IRoutes, server StrictServerInterface, h *Handler, chains Chains)` **já com as 6 rotas montadas** (ver nota de design no topo deste arquivo — aqui `Register` pode ter o formato completo de `identityhttp.Register` desde já, porque nada o chama antes de `T5`).
**Where**: `api/internal/estoque/http/handler.go`
**Depends on**: T2
**Reuses**: `financeiro/http/handler.go` e `identity/http/handler.go` como modelo direto.
**Requirement**: API-04.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Os 8 mapeamentos de erro implementados e testados individualmente
- [ ] Nenhum sentinel de `estoque/domain` tem seu `.Error()` usado como `code`
- [ ] `Register` compila com as 6 chamadas `r.POST`/`r.GET` já escritas, referenciando os métodos ainda inexistentes de `w.<Método>` do wrapper gerado — **isso é esperado e compila**, porque `w` vem de `NewStrictHandlerWithOptions(server, ...)` e `server` é só um parâmetro de interface; só a função que efetivamente chamar `Register` com um `*Handler` concreto (em `T5`) exige a interface completa
- [ ] Gate check `Full` executado por completo; confirmado que a única falha é `TestEveryRegisteredRouteHasAContractOperationAndViceVersa`/`TestRoutesAndContractFilesAreTheSameList` (falha esperada de composição intermediária) — qualquer outra falha bloqueia esta tarefa

**Tests**: unit — um teste por mapeamento de erro (8)
**Gate**: Full
**Commit**: `feat(estoque): infraestrutura do handler http e mapeamento de erros`

---

### T4: Handlers (6 métodos)

**What**: `CreateProduto`, `ListProdutos`, `CreateMovimentacao`, `ListMovimentacoes`, `CreateAjuste`, `GetSaldo` — os 6 métodos de `StrictServerInterface`, delegando inteiramente aos 6 casos de uso.
**Where**: `api/internal/estoque/http/produtos_handler.go`, `api/internal/estoque/http/movimentacoes_handler.go`, `api/internal/estoque/http/ajustes_saldo_handler.go`
**Depends on**: T3, T1 (precisa de `estoque.Module` já composto)
**Reuses**: `financeiro/http/{contas,lancamentos,saldo}_handler.go` como modelo direto.
**Requirement**: API-01, API-02, API-03.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Os ACs combinados de API-01/API-02/API-03 cobertos
- [ ] Nenhuma verificação de permissão na camada HTTP — delegada ao caso de uso
- [ ] `CreateMovimentacao` valida a presença de `movimentacao_de_id` quando `tipo=DEVOLUCAO` dentro do caso de uso (já implementado em `02-movimentacoes/T4`), nunca na camada HTTP
- [ ] Ao final desta tarefa, `*Handler` satisfaz `StrictServerInterface` por completo — confirmado compilando um teste local que chama `estoquehttp.Register(h, h, chains)` de verdade (ver nota de design)
- [ ] Gate check `Full` executado por completo; confirmado que a única falha é `TestEveryRegisteredRouteHasAContractOperationAndViceVersa`/`TestRoutesAndContractFilesAreTheSameList` (falha esperada de composição intermediária — `router.go` ainda não chama `Register`) — qualquer outra falha bloqueia esta tarefa

**Tests**: integration (requisição HTTP real contra o caso de uso real, via `Register` de verdade — harness local, não manual)
**Gate**: Full
**Commit**: `feat(estoque): handlers http de produtos, movimentacoes, ajustes e saldo`

---

### T5: Wiring final e testes de contrato

**What**: `router.go` chama `estoquehttp.New`+`Register` de verdade, monta as 6 rotas (todas atrás da chain `authenticated` — nenhuma upload, nenhuma chain dedicada), `estoqueSpec` somado ao validador de contrato e à checagem de paridade. `httpapi.Deps.Estoque` passa a ser lido de fato.
**Where**: `api/internal/httpapi/router.go`, `api/internal/httpapi/router_test.go`/`router_integration_test.go`/`e2e_test.go` (campo `Estoque` nos `Deps{}` de teste, mesma consequência mecânica de `financeiro/06-api-http/T8`), `api/internal/httpapi/routes_test.go` (só a contagem fixa de rotas, de 31 para 37)
**Depends on**: T4
**Reuses**: `httpapi/router.go`'s `parity`/`routeKeys`/`operationKeys` (já existentes, cobrem qualquer spec adicionada à lista).
**Requirement**: todos os AC de API-01 a API-04, agora ponta a ponta.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `httpapi.NewRouter` falha se rota e contrato de `estoque` divergirem
- [ ] Um teste e2e por recurso (produtos, movimentações, ajustes+saldo) através do roteador real
- [ ] Os 2 testes de paridade voltam a passar sem exceção
- [ ] `pnpm lint:api`/`pnpm gen:api:check` seguem verdes com `estoque.yaml` incluído
- [ ] Gate check passes (Full + Contract, sem nenhuma exceção remanescente)

**Tests**: integration (e2e, mesmo padrão de `httpapi/financeiro_e2e_test.go`)
**Gate**: Full + Contract
**Commit**: `feat(estoque): wiring final da api http e testes de contrato`

---

## Phase Execution Map

```
T2 → T3
T3 → T4
T1 → T4
T4 → T5
```

---

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | 1 composition root + wiring de infraestrutura já pronta | ✅ Granular |
| T2 | 1 contrato + geração de código | ✅ Granular |
| T3 | 1 arquivo, infraestrutura compartilhada de erro + rotas | ✅ Aceitável (mesmo padrão de `financeiro/06-api-http/T3`) |
| T4 | 3 arquivos, 6 métodos — um único grupo coeso, volume 2.5× menor que `financeiro` justifica 1 tarefa em vez de 4 | ✅ Aceitável |
| T5 | 1 wiring final, inevitavelmente tocando `router.go` + 3 arquivos de teste existentes | ✅ Aceitável |

---

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | None | None | ✅ Match |
| T3 | T2 | T2 | ✅ Match |
| T4 | T3, T1 | T3, T1 | ✅ Match |
| T5 | T4 | T4 (e T1 transitivamente) | ✅ Match |

---

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1: Composition root | `module.go` | unit | unit | ✅ OK |
| T2: Contrato | — | lint/generate | lint/generate | ✅ OK |
| T3: Handler infra | `http` (erro) | unit | unit | ✅ OK |
| T4: Handlers | `http` (métodos) | integration | integration | ✅ OK |
| T5: Wiring final | `httpapi` | integration (e2e) | integration | ✅ OK |
