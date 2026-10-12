import { PackageIcon } from "lucide-react";
import { expect, it } from "vitest";

import { visibleItems } from "@/components/app/nav";

import { estoqueNav } from "./nav";
import { allExcept, PERM } from "./test-helpers";

// INT-01 preservado (design 05, "Impacto nos requisitos"): mesma lista, mesma
// ordem e mesmas permissões; entram o ícone e a descrição (DSS-01 AC3).
it("seção Estoque com Produtos, exigindo estoque:produto:read", () => {
  expect(estoqueNav).toEqual({
    label: "Estoque",
    items: [
      {
        label: "Produtos",
        href: "/estoque/produtos",
        permission: "estoque:produto:read",
        icon: PackageIcon,
        description: "Produtos, saldos e movimentações do estoque.",
      },
    ],
  });
  expect(visibleItems([estoqueNav], ["estoque:saldo:read"])).toEqual([]);
});

it("o item Produtos aparece só com estoque:produto:read", () => {
  expect(visibleItems([estoqueNav], allExcept(PERM.produtoRead))).toEqual([]);
  expect(visibleItems([estoqueNav], [PERM.produtoRead])).toEqual([estoqueNav]);
});
