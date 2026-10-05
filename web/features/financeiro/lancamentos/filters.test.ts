import { describe, expect, it } from "vitest";

import { conta, lancamento } from "@/test/msw/fixtures";

import { contasParaLancamento, EMPTY_FILTERS, filterLancamentos, localDate } from "./filters";

// FWB-03 AC1 (filtros de tipo, status, conta e período no cliente; do mais
// novo para o mais antigo, WEB-D-016) e premissa "contas oferecidas num
// lançamento" (só ativas do tipo do lançamento).

const RECEITA_CRIADA = lancamento({ id: "l1", tipo: "RECEITA", status: "CRIADA", conta_id: "c1", criado_em: "2026-09-01T15:00:00Z" });
const RECEITA_RECEBIDA = lancamento({ id: "l2", tipo: "RECEITA", status: "RECEBIDA", conta_id: "c1", criado_em: "2026-09-10T15:00:00Z" });
const DESPESA_PAGA = lancamento({ id: "l3", tipo: "DESPESA", status: "PAGA", conta_id: "c2", criado_em: "2026-09-20T15:00:00Z" });
const DESPESA_CANCELADA = lancamento({ id: "l4", tipo: "DESPESA", status: "CANCELADA", conta_id: "c2", criado_em: "2026-10-01T15:00:00Z" });
const ALL = [RECEITA_CRIADA, DESPESA_PAGA, RECEITA_RECEBIDA, DESPESA_CANCELADA];

function ids(items: { id: string }[]) {
  return items.map((l) => l.id);
}

describe("ordenação", () => {
  it("sem filtros, devolve todos do mais novo para o mais antigo", () => {
    expect(ids(filterLancamentos(ALL, EMPTY_FILTERS))).toEqual(["l4", "l3", "l2", "l1"]);
  });

  it("não altera a lista recebida", () => {
    const copy = [...ALL];
    filterLancamentos(ALL, EMPTY_FILTERS);
    expect(ALL).toEqual(copy);
  });
});

describe("filtros", () => {
  it("por tipo", () => {
    expect(ids(filterLancamentos(ALL, { ...EMPTY_FILTERS, tipo: "RECEITA" }))).toEqual(["l2", "l1"]);
    expect(ids(filterLancamentos(ALL, { ...EMPTY_FILTERS, tipo: "DESPESA" }))).toEqual(["l4", "l3"]);
  });

  it.each([
    ["CRIADA", ["l1"]],
    ["RECEBIDA", ["l2"]],
    ["PAGA", ["l3"]],
    ["CANCELADA", ["l4"]],
  ] as const)("por status %s", (status, expected) => {
    expect(ids(filterLancamentos(ALL, { ...EMPTY_FILTERS, status }))).toEqual(expected);
  });

  it("por conta", () => {
    expect(ids(filterLancamentos(ALL, { ...EMPTY_FILTERS, contaId: "c2" }))).toEqual(["l4", "l3"]);
    expect(ids(filterLancamentos(ALL, { ...EMPTY_FILTERS, contaId: "inexistente" }))).toEqual([]);
  });

  it("por período de criação, com as duas pontas inclusivas", () => {
    expect(
      ids(filterLancamentos(ALL, { ...EMPTY_FILTERS, de: "2026-09-10", ate: "2026-09-20" })),
    ).toEqual(["l3", "l2"]);
  });

  it("só com data inicial ou só com data final", () => {
    expect(ids(filterLancamentos(ALL, { ...EMPTY_FILTERS, de: "2026-09-20" }))).toEqual(["l4", "l3"]);
    expect(ids(filterLancamentos(ALL, { ...EMPTY_FILTERS, ate: "2026-09-01" }))).toEqual(["l1"]);
  });

  it("o período usa a data no fuso de São Paulo", () => {
    // 02:30 UTC de 02/10 ainda é 01/10 em São Paulo (UTC−3).
    const madrugada = lancamento({ id: "l5", criado_em: "2026-10-02T02:30:00Z" });
    expect(localDate(madrugada.criado_em)).toBe("2026-10-01");
    expect(ids(filterLancamentos([madrugada], { ...EMPTY_FILTERS, ate: "2026-10-01" }))).toEqual(["l5"]);
    expect(ids(filterLancamentos([madrugada], { ...EMPTY_FILTERS, de: "2026-10-02" }))).toEqual([]);
  });

  it("combina todos os filtros", () => {
    expect(
      ids(
        filterLancamentos(ALL, {
          tipo: "DESPESA",
          status: "PAGA",
          contaId: "c2",
          de: "2026-09-01",
          ate: "2026-09-30",
        }),
      ),
    ).toEqual(["l3"]);
    expect(
      ids(filterLancamentos(ALL, { ...EMPTY_FILTERS, tipo: "RECEITA", status: "PAGA" })),
    ).toEqual([]);
  });
});

describe("contas oferecidas num lançamento", () => {
  const contas = [
    conta({ id: "r-b", tipo: "RECEITA", nome: "Patrocínios" }),
    conta({ id: "r-a", tipo: "RECEITA", nome: "Mensalidades" }),
    conta({ id: "r-inativa", tipo: "RECEITA", nome: "Rifas", ativo: false }),
    conta({ id: "d-a", tipo: "DESPESA", nome: "Transporte" }),
    conta({ id: "d-inativa", tipo: "DESPESA", nome: "Bandeiras", ativo: false }),
  ];

  it("só contas ativas do tipo de receita, por nome", () => {
    expect(ids(contasParaLancamento(contas, "RECEITA"))).toEqual(["r-a", "r-b"]);
  });

  it("só contas ativas do tipo de despesa", () => {
    expect(ids(contasParaLancamento(contas, "DESPESA"))).toEqual(["d-a"]);
  });

  it("sem contas carregadas, nenhuma opção", () => {
    expect(contasParaLancamento(undefined, "RECEITA")).toEqual([]);
  });
});
