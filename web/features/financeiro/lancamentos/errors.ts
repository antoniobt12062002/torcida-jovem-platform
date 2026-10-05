import type { ErrorCatalog } from "@/lib/api/problem";

// Mensagens dos erros de lançamentos (design, "Error Handling Strategy").

export const MESSAGE_CONTA_INVALIDA = "Escolha uma conta ativa do mesmo tipo do lançamento.";
export const MESSAGE_TIPO_INCOMPATIVEL = "O tipo do lançamento não combina com o tipo da conta.";
export const MESSAGE_AMOUNT_OUT_OF_RANGE = "Valor fora da faixa permitida.";
export const MESSAGE_IMUTAVEL = "Só lançamentos em aberto podem ser editados.";
export const MESSAGE_NAO_ENCONTRADO = "Lançamento não encontrado.";
export const MESSAGE_NAO_ESTA_EM_ABERTO = "Este lançamento não está mais em aberto.";
export const MESSAGE_JA_CANCELADO = "Este lançamento já foi cancelado.";
export const MESSAGE_MOTIVO_OBRIGATORIO = "Informe o motivo.";
export const MESSAGE_DEVOLUCAO_INVALIDA = "Só é possível devolver uma receita recebida.";

export const lancamentosCatalog: ErrorCatalog = {
  conta_invalida: MESSAGE_CONTA_INVALIDA,
  lancamento_tipo_incompativel: MESSAGE_TIPO_INCOMPATIVEL,
  amount_out_of_range: MESSAGE_AMOUNT_OUT_OF_RANGE,
  lancamento_imutavel: MESSAGE_IMUTAVEL,
  lancamento_nao_encontrado: MESSAGE_NAO_ENCONTRADO,
  lancamento_nao_pode_ser_recebido: MESSAGE_NAO_ESTA_EM_ABERTO,
  lancamento_nao_pode_ser_pago: MESSAGE_NAO_ESTA_EM_ABERTO,
  lancamento_ja_cancelado: MESSAGE_JA_CANCELADO,
  motivo_obrigatorio: MESSAGE_MOTIVO_OBRIGATORIO,
  devolucao_invalida: MESSAGE_DEVOLUCAO_INVALIDA,
};

/**
 * Erros sem `errors[]` que pertencem a um campo do formulário (FWB-03 AC5).
 * A API não diz qual valor saiu da faixa; a mensagem fica no valor bruto.
 */
export const FIELD_BY_CODE = {
  conta_invalida: "conta_id",
  lancamento_tipo_incompativel: "conta_id",
  amount_out_of_range: "valor_bruto_cents",
} as const;

/** Erros que indicam lista desatualizada: a lista é buscada de novo. */
export const STALE_CODES: ReadonlySet<string> = new Set([
  "lancamento_imutavel",
  "lancamento_nao_encontrado",
  "lancamento_nao_pode_ser_recebido",
  "lancamento_nao_pode_ser_pago",
  "lancamento_ja_cancelado",
  "devolucao_invalida",
]);
