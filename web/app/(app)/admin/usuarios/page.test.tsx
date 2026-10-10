import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { mockListUsers } from "@/features/identity/usuarios/test-helpers";
import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, user } from "@/test/msw/fixtures";

import UsuariosPage from "./page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => {
  resetNavigation();
  setLocation("/admin/usuarios");
});

/** As demais permissões de identity: nenhuma abre a rota. */
const IDENTITY_WITHOUT_READ = [
  "identity:user:create",
  "identity:user:update",
  "identity:role:assign",
  "identity:admin:grant",
  "identity:admin:revoke",
  "identity:user:reset_password",
];

// USR-01 AC1 e AC4: /admin/usuarios exige identity:user:read.
describe("/admin/usuarios", () => {
  it("com identity:user:read, monta a lista de usuários", async () => {
    mockMe(authContext({ permissions: ["identity:user:read"] }));
    mockListUsers(() => ({ items: [user()], next_cursor: null }));
    renderWithSession(<UsuariosPage />);
    expect(await screen.findByRole("heading", { name: "Usuários" })).toBeTruthy();
    expect(await screen.findByText("outra.pessoa@example.com")).toBeTruthy();
  });

  it("sem identity:user:read, mostra Sem acesso e não consulta a API", async () => {
    mockMe(authContext({ permissions: ["financeiro:conta:read"] }));
    const requests = mockListUsers(() => ({ items: [user()], next_cursor: null }));
    renderWithSession(<UsuariosPage />);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    expect(screen.queryByRole("heading", { name: "Usuários" })).toBeNull();
    expect(requests).toEqual([]);
  });

  // FND-04 AC3/AC4: a rota é liberada só por identity:user:read; as outras
  // permissões de identity (as "vizinhas") não a abrem.
  it("só com identity:user:read, sem nenhuma outra permissão, consulta GET /users", async () => {
    mockMe(authContext({ permissions: ["identity:user:read"] }));
    const requests = mockListUsers(() => ({ items: [user()], next_cursor: null }));
    renderWithSession(<UsuariosPage />);
    expect(await screen.findByText("outra.pessoa@example.com")).toBeTruthy();
    expect(screen.queryByRole("heading", { name: "Sem acesso" })).toBeNull();
    expect(requests).toEqual([{ limit: "50" }]);
  });

  it("com todas as outras permissões de identity, menos identity:user:read: Sem acesso e nenhuma consulta", async () => {
    mockMe(authContext({ permissions: IDENTITY_WITHOUT_READ }));
    const requests = mockListUsers(() => ({ items: [user()], next_cursor: null }));
    renderWithSession(<UsuariosPage />);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    expect(screen.queryByRole("heading", { name: "Usuários" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Novo usuário" })).toBeNull();
    expect(screen.queryByText("outra.pessoa@example.com")).toBeNull();
    expect(requests).toEqual([]);
  });

  it.each(IDENTITY_WITHOUT_READ.map((permission) => ({ permission })))(
    "só com $permission: Sem acesso e nenhuma consulta",
    async ({ permission }) => {
      mockMe(authContext({ permissions: [permission] }));
      const requests = mockListUsers(() => ({ items: [user()], next_cursor: null }));
      renderWithSession(<UsuariosPage />);
      expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
      expect(screen.queryByRole("heading", { name: "Usuários" })).toBeNull();
      expect(requests).toEqual([]);
    },
  );
});
