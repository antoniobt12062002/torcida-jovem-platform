import type { components } from "@/lib/api/estoque";

// Corpo das movimentações manuais (EWB-02 AC2-4, WEB-D-007). A origem é
// sempre INVENTARIO: as origens VENDA, COMPRA e EVENTO são reservadas para
// integrações e AJUSTE_MANUAL só existe no fluxo de ajuste.

export type CreateMovimentacaoRequest = components["schemas"]["CreateMovimentacaoRequest"];
type Movimentacao = components["schemas"]["Movimentacao"];

export const ORIGEM_MANUAL = "INVENTARIO" satisfies CreateMovimentacaoRequest["origem"];

export type MovimentacaoManual =
  | { tipo: "ENTRADA" | "SAIDA"; produtoId: string; quantidade: number }
  | { tipo: "DEVOLUCAO"; produtoId: string; quantidade: number; movimentacaoDeId: string };

export function movimentacaoPayload(input: MovimentacaoManual): CreateMovimentacaoRequest {
  const body: CreateMovimentacaoRequest = {
    tipo: input.tipo,
    produto_id: input.produtoId,
    quantidade: input.quantidade,
    origem: ORIGEM_MANUAL,
  };
  if (input.tipo === "DEVOLUCAO") body.movimentacao_de_id = input.movimentacaoDeId;
  return body;
}

/** Só entradas e saídas podem ser devolvidas (a API recusa as demais). */
export function podeDevolver(mov: Pick<Movimentacao, "tipo">): boolean {
  return mov.tipo === "ENTRADA" || mov.tipo === "SAIDA";
}
