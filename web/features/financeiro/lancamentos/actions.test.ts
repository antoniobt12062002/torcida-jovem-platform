import { describe, expect, it } from "vitest";

import { type LancamentoAction, lancamentoActions } from "./actions";
import { type StatusLancamento, STATUSES, type TipoConta, TIPOS } from "./labels";

// FWB-03 AC4 (editar só em CRIADA), FWB-04 AC6 (devolver só em receita
// RECEBIDA), FWB-04 AC7 (receber só em receita CRIADA, pagar só em despesa
// CRIADA, cancelar só em não cancelados) e casos de borda (só leitura não
// recebe ação de escrita). Todas as combinações de tipo, status e permissão.

const READ = "financeiro:lancamento:read";

const PERMISSION: Record<LancamentoAction, string> = {
  editar: "financeiro:lancamento:update",
  receber: "financeiro:lancamento:receive",
  pagar: "financeiro:lancamento:pay",
  cancelar: "financeiro:lancamento:cancel",
  devolver: "financeiro:lancamento:create",
};

const ALL_PERMISSIONS = [READ, ...Object.values(PERMISSION)];

/** Regra da spec, escrita como tabela: em que tipo e status cada ação existe. */
const EXPECTED: Record<LancamentoAction, (tipo: TipoConta, status: StatusLancamento) => boolean> = {
  editar: (_, status) => status === "CRIADA",
  receber: (tipo, status) => tipo === "RECEITA" && status === "CRIADA",
  pagar: (tipo, status) => tipo === "DESPESA" && status === "CRIADA",
  cancelar: (_, status) => status !== "CANCELADA",
  devolver: (tipo, status) => tipo === "RECEITA" && status === "RECEBIDA",
};

const ACTIONS = Object.keys(PERMISSION) as LancamentoAction[];

describe("todas as combinações de tipo × status × permissão", () => {
  for (const tipo of TIPOS) {
    for (const status of STATUSES) {
      for (const action of ACTIONS) {
        const expected = EXPECTED[action](tipo, status);

        it(`${tipo} ${status}: ${action} com a permissão → ${expected}`, () => {
          expect(lancamentoActions({ tipo, status }, [READ, PERMISSION[action]]).includes(action)).toBe(expected);
          expect(lancamentoActions({ tipo, status }, ALL_PERMISSIONS).includes(action)).toBe(expected);
        });

        it(`${tipo} ${status}: ${action} sem a permissão → false`, () => {
          const others = ALL_PERMISSIONS.filter((p) => p !== PERMISSION[action]);
          expect(lancamentoActions({ tipo, status }, others).includes(action)).toBe(false);
        });
      }

      it(`${tipo} ${status}: só leitura não oferece ação`, () => {
        expect(lancamentoActions({ tipo, status }, [READ])).toEqual([]);
      });
    }
  }
});

describe("exemplos completos", () => {
  it("receita CRIADA da tesouraria", () => {
    expect(lancamentoActions({ tipo: "RECEITA", status: "CRIADA" }, ALL_PERMISSIONS)).toEqual([
      "receber",
      "editar",
      "cancelar",
    ]);
  });

  it("despesa CRIADA da tesouraria", () => {
    expect(lancamentoActions({ tipo: "DESPESA", status: "CRIADA" }, ALL_PERMISSIONS)).toEqual([
      "pagar",
      "editar",
      "cancelar",
    ]);
  });

  it("receita RECEBIDA da tesouraria", () => {
    expect(lancamentoActions({ tipo: "RECEITA", status: "RECEBIDA" }, ALL_PERMISSIONS)).toEqual([
      "devolver",
      "cancelar",
    ]);
  });

  it("despesa PAGA da tesouraria", () => {
    expect(lancamentoActions({ tipo: "DESPESA", status: "PAGA" }, ALL_PERMISSIONS)).toEqual(["cancelar"]);
  });

  it("cancelado não oferece nada", () => {
    for (const tipo of TIPOS) {
      expect(lancamentoActions({ tipo, status: "CANCELADA" }, ALL_PERMISSIONS)).toEqual([]);
    }
  });

  it("sem nenhuma permissão, nenhuma ação", () => {
    expect(lancamentoActions({ tipo: "RECEITA", status: "CRIADA" }, [])).toEqual([]);
  });
});
