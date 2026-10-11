import type { components } from "@/lib/api/estoque";
import type { ErrorCatalog } from "@/lib/api/problem";

// Catálogo de erros do estoque (design 05, "Error Handling Strategy") e
// rótulos em português dos tipos e origens de movimentação.

export const MESSAGE_CODIGO_DUPLICADO = "Já existe um produto com este código.";
export const MESSAGE_SALDO_INSUFICIENTE = "Saldo insuficiente para esta saída.";
export const MESSAGE_PRODUTO_NAO_ENCONTRADO = "Produto não encontrado.";
export const MESSAGE_QUANTIDADE_INVALIDA = "Quantidade inválida.";
export const MESSAGE_DEVOLUCAO_INVALIDA = "Só é possível devolver uma entrada ou saída deste produto.";
export const MESSAGE_MOTIVO_OBRIGATORIO = "Informe o motivo do ajuste.";

export const estoqueCatalog: ErrorCatalog = {
  codigo_duplicado: MESSAGE_CODIGO_DUPLICADO,
  saldo_insuficiente: MESSAGE_SALDO_INSUFICIENTE,
  produto_nao_encontrado: MESSAGE_PRODUTO_NAO_ENCONTRADO,
  quantidade_invalida: MESSAGE_QUANTIDADE_INVALIDA,
  devolucao_invalida: MESSAGE_DEVOLUCAO_INVALIDA,
  motivo_obrigatorio: MESSAGE_MOTIVO_OBRIGATORIO,
};

export type TipoMovimentacao = components["schemas"]["TipoMovimentacao"];
export type OrigemMovimentacao = components["schemas"]["OrigemMovimentacao"];

export const tipoLabels: Readonly<Record<TipoMovimentacao, string>> = {
  ENTRADA: "Entrada",
  SAIDA: "Saída",
  AJUSTE: "Ajuste",
  DEVOLUCAO: "Devolução",
};

export const origemLabels: Readonly<Record<OrigemMovimentacao, string>> = {
  VENDA: "Venda",
  COMPRA: "Compra",
  EVENTO: "Evento",
  INVENTARIO: "Inventário",
  AJUSTE_MANUAL: "Ajuste manual",
};
