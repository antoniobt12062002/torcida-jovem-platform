import { describe, expect, it } from "vitest";

import { lancamentoActions } from "./actions";
import { STATUSES, TIPOS } from "./labels";

// FWB-03 AC4: editar só em lançamentos CRIADA e só com
// financeiro:lancamento:update; casos de borda: sessão só de leitura não
// recebe ação de escrita.

const READ = "financeiro:lancamento:read";
const UPDATE = "financeiro:lancamento:update";

describe("editar", () => {
  for (const tipo of TIPOS) {
    for (const status of STATUSES) {
      it(`${tipo} ${status} com permissão de editar`, () => {
        const actions = lancamentoActions({ tipo, status }, [READ, UPDATE]);
        expect(actions.includes("editar")).toBe(status === "CRIADA");
      });

      it(`${tipo} ${status} só com leitura não oferece ação`, () => {
        expect(lancamentoActions({ tipo, status }, [READ])).toEqual([]);
      });
    }
  }

  it("sem nenhuma permissão, nenhuma ação", () => {
    expect(lancamentoActions({ tipo: "RECEITA", status: "CRIADA" }, [])).toEqual([]);
  });
});
