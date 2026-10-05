import type { Lancamento } from "./labels";

// Ações disponíveis para um lançamento, pelo status, pelo tipo e pelas
// permissões efetivas da sessão (FWB-03 AC4). A API continua sendo a
// autoridade; aqui só se evita oferecer o que ela recusaria.

export const LANCAMENTO_PERMISSIONS = {
  read: "financeiro:lancamento:read",
  create: "financeiro:lancamento:create",
  update: "financeiro:lancamento:update",
} as const;

export type LancamentoAction = "editar";

export function lancamentoActions(
  lancamento: Pick<Lancamento, "tipo" | "status">,
  permissions: readonly string[],
): LancamentoAction[] {
  const can = (permission: string) => permissions.includes(permission);
  const actions: LancamentoAction[] = [];
  if (lancamento.status === "CRIADA" && can(LANCAMENTO_PERMISSIONS.update)) actions.push("editar");
  return actions;
}
