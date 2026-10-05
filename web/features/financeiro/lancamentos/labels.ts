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

export const TIPOS = Object.keys(TIPO_LABEL) as TipoConta[];
export const FORMAS_PAGAMENTO = Object.keys(FORMA_PAGAMENTO_LABEL) as FormaPagamento[];
export const STATUSES = Object.keys(STATUS_LABEL) as StatusLancamento[];
