import type { Lancamento } from "./labels";

// Ações disponíveis para um lançamento, pelo status, pelo tipo e pelas
// permissões efetivas da sessão (FWB-03 AC4, FWB-04 AC6-7, casos de borda).
// A API continua sendo a autoridade; aqui só se evita oferecer o que ela
// recusaria.

export const LANCAMENTO_PERMISSIONS = {
  read: "financeiro:lancamento:read",
  create: "financeiro:lancamento:create",
  update: "financeiro:lancamento:update",
  receive: "financeiro:lancamento:receive",
  pay: "financeiro:lancamento:pay",
  cancel: "financeiro:lancamento:cancel",
} as const;

export type LancamentoAction = "editar" | "receber" | "pagar" | "cancelar" | "devolver";

/** Permissão exigida por cada ação. */
export const ACTION_PERMISSION: Readonly<Record<LancamentoAction, string>> = {
  editar: LANCAMENTO_PERMISSIONS.update,
  receber: LANCAMENTO_PERMISSIONS.receive,
  pagar: LANCAMENTO_PERMISSIONS.pay,
  cancelar: LANCAMENTO_PERMISSIONS.cancel,
  devolver: LANCAMENTO_PERMISSIONS.create,
};

function allowedByState({ tipo, status }: Pick<Lancamento, "tipo" | "status">, action: LancamentoAction): boolean {
  switch (action) {
    case "editar":
      return status === "CRIADA";
    case "receber":
      return tipo === "RECEITA" && status === "CRIADA";
    case "pagar":
      return tipo === "DESPESA" && status === "CRIADA";
    case "cancelar":
      return status !== "CANCELADA";
    case "devolver":
      return tipo === "RECEITA" && status === "RECEBIDA";
  }
}

const ORDER: readonly LancamentoAction[] = ["receber", "pagar", "editar", "devolver", "cancelar"];

export function lancamentoActions(
  lancamento: Pick<Lancamento, "tipo" | "status">,
  permissions: readonly string[],
): LancamentoAction[] {
  return ORDER.filter(
    (action) => allowedByState(lancamento, action) && permissions.includes(ACTION_PERMISSION[action]),
  );
}
