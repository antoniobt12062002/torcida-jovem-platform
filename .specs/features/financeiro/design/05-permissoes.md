# Permissões — Design

**Spec**: `.specs/features/financeiro/spec/05-permissoes.md`
**Status**: Approved (2026-09-29). Sem entidade nova — é a `Contribution` publicada para `identity.app.BuildMatrix`.

## Architecture Overview

```mermaid
graph TD
    F["financeiro.Contribution()<br/>(substitui o placeholder de roles_matrix.go)"] --> BM[identity.app.BuildMatrix]
    BM --> M[(matriz agregada)]
    M --> AZ[platform/authz.Require, usado por 01-04]
```

O conteúdo desta `Contribution` é preenchido incrementalmente: cada uma de `01`-`04`, na sua própria fase de tarefas, adiciona as permissões que sua própria operação introduz — nunca antecipadas aqui.

## Code Reuse Analysis

| Component | Location | How to Use |
|---|---|---|
| Agregação da matriz | `identity/app/roles_matrix.go` (`BuildMatrix`, `Contribution`) | `financeiro` publica uma `Contribution` no mesmo formato já usado por `identity` e `audit` |
| Placeholder a substituir | `identity/app/roles_matrix.go:132-144` (`FoundationContributions`) | Removido quando `financeiro.Contribution()` existir; `cmd/api/main.go` e `cmd/bootstrap-admin/main.go` passam a incluir `financeiro.Contribution()` no lugar |

## Components

### `financeiro` (pacote raiz do módulo)

- **Interface**: `func Contribution() app.Contribution` — devolve `Module: "financeiro"`, a lista de `authz.Definition` e o mapa `Grants` por papel.
- **Dependencies**: `identity/app` apenas — o tipo `Contribution` e o tipo/constantes de papel re-exportados (`app.Role`, `app.RoleConselhoFiscal` etc., `FIN-D-018`). **Nunca** `identity/domain` diretamente: `architecture_test.go` proíbe um módulo de importar o `domain` de outro (só `app` e o pacote raiz são permitidos) — achado durante a análise de `05-permissoes`, corrigido por `FIN-D-018` antes de `T1` ser implementada.

## Tech Decisions (only non-obvious ones)

| Decision | Choice | Rationale |
|---|---|---|
| Onde a `Contribution` é montada | Uma função no pacote raiz de `financeiro` (`FIN-D-017`) | `app/` contém casos de uso e portas, não a contribuição do módulo para o RBAC — `identity.New`/`FoundationContributions` não é o mesmo padrão (pacotes diferentes: `identity` vs. `identity/app`) |
| Como `financeiro` obtém o vocabulário de papéis | `identity/app` reexporta `Role` (alias de `identity/domain.Role`) e as 8 constantes (`FIN-D-018`) | `financeiro` não pode importar `identity/domain` (fronteira de módulo); `identity/app` já é a superfície pública usada para `Contribution` |
| Quando o placeholder é removido | Como parte da última tarefa desta spec, depois que `01`-`04` já contribuíram suas permissões | Evita remover o placeholder antes de ter o substituto completo — ver `financeiro/STATE.md`, ordem real de execução |
