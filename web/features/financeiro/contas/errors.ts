import { MESSAGE_FORBIDDEN } from "@/components/app/api-error";
import type { components } from "@/lib/api/financeiro";
import type { ErrorCatalog } from "@/lib/api/problem";

// Catálogo de erros do plano de contas (FWB-01 AC4 e AC6; design 04).

export const contasCatalog: ErrorCatalog = {
  conta_ja_utilizada: "Esta conta já foi usada em um lançamento e não pode ser renomeada.",
  conta_tipo_incompativel: "A subconta precisa ter o mesmo tipo da conta pai.",
  conta_nao_encontrada: "Conta não encontrada. Atualize a página e tente de novo.",
};

/** Catálogo dos formulários: inclui o 403 forbidden (FND-04 AC5). */
export const contasFormCatalog: ErrorCatalog = {
  ...contasCatalog,
  forbidden: MESSAGE_FORBIDDEN,
};

export const tipoContaLabel: Record<components["schemas"]["TipoConta"], string> = {
  RECEITA: "Receita",
  DESPESA: "Despesa",
};
