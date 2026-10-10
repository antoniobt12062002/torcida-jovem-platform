import { useQuery } from "@tanstack/react-query";
import { screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import type { ReactElement } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { useSession } from "@/lib/session/session-provider";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { RequirePermission } from "./require-permission";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const AREA = "/api/v1/financeiro/contas";

function SessionReady() {
  return useSession().status === "authenticated" ? <span>sessão pronta</span> : null;
}

/** Tela de exemplo que consulta a API da área ao montar. */
function AreaScreen() {
  const query = useQuery({
    queryKey: ["area-de-teste"],
    queryFn: async () => (await fetch(apiUrl(AREA))).json(),
  });
  return <p>{query.data ? "dados da área" : "tela da área"}</p>;
}

function mockArea() {
  const calls = { count: 0 };
  server.use(
    http.get(apiUrl(AREA), () => {
      calls.count += 1;
      return HttpResponse.json({ items: [] });
    }),
  );
  return calls;
}

async function renderWith(permissions: string[], ui: ReactElement) {
  mockMe(authContext({ permissions }));
  renderWithSession(
    <>
      <SessionReady />
      {ui}
    </>,
  );
  await screen.findByText("sessão pronta");
}

// FND-04 AC4.
describe("modo página", () => {
  it("sem a permissão, mostra Sem acesso e não chama a API da área", async () => {
    const calls = mockArea();
    await renderWith(
      ["estoque.produtos.read"],
      <RequirePermission permission="financeiro.contas.read" mode="page">
        <AreaScreen />
      </RequirePermission>,
    );
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    expect(screen.queryByText(/tela da área|dados da área/)).toBeNull();
    // Dá tempo a uma consulta que não deveria existir.
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(calls.count).toBe(0);
  });

  it("com a permissão, mostra a área", async () => {
    const calls = mockArea();
    await renderWith(
      ["financeiro.contas.read"],
      <RequirePermission permission="financeiro.contas.read" mode="page">
        <AreaScreen />
      </RequirePermission>,
    );
    expect(await screen.findByText("dados da área")).toBeTruthy();
    expect(screen.queryByRole("heading", { name: "Sem acesso" })).toBeNull();
    expect(calls.count).toBe(1);
  });

  it("com várias permissões exigidas, falta de uma basta para negar", async () => {
    await renderWith(
      ["financeiro.contas.read"],
      <RequirePermission permission={["financeiro.contas.read", "financeiro.contas.write"]} mode="page">
        <p>tela da área</p>
      </RequirePermission>,
    );
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    expect(screen.queryByText("tela da área")).toBeNull();
  });
});

// FND-04 AC3: ações exibidas só com a permissão efetiva.
describe("modo ação", () => {
  it("sem a permissão, não renderiza nada", async () => {
    await renderWith(
      ["financeiro.contas.read"],
      <RequirePermission permission="financeiro.contas.write" mode="action">
        <button>Nova conta</button>
      </RequirePermission>,
    );
    expect(screen.queryByRole("button", { name: "Nova conta" })).toBeNull();
    expect(screen.queryByText("Sem acesso")).toBeNull();
  });

  it("com a permissão, renderiza a ação", async () => {
    await renderWith(
      ["financeiro.contas.write"],
      <RequirePermission permission="financeiro.contas.write" mode="action">
        <button>Nova conta</button>
      </RequirePermission>,
    );
    expect(await screen.findByRole("button", { name: "Nova conta" })).toBeTruthy();
  });
});
