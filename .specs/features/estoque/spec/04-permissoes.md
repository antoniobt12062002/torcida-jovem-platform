# Permissões Specification

## Problem Statement

`01`-`03` implementam casos de uso que já chamam `Authz.Require` com nomes de permissão fixados (`estoque:produto:create` etc.), mas nenhuma `Contribution` publica essas permissões para a matriz de RBAC agregada (`identity/app.BuildMatrix`). Sem isso, nenhum papel realmente as recebe, e `authz.NewAuthorizer` nunca as reconhece como válidas. Esta sub-spec publica `estoque.Contribution()`, transversal às demais — mesmo papel que `05-permissoes` tem em `financeiro`.

## Goals

- [ ] As 6 permissões de `estoque` (`produto:create`, `produto:read`, `movimentacao:create`, `movimentacao:adjust`, `movimentacao:read`, `saldo:read`) existem na matriz de RBAC de produção.
- [ ] `ESTOQUE_LOJA` recebe as 6; `DIRETORIA` e `CONSELHO_FISCAL` recebem só as 3 de leitura; `PRESIDENTE` recebe tudo automaticamente (mecanismo já existente, inalterado); os demais papéis recebem nenhuma.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Qualquer nova permissão alem das 6 já fechadas (decisão do mantenedor) | Não há mais nenhuma operação em `01`-`03` sem permissão definida. |
| Alterar `forbiddenToConselhoFiscal` ou qualquer mecanismo de RBAC já existente | Reaproveitado como está — `CONSELHO_FISCAL` já não recebe `create`/`adjust`/`update`/`deactivate`/etc. por essa lista; as permissões de estoque seguem a mesma convenção sem precisar de nova entrada ali, porque `estoque` nunca concede `create`/`adjust` ao Conselho Fiscal na própria `Contribution`. |

---

## Assumptions & Open Questions

**Open questions:** none — a matriz completa (seção abaixo) foi aprovada pelo mantenedor (EST-D-004) antes desta sub-spec ser escrita.

---

## User Stories

### P1: Publicar a contribuição de RBAC de `estoque` ⭐ MVP

**User Story**: Como mantenedor do sistema, quero que `estoque.Contribution()` seja agregada à matriz de produção, para que as permissões de estoque funcionem de verdade em `cmd/api` e `cmd/bootstrap-admin`.

**Why P1**: Sem isso, toda chamada a `Authz.Require` em `01`-`03` falha (nenhum papel tem a permissão), mesmo que o código já exista.

**Acceptance Criteria**:

1. The system SHALL expose `estoque.Contribution()` no pacote raiz do módulo (`api/internal/estoque/module.go`), mesmo padrão de `financeiro.Contribution()`.
2. WHEN `estoque.Contribution()` é agregada a `identity/app.BuildMatrix` junto de `FoundationContributions()` e `financeiro.Contribution()` THEN o sistema SHALL conceder exatamente as 6 permissões a `ESTOQUE_LOJA`, as 3 de leitura a `DIRETORIA` e a `CONSELHO_FISCAL`, e nenhuma aos demais papéis (exceto `PRESIDENTE`, automático).
3. The system SHALL referenciar as constantes `Perm*` já declaradas pelos casos de uso de `01`-`03` em `estoque/app` (nunca strings literais independentes em `module.go`) — mesma decisão de `FIN-D-019`, pelo mesmo motivo (elimina drift entre o catálogo de RBAC e o que `Authz.Require` de fato verifica).

**Independent Test**: construir a matriz de produção (`FoundationContributions()`+`financeiro.Contribution()`+`estoque.Contribution()`), confirmar as permissões efetivas de cada papel.

---

## Edge Cases

- IF `estoque.Contribution()` for agregada sem `financeiro.Contribution()` (ex. um ambiente de teste que só monta `estoque`) THEN o sistema SHALL continuar funcionando — `BuildMatrix` já aceita qualquer subconjunto de `Contribution`s, sem dependência entre elas.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| PERM-01 | P1: Publicar a contribuição de RBAC | Implementing | Verified |

**Coverage:** 1 total, 1 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] A matriz de produção, com `estoque.Contribution()` agregada, concede exatamente a tabela de permissões aprovada (EST-D-004) — nem mais, nem menos.
