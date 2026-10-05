import { describe, expect, it } from "vitest";

import { movimentacao } from "@/test/msw/fixtures";

import { movimentacaoPayload, podeDevolver } from "./payload";

// EWB-02 AC2-4 e WEB-D-007: o corpo das movimentações manuais sempre leva
// origem INVENTARIO; nenhuma origem reservada é montada aqui.

const PRODUTO = "00000000-0000-4000-8000-000000000401";
const REF = "00000000-0000-4000-8000-000000000501";

describe("movimentacaoPayload", () => {
  it("entrada: tipo, produto, quantidade e origem INVENTARIO, nada mais", () => {
    expect(movimentacaoPayload({ tipo: "ENTRADA", produtoId: PRODUTO, quantidade: 10 })).toStrictEqual({
      tipo: "ENTRADA",
      produto_id: PRODUTO,
      quantidade: 10,
      origem: "INVENTARIO",
    });
  });

  it("saída: tipo, produto, quantidade e origem INVENTARIO, nada mais", () => {
    expect(movimentacaoPayload({ tipo: "SAIDA", produtoId: PRODUTO, quantidade: 4 })).toStrictEqual({
      tipo: "SAIDA",
      produto_id: PRODUTO,
      quantidade: 4,
      origem: "INVENTARIO",
    });
  });

  it("devolução: leva movimentacao_de_id e origem INVENTARIO", () => {
    expect(
      movimentacaoPayload({ tipo: "DEVOLUCAO", produtoId: PRODUTO, quantidade: 4, movimentacaoDeId: REF }),
    ).toStrictEqual({
      tipo: "DEVOLUCAO",
      produto_id: PRODUTO,
      quantidade: 4,
      origem: "INVENTARIO",
      movimentacao_de_id: REF,
    });
  });

  it("ignora qualquer origem que venha junto na entrada", () => {
    const input = { tipo: "ENTRADA", produtoId: PRODUTO, quantidade: 1, origem: "VENDA" } as Parameters<
      typeof movimentacaoPayload
    >[0];
    expect(movimentacaoPayload(input).origem).toBe("INVENTARIO");
  });
});

// Caso de borda: "Devolver" só em ENTRADA e SAIDA.
describe("podeDevolver", () => {
  it.each([
    ["ENTRADA", true],
    ["SAIDA", true],
    ["AJUSTE", false],
    ["DEVOLUCAO", false],
  ] as const)("%s → %s", (tipo, expected) => {
    expect(podeDevolver(movimentacao({ tipo }))).toBe(expected);
  });
});
