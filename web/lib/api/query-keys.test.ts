import { describe, expect, it } from "vitest";

import { queryKeys } from "./query-keys";

const A = "00000000-0000-4000-8000-0000000000aa";
const B = "00000000-0000-4000-8000-0000000000bb";

function startsWith(key: readonly unknown[], prefix: readonly unknown[]) {
  return prefix.every((part, i) => JSON.stringify(key[i]) === JSON.stringify(part));
}

// Design 01-fundacao (query-keys.ts) e WEB-D-010: chaves de todos os recursos
// do contrato, num só lugar, para frentes paralelas compartilharem cache.
describe("queryKeys", () => {
  it("cobre todos os recursos do design, sem chaves repetidas", () => {
    const keys = [
      queryKeys.session.me,
      queryKeys.identity.users(),
      queryKeys.financeiro.contas,
      queryKeys.financeiro.lancamentos,
      queryKeys.financeiro.saldo,
      queryKeys.financeiro.comprovantes(A),
      queryKeys.estoque.produtos,
      queryKeys.estoque.movimentacoes(A),
      queryKeys.estoque.saldo(A),
    ];
    const serialized = keys.map((k) => JSON.stringify(k));
    expect(new Set(serialized).size).toBe(keys.length);
  });

  it("separa recursos por id", () => {
    expect(queryKeys.financeiro.comprovantes(A)).not.toEqual(queryKeys.financeiro.comprovantes(B));
    expect(queryKeys.estoque.movimentacoes(A)).not.toEqual(queryKeys.estoque.movimentacoes(B));
    expect(queryKeys.estoque.saldo(A)).not.toEqual(queryKeys.estoque.saldo(B));
  });

  it("separa listas de usuários por filtro e todas compartilham o prefixo de invalidação", () => {
    const ativos = queryKeys.identity.users({ active: true });
    const tesouraria = queryKeys.identity.users({ role: "TESOURARIA" });
    expect(ativos).not.toEqual(tesouraria);
    expect(startsWith(ativos, queryKeys.identity.usersAll)).toBe(true);
    expect(startsWith(tesouraria, queryKeys.identity.usersAll)).toBe(true);
  });

  it("agrupa as chaves de cada módulo sob o prefixo do módulo", () => {
    for (const key of [
      queryKeys.financeiro.contas,
      queryKeys.financeiro.lancamentos,
      queryKeys.financeiro.saldo,
      queryKeys.financeiro.comprovantes(A),
    ]) {
      expect(startsWith(key, queryKeys.financeiro.all)).toBe(true);
    }
    for (const key of [
      queryKeys.estoque.produtos,
      queryKeys.estoque.movimentacoes(A),
      queryKeys.estoque.saldo(A),
    ]) {
      expect(startsWith(key, queryKeys.estoque.all)).toBe(true);
    }
    expect(startsWith(queryKeys.financeiro.saldo, queryKeys.estoque.all)).toBe(false);
  });
});
