import { ListTreeIcon, ReceiptIcon, WalletIcon } from "lucide-react";
import { describe, expect, it } from "vitest";

import { visibleItems } from "@/components/app/nav";

import { financeiroNav } from "./nav";

// Navegação do módulo (design 04, pontos de integração): os três itens do
// financeiro, cada um com a permissão de leitura da sua área, o ícone do
// shell e a frase do atalho do início (DSS-01 AC3).

describe("financeiroNav", () => {
  it("é a seção Financeiro com Lançamentos, Contas e Saldo e as permissões de leitura", () => {
    expect(financeiroNav).toEqual({
      label: "Financeiro",
      items: [
        { label: "Lançamentos", href: "/financeiro/lancamentos", permission: "financeiro:lancamento:read", icon: ReceiptIcon, description: "Receitas e despesas da torcida, com os comprovantes." },
        { label: "Contas", href: "/financeiro/contas", permission: "financeiro:conta:read", icon: ListTreeIcon, description: "Plano de contas de receita e de despesa." },
        { label: "Saldo", href: "/financeiro/saldo", permission: "financeiro:saldo:read", icon: WalletIcon, description: "Saldo atual: recebidos menos pagos." },
      ],
    });
  });

  it("mostra só os itens cuja permissão de leitura a pessoa tem", () => {
    expect(visibleItems([financeiroNav], ["financeiro:conta:read"])).toEqual([
      {
        label: "Financeiro",
        items: [{ label: "Contas", href: "/financeiro/contas", permission: "financeiro:conta:read", icon: ListTreeIcon, description: "Plano de contas de receita e de despesa." }],
      },
    ]);
    expect(visibleItems([financeiroNav], ["estoque:produto:read"])).toEqual([]);
  });
});
