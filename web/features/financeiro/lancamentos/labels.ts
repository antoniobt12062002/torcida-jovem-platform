import { ArrowDownToLineIcon, ArrowUpFromLineIcon, BanIcon, CircleDashedIcon } from "lucide-react";

import type { StatusDef } from "@/components/app/status-badge";
import type { components } from "@/lib/api/financeiro";

// Tipos do contrato e rótulos em português para as telas de lançamentos.

export type Conta = components["schemas"]["Conta"];
export type ContaList = components["schemas"]["ContaList"];
export type Lancamento = components["schemas"]["Lancamento"];
export type LancamentoList = components["schemas"]["LancamentoList"];
export type TipoConta = components["schemas"]["TipoConta"];
export type FormaPagamento = components["schemas"]["FormaPagamento"];
export type StatusLancamento = components["schemas"]["StatusLancamento"];

export const TIPO_LABEL: Readonly<Record<TipoConta, string>> = {
  RECEITA: "Receita",
  DESPESA: "Despesa",
};

export const FORMA_PAGAMENTO_LABEL: Readonly<Record<FormaPagamento, string>> = {
  PIX: "Pix",
  CARTAO: "Cartão",
  DINHEIRO: "Dinheiro",
  TRANSFERENCIA: "Transferência",
  OUTROS: "Outros",
};

export const STATUS_LABEL: Readonly<Record<StatusLancamento, string>> = {
  CRIADA: "Criada",
  RECEBIDA: "Recebida",
  PAGA: "Paga",
  CANCELADA: "Cancelada",
};

/**
 * Status do lançamento com tom e ícone (DSM-03 AC3, DD-3 de WDS-D-026): rótulo
 * e ícone distintos entre si. `RECEBIDA` e `PAGA` dividem o tom de sucesso e se
 * distinguem pelo ícone de direção (entrada e saída); `CANCELADA` é neutra,
 * porque cancelamento não é erro.
 */
export const STATUS_DEF: Readonly<Record<StatusLancamento, StatusDef>> = {
  CRIADA: { label: STATUS_LABEL.CRIADA, tone: "neutral", icon: CircleDashedIcon },
  RECEBIDA: { label: STATUS_LABEL.RECEBIDA, tone: "success", icon: ArrowDownToLineIcon },
  PAGA: { label: STATUS_LABEL.PAGA, tone: "success", icon: ArrowUpFromLineIcon },
  CANCELADA: { label: STATUS_LABEL.CANCELADA, tone: "neutral", icon: BanIcon },
};

export const TIPOS = Object.keys(TIPO_LABEL) as TipoConta[];
export const FORMAS_PAGAMENTO = Object.keys(FORMA_PAGAMENTO_LABEL) as FormaPagamento[];
export const STATUSES = Object.keys(STATUS_LABEL) as StatusLancamento[];
