import { describe, expect, it } from "vitest";

import { conta } from "@/test/msw/fixtures";

import { buildContaTree } from "./tree";

// FWB-01 AC1: árvore por parent_id. Plano de contas (assunção da spec): contas
// inativas continuam na árvore. Contas órfãs (pai fora da lista) vão para a raiz.

const raiz = conta({ id: "r1", nome: "Receitas", parent_id: null });
const sub = conta({ id: "s1", nome: "Mensalidades", parent_id: "r1" });
const neta = conta({ id: "n1", nome: "Mensalidade sócio", parent_id: "s1" });
const inativa = conta({ id: "s2", nome: "Rifas", parent_id: "r1", ativo: false });
const despesa = conta({ id: "r2", nome: "Despesas", tipo: "DESPESA", parent_id: null });
const orfa = conta({ id: "o1", nome: "Órfã", parent_id: "nao-existe" });

function shape(nodes: ReturnType<typeof buildContaTree>): unknown[] {
  return nodes.map((n) => ({ id: n.conta.id, children: shape(n.children) }));
}

describe("buildContaTree", () => {
  it("lista vazia gera árvore vazia", () => {
    expect(buildContaTree([])).toEqual([]);
  });

  it("aninha subcontas sob o pai, em qualquer profundidade, mantendo a ordem da lista", () => {
    const tree = buildContaTree([neta, sub, raiz, inativa, despesa]);
    expect(shape(tree)).toEqual([
      {
        id: "r1",
        children: [
          { id: "s1", children: [{ id: "n1", children: [] }] },
          { id: "s2", children: [] },
        ],
      },
      { id: "r2", children: [] },
    ]);
  });

  it("mantém contas inativas na árvore, com o próprio dado", () => {
    const tree = buildContaTree([raiz, inativa]);
    expect(tree[0].children[0].conta).toEqual(inativa);
  });

  it("põe na raiz a conta cujo pai não está na lista", () => {
    expect(shape(buildContaTree([raiz, orfa]))).toEqual([
      { id: "r1", children: [] },
      { id: "o1", children: [] },
    ]);
  });

  it("não altera a lista recebida", () => {
    const list = [sub, raiz];
    const copy = structuredClone(list);
    buildContaTree(list);
    expect(list).toEqual(copy);
  });
});
