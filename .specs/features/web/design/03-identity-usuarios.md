# Identity — Administração de Usuários (Web) — Design

**Spec**: `.specs/features/web/spec/03-identity-usuarios.md`
**Pré-requisito**: sub-spec 01 completa (F5 integrada).

## Architecture Overview

```
app/(app)/admin/usuarios/page.tsx
  └─ features/identity/usuarios/users-page.tsx
        ├─ users-table.tsx  (linhas + menu de ações por permissão e estado)
        ├─ dialogs/*.tsx    (criar, desativar, reativar, promover, papéis, retirar, senha temporária)
        └─ hooks.ts         (useInfiniteQuery + mutações) ─→ lib/api/client.ts (identity)
features/identity/nav.ts    (item "Usuários", permissão identity:user:read)
```

## Code Reuse Analysis

### Existing Components to Leverage

| Existente | Uso |
| --- | --- |
| `components/app` (`RequirePermission`, estados, `ConfirmDialog` com motivo, `StatusBadge`, `format.ts`) | página, ações, confirmações |
| `components/ui` (`table`, `dropdown-menu`, `dialog`, `checkbox`, `select`) | lista, menu, papéis, filtros |
| `lib/api/query-keys.ts` (`identity.users(filters)`) | cache e invalidação |
| `lib/forms` | erros de criação |

### Integration Points

| Ponto | Como |
| --- | --- |
| Navegação | `features/identity/nav.ts` exporta os itens; a INT liga. |
| Sessão | `useSession().context.user.id` para o caso de borda da própria linha. |

## Components

| Arquivo | Função |
| --- | --- |
| `features/identity/nav.ts` | `identityNav: NavItem[]` |
| `features/identity/usuarios/hooks.ts` | `useUsers(filters)` com `useInfiniteQuery` (`limit`, `cursor`); mutações com invalidação de `identity.users` |
| `features/identity/usuarios/roles.ts` | Lista e rótulos pt-BR dos papéis administrativos |
| `features/identity/usuarios/errors.ts` | Catálogo `code → mensagem` |
| `features/identity/usuarios/users-page.tsx` | Filtros, tabela, "Carregar mais", "Novo usuário" |
| `features/identity/usuarios/users-table.tsx` | Linhas e menu de ações (USR-04 AC6 e casos de borda) |
| `features/identity/usuarios/dialogs/create-user-dialog.tsx` | USR-02 |
| `features/identity/usuarios/dialogs/status-dialogs.tsx` | USR-03 |
| `features/identity/usuarios/dialogs/admin-dialogs.tsx` | USR-04 |
| `features/identity/usuarios/dialogs/temporary-password-dialog.tsx` | USR-05 |
| `app/(app)/admin/usuarios/page.tsx` | `RequirePermission("identity:user:read")` + `UsersPage` |

## Data Models

`User`, `UserPage`, `CreateUserRequest`, `SetRolesRequest`, `GrantAdminRequest`, `RevokeAdminRequest`, `ResetUserPasswordRequest`, `TemporaryPassword` dos `.d.ts`.

## Error Handling Strategy

Catálogo em `errors.ts` com uma mensagem por `code` de USR-02 a USR-05. Exemplos: `last_admin` → "Não é possível: o sistema ficaria sem ninguém que possa conceder acesso."; `privilege_escalation` → "Você não pode conceder ou alterar permissões que não tem."; `self_change_forbidden` → "Você não pode fazer isso com a sua própria conta.". O restante segue FND-06.

## Risks & Concerns

| Risco | Mitigação |
| --- | --- |
| Senha temporária ficar no cache | A mutação não grava no `QueryClient`; o valor vive só no estado do diálogo; `gcTime` irrelevante porque não é consulta. Teste confere. |
| Lista desatualizada depois de uma ação | Invalidação de todas as chaves `identity.users(*)`. |

## Tech Decisions

| Decisão | Motivo |
| --- | --- |
| Diálogos em vez de página de detalhe | Sem `GET /users/{id}`; a linha tem tudo. |
| Motivo mínimo de 10 caracteres conferido no front para promoção e senha temporária | Evita ida à API para um erro previsível; a API continua conferindo (`reason_required`). |
