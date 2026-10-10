import { describe, expect, it } from "vitest";

import { visibleItems } from "@/components/app/nav";

import { financeiroNav } from "./nav";

// Navegação do módulo (design 04, pontos de integração): os três itens do
// financeiro, cada um com a permissão de leitura da sua área.

describe("financeiroNav", () => {
  it("é a seção Financeiro com Lançamentos, Contas e Saldo e as permissões de leitura", () => {
    expect(financeiroNav).toEqual({
      label: "Financeiro",
      items: [
        { label: "Lançamentos", href: "/financeiro/lancamentos", permission: "financeiro:lancamento:read" },
        { label: "Contas", href: "/financeiro/contas", permission: "financeiro:conta:read" },
        { label: "Saldo", href: "/financeiro/saldo", permission: "financeiro:saldo:read" },
      ],
    });
  });

  it("mostra só os itens cuja permissão de leitura a pessoa tem", () => {
    expect(visibleItems([financeiroNav], ["financeiro:conta:read"])).toEqual([
      {
        label: "Financeiro",
        items: [{ label: "Contas", href: "/financeiro/contas", permission: "financeiro:conta:read" }],
      },
    ]);
    expect(visibleItems([financeiroNav], ["estoque:produto:read"])).toEqual([]);
  });
});
