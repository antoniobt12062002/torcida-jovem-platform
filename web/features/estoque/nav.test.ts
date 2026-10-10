import { expect, it } from "vitest";

import { visibleItems } from "@/components/app/nav";

import { estoqueNav } from "./nav";
import { allExcept, PERM } from "./test-helpers";

it("seção Estoque com Produtos, exigindo estoque:produto:read", () => {
  expect(estoqueNav).toEqual({
    label: "Estoque",
    items: [{ label: "Produtos", href: "/estoque/produtos", permission: "estoque:produto:read" }],
  });
  expect(visibleItems([estoqueNav], ["estoque:saldo:read"])).toEqual([]);
});

it("o item Produtos aparece só com estoque:produto:read", () => {
  expect(visibleItems([estoqueNav], allExcept(PERM.produtoRead))).toEqual([]);
  expect(visibleItems([estoqueNav], [PERM.produtoRead])).toEqual([estoqueNav]);
});
