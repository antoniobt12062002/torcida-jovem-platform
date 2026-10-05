import { describe, expect, it } from "vitest";

import { type NavEntry, visibleItems } from "./nav";

// FND-05 AC1 e FND-04 AC3: a navegação mostra só os itens cuja permissão está
// no contexto de sessão; um item sem permissão exigida vale para qualquer
// pessoa autenticada.

const items: NavEntry[] = [
  { label: "Início", href: "/inicio" },
  { label: "Usuários", href: "/usuarios", permission: "identity.users.read" },
  {
    label: "Financeiro",
    items: [
      { label: "Contas", href: "/financeiro/contas", permission: "financeiro.contas.read" },
      { label: "Lançamentos", href: "/financeiro/lancamentos", permission: "financeiro.lancamentos.read" },
    ],
  },
  {
    label: "Estoque",
    items: [{ label: "Produtos", href: "/estoque/produtos", permission: "estoque.produtos.read" }],
  },
];

describe("visibleItems", () => {
  it("sem permissões, mantém só os itens sem permissão exigida e remove seções vazias", () => {
    expect(visibleItems(items, [])).toEqual([{ label: "Início", href: "/inicio" }]);
  });

  it("mantém os itens cuja permissão está no contexto, inclusive dentro de seções", () => {
    expect(visibleItems(items, ["financeiro.contas.read", "identity.users.read"])).toEqual([
      { label: "Início", href: "/inicio" },
      { label: "Usuários", href: "/usuarios", permission: "identity.users.read" },
      {
        label: "Financeiro",
        items: [{ label: "Contas", href: "/financeiro/contas", permission: "financeiro.contas.read" }],
      },
    ]);
  });

  it("não decide pelo nome do papel", () => {
    expect(visibleItems(items, ["ADMIN", "TESOUREIRO"])).toEqual([{ label: "Início", href: "/inicio" }]);
  });
});
