import { describe, expect, it } from "vitest";

import { produto } from "@/test/msw/fixtures";

import { searchProdutos } from "./search";

// EWB-01 AC1: busca por código ou nome, no cliente (WEB-D-016).

const camisa = produto({ id: "p1", codigo: "CAM-001", nome: "Camisa oficial" });
const bone = produto({ id: "p2", codigo: "BON-010", nome: "Boné azul" });
const caneca = produto({ id: "p3", codigo: "CAN-002", nome: "Caneca da torcida" });
const all = [camisa, bone, caneca];

describe("searchProdutos", () => {
  it("termo vazio ou só espaços devolve a lista inteira, na mesma ordem", () => {
    expect(searchProdutos(all, "")).toEqual(all);
    expect(searchProdutos(all, "   ")).toEqual(all);
  });

  it("encontra pelo código, sem diferenciar maiúsculas", () => {
    expect(searchProdutos(all, "bon-010")).toEqual([bone]);
  });

  it("encontra por parte do nome", () => {
    expect(searchProdutos(all, "torcida")).toEqual([caneca]);
  });

  it("ignora acentos e espaços nas pontas", () => {
    expect(searchProdutos(all, "  bone ")).toEqual([bone]);
    expect(searchProdutos(all, "CANECA")).toEqual([caneca]);
  });

  it("termo que casa código de um e nome de outro devolve os dois", () => {
    expect(searchProdutos(all, "ca")).toEqual([camisa, caneca]);
  });

  it("sem correspondência devolve lista vazia", () => {
    expect(searchProdutos(all, "inexistente")).toEqual([]);
  });

  it("não procura em outros campos (unidade de medida)", () => {
    expect(searchProdutos(all, "UN")).toEqual([]);
  });
});
