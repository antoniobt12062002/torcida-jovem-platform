import type { User } from "./hooks";

// Ações de linha oferecidas conforme a permissão e o estado do usuário
// (USR-04 AC6). Na linha da própria pessoa autenticada nenhuma ação aparece:
// a API recusaria com self_change_forbidden. A API continua sendo a autoridade.

export type UserAction =
  | "deactivate"
  | "reactivate"
  | "grant"
  | "setRoles"
  | "revoke"
  | "resetPassword";

export const ACTION_LABELS: Readonly<Record<UserAction, string>> = {
  deactivate: "Desativar",
  reactivate: "Reativar",
  grant: "Promover a administrativo",
  setRoles: "Editar papéis",
  revoke: "Retirar acesso administrativo",
  resetPassword: "Gerar senha temporária",
};

export function availableActions(
  user: User,
  sessionUserId: string | null | undefined,
  permissions: readonly string[],
): UserAction[] {
  if (sessionUserId != null && user.id === sessionUserId) return [];
  const can = (permission: string) => permissions.includes(permission);
  const isAdmin = user.admin_membership != null;
  const actions: UserAction[] = [];
  if (!isAdmin && can("identity:admin:grant")) actions.push("grant");
  if (isAdmin && can("identity:role:assign")) actions.push("setRoles");
  if (isAdmin && can("identity:admin:revoke")) actions.push("revoke");
  if (can("identity:user:reset_password")) actions.push("resetPassword");
  if (can("identity:user:update")) actions.push(user.active ? "deactivate" : "reactivate");
  return actions;
}
