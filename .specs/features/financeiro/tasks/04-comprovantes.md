# Comprovantes — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/financeiro/spec/04-comprovantes.md`
**Design**: `.specs/features/financeiro/design/04-comprovantes.md`
**Status**: Approved (2026-09-29); `tests-first`, um commit atômico por tarefa.

**Pré-requisito cruzado**: `T1` precisa da migração de `02-lancamentos` (para validar existência do `lancamento_id`) e de `platform/documents` (já concluído e validado, `fundacao-documentos`).

## Test Coverage Matrix

| Layer | Tests | Notes |
|---|---|---|
| Casos de uso (`app`) | integration | Postgres + Garage (S3) reais via testcontainers, mesma infraestrutura de `fundacao-documentos` |

## Gate Check Commands

| Level | When | Command |
|---|---|---|
| Full | fim de cada tarefa | `cd api && go vet ./... && go test -tags=integration ./...` |

## Execution Plan

```
T1 → T2
T1 → T3
```

## Task Breakdown

### T1: `AnexarComprovante`

**What**: Anexa um documento a um lançamento via `documents.Service.Store`.
**Where**: `api/internal/financeiro/comprovantes/service.go`
**Depends on**: None
**Reuses**: `platform/documents.Service`, `platform/authz`
**Requirement**: CMP-01 (AC1-AC4)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Chama `Store` com `owner_type="financeiro.lancamento"`, `owner_id`, `RequiredPermission="financeiro:comprovante:create"`
- [ ] Lançamento inexistente: `lancamento_nao_encontrado`, antes de chamar `Store`
- [ ] Aceita anexar em lançamento `CANCELADA` sem restrição
- [ ] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona anexo de comprovante a lancamento`

---

### T2: `ConsultarComprovante`

**What**: Gera URL assinada via `documents.Service.AccessURL`.
**Where**: `api/internal/financeiro/comprovantes/service.go`
**Depends on**: T1
**Reuses**: idem T1
**Requirement**: CMP-02 (AC1, AC2)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Retorna URL assinada funcional (download real, roundtrip)
- [ ] Sem `financeiro:comprovante:read`, recusa, nenhuma URL
- [ ] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona consulta de comprovante`

---

### T3: `ListarComprovantes`

**What**: Lista comprovantes de um lançamento via `documents.Service.ListByOwner`, sem gerar URL nem auditar, sem alterar `status`.
**Where**: `api/internal/financeiro/comprovantes/service.go`
**Depends on**: T1
**Reuses**: idem T1
**Requirement**: CMP-03 (AC1-AC3)

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Lista todos os comprovantes do lançamento, mais novo primeiro
- [ ] Não gera URL, não audita
- [ ] Não altera `status` do lançamento (prova negativa: `status` inalterado antes/depois)
- [ ] Gate check passes

**Tests**: integration
**Gate**: full
**Commit**: `feat(financeiro): adiciona listagem de comprovantes por lancamento`

---

## Task Granularity Check

| Task | Scope | Status |
|---|---|---|
| T1-T3 | 1 arquivo, 1 caso de uso cada | ✅ Granular |
