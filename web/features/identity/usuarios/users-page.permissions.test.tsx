import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, ids, user } from "@/test/msw/fixtures";

import { mockListUsers, openRowMenu } from "./test-helpers";
import { UsersPage } from "./users-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

// FND-04 AC3 na administração de usuários: cada ação de escrita aparece pela
// permissão nomeada na spec, e só por ela. Para cada ação há um caso positivo
// isolado (leitura + só a permissão da ação) e um negativo cruzado (leitura +
// todas as outras escritas do módulo, menos a da ação).
// USR-02 AC4, USR-03 AC1/AC2, USR-04 AC1/AC3/AC4/AC6, USR-05 AC1.

const READ = "identity:user:read";

const WRITES = [
  "identity:user:create",
  "identity:user:update",
  "identity:admin:grant",
  "identity:role:assign",
  "identity:admin:revoke",
  "identity:user:reset_password",
] as const;

type Write = (typeof WRITES)[number];

const MEMBERSHIP = { reason: "Assumiu a tesouraria", granted_at: "2026-10-01T12:00:00Z" };

const eu = user({
  id: ids.user,
  name: "Pessoa Teste",
  email: "pessoa.teste@example.com",
  roles: ["ASSOCIADO", "PRESIDENTE"],
  admin_membership: MEMBERSHIP,
});
const gabi = user({
  id: "00000000-0000-4000-8000-0000000000a6",
  name: "Gabi Associada",
  email: "gabi@example.com",
});
const hana = user({
  id: "00000000-0000-4000-8000-0000000000a7",
  name: "Hana Tesoureira",
  email: "hana@example.com",
  roles: ["ASSOCIADO", "TESOURARIA"],
  admin_membership: MEMBERSHIP,
});
const ivo = user({
  id: "00000000-0000-4000-8000-0000000000a8",
  name: "Ivo Inativo",
  email: "ivo@example.com",
  active: false,
});

const NEW_USER = "Novo usuário";
const GRANT = "Promover a administrativo";
const SET_ROLES = "Editar papéis";
const REVOKE = "Retirar acesso administrativo";
const RESET = "Gerar senha temporária";
const DEACTIVATE = "Desativar";
const REACTIVATE = "Reativar";

/** O que a tela oferece: o botão do cabeçalho e o menu de cada linha. */
type Offered = { novo: boolean; eu: string[]; gabi: string[]; hana: string[]; ivo: string[] };

/** Com as seis permissões de escrita (a própria linha nunca tem ações). */
const EVERYTHING: Offered = {
  novo: true,
  eu: [],
  gabi: [GRANT, RESET, DEACTIVATE],
  hana: [SET_ROLES, REVOKE, RESET, DEACTIVATE],
  ivo: [GRANT, RESET, REACTIVATE],
};

const NOTHING: Offered = { novo: false, eu: [], gabi: [], hana: [], ivo: [] };

/** O que cada permissão libera sozinha, por extenso. */
const ALONE: Readonly<Record<Write, Offered>> = {
  "identity:user:create": { ...NOTHING, novo: true },
  "identity:user:update": { ...NOTHING, gabi: [DEACTIVATE], hana: [DEACTIVATE], ivo: [REACTIVATE] },
  "identity:admin:grant": { ...NOTHING, gabi: [GRANT], ivo: [GRANT] },
  "identity:role:assign": { ...NOTHING, hana: [SET_ROLES] },
  "identity:admin:revoke": { ...NOTHING, hana: [REVOKE] },
  "identity:user:reset_password": { ...NOTHING, gabi: [RESET], hana: [RESET], ivo: [RESET] },
};

/** Tudo, menos o que a permissão libera. */
function without(permission: Write): Offered {
  const removed = ALONE[permission];
  const rest = (all: string[], gone: string[]) => all.filter((label) => !gone.includes(label));
  return {
    novo: EVERYTHING.novo && !removed.novo,
    eu: [],
    gabi: rest(EVERYTHING.gabi, removed.gabi),
    hana: rest(EVERYTHING.hana, removed.hana),
    ivo: rest(EVERYTHING.ivo, removed.ivo),
  };
}

async function renderPage(permissions: readonly string[]) {
  const requests = mockListUsers(() => ({ items: [eu, gabi, hana, ivo], next_cursor: null }));
  mockMe(authContext({ permissions: [...permissions] }));
  renderWithSession(<UsersPage />);
  await screen.findByText("ivo@example.com");
  return requests;
}

/** Rótulos do menu da linha; sem botão de ações, lista vazia. */
async function menuOf(name: string): Promise<string[]> {
  if (!screen.queryByRole("button", { name: `Ações de ${name}` })) return [];
  const labels = await openRowMenu(name);
  fireEvent.keyDown(screen.getByRole("menu"), { key: "Escape" });
  await waitFor(() => expect(screen.queryByRole("menu")).toBeNull());
  return labels;
}

async function offered(): Promise<Offered> {
  return {
    novo: screen.queryByRole("button", { name: NEW_USER }) !== null,
    eu: await menuOf("Pessoa Teste"),
    gabi: await menuOf("Gabi Associada"),
    hana: await menuOf("Hana Tesoureira"),
    ivo: await menuOf("Ivo Inativo"),
  };
}

describe("permissão de cada ação, isolada", () => {
  it.each(WRITES.map((permission) => ({ permission })))(
    "leitura + só $permission: aparece só o que ela libera",
    async ({ permission }) => {
      await renderPage([READ, permission]);
      expect(await offered()).toEqual(ALONE[permission]);
    },
  );

  it.each(WRITES.map((permission) => ({ permission })))(
    "leitura + todas as escritas menos $permission: o que ela libera some, o resto fica",
    async ({ permission }) => {
      await renderPage([READ, ...WRITES.filter((other) => other !== permission)]);
      expect(await offered()).toEqual(without(permission));
    },
  );

  it("só leitura: nenhuma ação de escrita (critério de sucesso da spec)", async () => {
    await renderPage([READ]);
    expect(await offered()).toEqual(NOTHING);
    expect(screen.queryByRole("button", { name: /^Ações de / })).toBeNull();
  });

  it("todas as escritas: cada linha recebe as ações do seu estado, e a própria linha nenhuma", async () => {
    await renderPage([READ, ...WRITES]);
    expect(await offered()).toEqual(EVERYTHING);
  });
});

// USR-04 AC3 e AC4 (G2): as duas permissões da linha com vínculo não se
// confundem entre si.
describe("usuário com vínculo: Editar papéis e Retirar acesso", () => {
  it("identity:role:assign libera só Editar papéis", async () => {
    await renderPage([READ, "identity:role:assign"]);
    expect(await openRowMenu("Hana Tesoureira")).toEqual([SET_ROLES]);
    expect(screen.queryByRole("menuitem", { name: REVOKE })).toBeNull();
  });

  it("identity:admin:revoke libera só Retirar acesso administrativo", async () => {
    await renderPage([READ, "identity:admin:revoke"]);
    expect(await openRowMenu("Hana Tesoureira")).toEqual([REVOKE]);
    expect(screen.queryByRole("menuitem", { name: SET_ROLES })).toBeNull();
  });

  it("escolher Editar papéis abre o diálogo de papéis da pessoa", async () => {
    await renderPage([READ, "identity:role:assign"]);
    await openRowMenu("Hana Tesoureira");
    fireEvent.click(screen.getByRole("menuitem", { name: SET_ROLES }));
    expect(
      await screen.findByRole("dialog", { name: "Editar papéis de Hana Tesoureira" }),
    ).toBeTruthy();
  });
});
