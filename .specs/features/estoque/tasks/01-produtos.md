# Produtos (SKU de Estoque) — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/estoque/spec/01-produtos.md`
**Design**: `.specs/features/estoque/design/01-produtos.md`
**Status**: Approved (2026-10-03); tests-first, um commit atômico por tarefa.

**Pré-requisito**: nenhum — primeira sub-spec de `estoque`, sem dependência de outro módulo além de `platform` (já mesclado).

## Test Coverage Matrix

> Gerado por amostragem de `financeiro/{01-plano-de-contas,02-lancamentos}` — mesmas convenções.

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `domain` (entidade+sentinels) | none | build gate only | `api/internal/estoque/domain/*.go` | gate `Full` |
| `infra` (repositório) | integration | caminhos de consulta principais + tradução de erro (`UNIQUE`→`ErrCodigoDuplicado`) | `api/internal/estoque/infra/*_test.go` (`//go:build integration`) | `go test -tags=integration ./...` |
| `app` (caso de uso) | unit (com doubles) | 1:1 com os ACs de PRD-01/PRD-02, incluindo todo edge case listado | `api/internal/estoque/app/*_test.go` | `go test ./...` |

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

### T1: Migration `produtos_estoque`

**What**: `000008_estoque_produtos.up.sql`/`.down.sql` — tabela `produtos_estoque` (`id`, `codigo` único, `nome`, `unidade_medida`, `criado_em`), sem `ativo`/`atualizado_em` (`EST-D-007`).
**Where**: `api/migrations/000008_estoque_produtos.up.sql`, `api/migrations/000008_estoque_produtos.down.sql`
**Depends on**: None
**Reuses**: estrutura de `000006_financeiro_contas` como modelo (mesmo estilo de migration do projeto).
**Requirement**: infraestrutura para PRD-01/PRD-02.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] `up` cria a tabela com `codigo UNIQUE NOT NULL`
- [x] `down` remove a tabela, sem deixar órfão nenhuma outra migration
- [x] Teste de migração confirma up+down em sequência (mesmo padrão de `platform/database`'s testes de migration)
- [x] Gate check passes

**Tests**: integration (teste de migração, Postgres real)
**Gate**: Full
**Commit**: `feat(estoque): migration de produtos de estoque`

---

### T2: `domain.Produto` e sentinels

**What**: `Produto{ID, Codigo, Nome, UnidadeMedida, CriadoEm}`, `ErrProdutoNaoEncontrado`, `ErrCodigoDuplicado`.
**Where**: `api/internal/estoque/domain/produto.go`
**Depends on**: None (não depende da migration para compilar; só para ser exercitado de verdade em T3)
**Reuses**: formato de `financeiro/domain.Conta`.
**Requirement**: PRD-01, PRD-02.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [x] Struct e os 2 sentinels declarados, sem nenhum método de comportamento (entidade simples)
- [x] Gate check passes

**Tests**: none (entidade simples, build gate only — consistente com a Coverage Expectation da matriz)
**Gate**: Full
**Commit**: `feat(estoque): entidade de produto e sentinels de dominio`

---

### T3: `infra.ProdutoRepository`

**What**: `Criar`, `Buscar`, `Listar` sobre `produtos_estoque`; traduz violação de `UNIQUE` em `domain.ErrCodigoDuplicado`.
**Where**: `api/internal/estoque/infra/produto_repository.go`
**Depends on**: T1, T2
**Reuses**: estilo de `financeiro/infra.ContaRepository`; tradução de erro de `UNIQUE`, estilo de `identity/infra.UserRepository` para e-mail duplicado.
**Requirement**: PRD-01, PRD-02.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `Criar` traduz violação de `codigo` duplicado em `domain.ErrCodigoDuplicado`, nunca o erro de driver cru
- [ ] `Buscar` devolve `domain.ErrProdutoNaoEncontrado` quando o id não existe
- [ ] Teste de integração com duas criações concorrentes do mesmo `codigo`: só uma conclui, a outra recebe `ErrCodigoDuplicado`
- [ ] Gate check passes

**Tests**: integration (Postgres real via `testutil.NewTestDB`)
**Gate**: Full
**Commit**: `feat(estoque): repositorio de produtos`

---

### T4: `app.CriarProduto`

**What**: implementa PRD-01 — permissão, validação de campo vazio, persistência (sem `Tx`/transação composta — é uma única escrita, mesmo padrão de `financeiro/app.CriarConta` quando não há passo adicional a amarrar na mesma transação além da auditoria).
**Where**: `api/internal/estoque/app/criar_produto.go`
**Depends on**: T3
**Reuses**: template de `financeiro/app.CriarConta`.
**Requirement**: PRD-01.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Os 4 ACs de PRD-01 cobertos (sucesso, código duplicado, campo vazio, sem permissão)
- [ ] Auditoria `produto.create` registrada na mesma transação da criação
- [ ] Gate check passes

**Tests**: unit (doubles para `Authz`/`Produtos`/`Audit`/`Tx`)
**Gate**: Full
**Commit**: `feat(estoque): caso de uso criar produto`

---

### T5: `app.ListarProdutos`

**What**: implementa PRD-02 — consulta pura, sem `Tx`/`Audit`.
**Where**: `api/internal/estoque/app/listar_produtos.go`
**Depends on**: T3
**Reuses**: template de `financeiro/app.ListarContas`.
**Requirement**: PRD-02.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Os 2 ACs de PRD-02 cobertos (sucesso sem paginação, sem permissão)
- [ ] Gate check passes

**Tests**: unit (doubles)
**Gate**: Full
**Commit**: `feat(estoque): caso de uso listar produtos`

---

## Phase Execution Map

```
T1 → T3
T2 → T3
T3 → T4
T3 → T5
```

(T1/T2 são independentes entre si; T3 depende dos dois; T4/T5 dependem só de T3 e são independentes entre si.)

---

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1: Migration | 1 par up/down | ✅ Granular |
| T2: Entidade+sentinels | 1 arquivo | ✅ Granular |
| T3: Repositório | 1 arquivo | ✅ Granular |
| T4: Caso de uso CriarProduto | 1 arquivo | ✅ Granular |
| T5: Caso de uso ListarProdutos | 1 arquivo | ✅ Granular |

---

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | None | None | ✅ Match |
| T3 | T1, T2 | T1, T2 | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | T3 | T3 | ✅ Match |

---

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1: Migration | — | integration (teste de migração) | integration | ✅ OK |
| T2: Entidade | `domain` | none | none | ✅ OK |
| T3: Repositório | `infra` | integration | integration | ✅ OK |
| T4: CriarProduto | `app` | unit | unit | ✅ OK |
| T5: ListarProdutos | `app` | unit | unit | ✅ OK |
