import { describe, expect, it } from "vitest";

import { visibleItems } from "@/components/app/nav";

import { identityNav, identityNavSection } from "./nav";

// USR-01 AC4: o item "Usuários" leva a /admin/usuarios e só aparece com
// identity:user:read.

const IDENTITY_WITHOUT_READ = [
  "identity:user:create",
  "identity:user:update",
  "identity:role:assign",
  "identity:admin:grant",
  "identity:admin:revoke",
  "identity:user:reset_password",
];

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

  // FND-04 AC3: o item depende só de identity:user:read, não das permissões
  // vizinhas do módulo.
  it("some com todas as outras permissões de identity, menos identity:user:read", () => {
    expect(visibleItems([identityNavSection], IDENTITY_WITHOUT_READ)).toEqual([]);
    expect(visibleItems(identityNav, IDENTITY_WITHOUT_READ)).toEqual([]);
  });

  it.each(IDENTITY_WITHOUT_READ.map((permission) => ({ permission })))(
    "some só com $permission",
    ({ permission }) => {
      expect(visibleItems([identityNavSection], [permission])).toEqual([]);
    },
  );

  it("aparece só com identity:user:read, junto ou não das outras", () => {
    expect(visibleItems(identityNav, ["identity:user:read"])).toEqual(identityNav);
    expect(
      visibleItems([identityNavSection], [...IDENTITY_WITHOUT_READ, "identity:user:read"]),
    ).toEqual([identityNavSection]);
  });
});
