import { House } from "lucide-react";
import { describe, expect, it } from "vitest";

import { estoqueNav } from "@/features/estoque/nav";
import { financeiroNav } from "@/features/financeiro/nav";
import { identityNavSection } from "@/features/identity/nav";

import { isNavSection, type NavEntry } from "./nav";
import { navItems } from "./nav-items";

// INT-01 AC1: o registro global junta os itens dos módulos, mais "Início", na
// ordem da spec (Início, Financeiro, Estoque, Administração), sem redefinir
// caminho nem permissão de nenhum item.

describe("navItems", () => {
  // DSS-01 AC3 e DSS-06 AC2: "Início" ganha ícone e descrição (design: House).
  it("começa por Início, em /inicio, sem permissão exigida, com ícone e descrição", () => {
    expect(navItems[0]).toEqual({
      label: "Início",
      href: "/inicio",
      icon: House,
      description: "Atalhos para as áreas que a sua conta pode usar.",
    });
  });

  it("segue com Financeiro, Estoque e Administração, nessa ordem", () => {
    expect(navItems.map((entry) => entry.label)).toEqual(["Início", "Financeiro", "Estoque", "Administração"]);
  });

  it("usa as próprias seções exportadas pelos módulos, sem cópia nem redefinição", () => {
    expect(navItems[1]).toBe(financeiroNav);
    expect(navItems[2]).toBe(estoqueNav);
    expect(navItems[3]).toBe(identityNavSection);
  });

  // Os módulos acrescentam ícone e descrição aos próprios itens (DSS-01 AC3);
  // a comparação projeta cada item em rótulo, caminho e permissão.
  it("mantém os caminhos e as permissões definidos pelos módulos", () => {
    expect(navItems.slice(1).map(project)).toEqual([
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

function project(entry: NavEntry) {
  if (!isNavSection(entry)) return { label: entry.label, href: entry.href, permission: entry.permission };
  return {
    label: entry.label,
    items: entry.items.map(({ label, href, permission }) => ({ label, href, permission })),
  };
}
