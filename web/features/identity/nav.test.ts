import { describe, expect, it } from "vitest";

import { visibleItems } from "@/components/app/nav";

import { identityNav, identityNavSection } from "./nav";

// USR-01 AC4: o item "Usuários" leva a /admin/usuarios e só aparece com
// identity:user:read.

describe("navegação de identity", () => {
  it("expõe o item Usuários com a permissão identity:user:read", () => {
    expect(identityNav).toEqual([
      { label: "Usuários", href: "/admin/usuarios", permission: "identity:user:read" },
    ]);
    expect(identityNavSection).toEqual({ label: "Administração", items: identityNav });
  });

  it("some sem identity:user:read e aparece com ela", () => {
    expect(visibleItems([identityNavSection], ["financeiro:conta:read"])).toEqual([]);
    expect(visibleItems([identityNavSection], ["identity:user:read"])).toEqual([identityNavSection]);
  });
});
