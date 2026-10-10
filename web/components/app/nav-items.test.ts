import { describe, expect, it } from "vitest";

import { estoqueNav } from "@/features/estoque/nav";
import { financeiroNav } from "@/features/financeiro/nav";
import { identityNavSection } from "@/features/identity/nav";

import { navItems } from "./nav-items";

// INT-01 AC1: o registro global junta os itens dos módulos, mais "Início", na
// ordem da spec (Início, Financeiro, Estoque, Administração), sem redefinir
// caminho nem permissão de nenhum item.

describe("navItems", () => {
  it("começa por Início, em /inicio, sem permissão exigida", () => {
    expect(navItems[0]).toEqual({ label: "Início", href: "/inicio" });
  });

  it("segue com Financeiro, Estoque e Administração, nessa ordem", () => {
    expect(navItems.map((entry) => entry.label)).toEqual(["Início", "Financeiro", "Estoque", "Administração"]);
  });

  it("usa as próprias seções exportadas pelos módulos, sem cópia nem redefinição", () => {
    expect(navItems[1]).toBe(financeiroNav);
    expect(navItems[2]).toBe(estoqueNav);
    expect(navItems[3]).toBe(identityNavSection);
  });

  it("mantém os caminhos e as permissões definidos pelos módulos", () => {
    expect(navItems.slice(1)).toEqual([
      {
        label: "Financeiro",
        items: [
          { label: "Lançamentos", href: "/financeiro/lancamentos", permission: "financeiro:lancamento:read" },
          { label: "Contas", href: "/financeiro/contas", permission: "financeiro:conta:read" },
          { label: "Saldo", href: "/financeiro/saldo", permission: "financeiro:saldo:read" },
        ],
      },
      {
        label: "Estoque",
        items: [{ label: "Produtos", href: "/estoque/produtos", permission: "estoque:produto:read" }],
      },
      {
        label: "Administração",
        items: [{ label: "Usuários", href: "/admin/usuarios", permission: "identity:user:read" }],
      },
    ]);
  });
});
