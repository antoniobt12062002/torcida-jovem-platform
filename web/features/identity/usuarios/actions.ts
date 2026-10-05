import type { User } from "./hooks";

// Ações de linha oferecidas conforme a permissão e o estado do usuário. Na
// linha da própria pessoa autenticada nenhuma ação aparece: a API recusaria
// com self_change_forbidden. A API continua sendo a autoridade.

export type UserAction = "deactivate" | "reactivate";

export const ACTION_LABELS: Readonly<Record<UserAction, string>> = {
  deactivate: "Desativar",
  reactivate: "Reativar",
};

export function availableActions(
  user: User,
  sessionUserId: string | null | undefined,
  permissions: readonly string[],
): UserAction[] {
  if (sessionUserId != null && user.id === sessionUserId) return [];
  const can = (permission: string) => permissions.includes(permission);
  const actions: UserAction[] = [];
  if (can("identity:user:update")) actions.push(user.active ? "deactivate" : "reactivate");
  return actions;
}
