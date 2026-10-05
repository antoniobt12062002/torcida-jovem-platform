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
});
