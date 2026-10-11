import { fireEvent, screen, waitFor } from "@testing-library/react";
import { delay, http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { navigations, resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, problem } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import AppLayout from "./layout";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

function renderGate() {
  return renderWithSession(
    <AppLayout>
      <p>conteúdo protegido</p>
    </AppLayout>,
  );
}

// FND-04 AC2.
it("enquanto consulta a sessão, mostra carregando e nada do conteúdo", async () => {
  setLocation("/financeiro/contas");
  server.use(
    http.get(apiUrl("/api/v1/auth/me"), async () => {
      await delay("infinite");
      return HttpResponse.json(authContext());
    }),
  );
  renderGate();
  expect(screen.getByRole("status").textContent).toContain("Carregando");
  expect(screen.queryByText("conteúdo protegido")).toBeNull();
});

// FND-04 AC1.
describe("sem sessão", () => {
  it("leva a /entrar com o caminho pedido em next, sem mostrar o conteúdo", async () => {
    setLocation("/financeiro/contas?tipo=RECEITA");
    mockMe(null);
    renderGate();
    await waitFor(() => expect(navigations()).toHaveLength(1));
    const url = new URL(navigations()[0], window.location.origin);
    expect(url.pathname).toBe("/entrar");
    expect(url.searchParams.get("next")).toBe("/financeiro/contas?tipo=RECEITA");
    expect(screen.queryByText("conteúdo protegido")).toBeNull();
  });
});

describe("com sessão", () => {
  it("mostra o conteúdo", async () => {
    setLocation("/financeiro/contas");
    mockMe(authContext());
    renderGate();
    expect(await screen.findByText("conteúdo protegido")).toBeTruthy();
    expect(navigations()).toEqual([]);
  });

  // FND-03 AC4.
  it("com troca obrigatória fora de /conta/senha, leva a /conta/senha sem mostrar o conteúdo", async () => {
    setLocation("/financeiro/contas");
    mockMe(authContext({ must_change_password: true }));
    renderGate();
    await waitFor(() => expect(navigations()).toEqual(["/conta/senha"]));
    expect(screen.queryByText("conteúdo protegido")).toBeNull();
  });

  it("com troca obrigatória em /conta/senha, mostra a tela de troca", async () => {
    setLocation("/conta/senha");
    mockMe(authContext({ must_change_password: true }));
    renderGate();
    expect(await screen.findByText("conteúdo protegido")).toBeTruthy();
    expect(navigations()).toEqual([]);
  });
});

// FND-06 AC5 aplicado à consulta da sessão.
describe("API indisponível ao consultar a sessão", () => {
  it("informa a indisponibilidade, oferece tentar de novo e não mostra o conteúdo", async () => {
    setLocation("/financeiro/contas");
    let calls = 0;
    server.use(
      http.get(apiUrl("/api/v1/auth/me"), () => {
        calls += 1;
        return calls === 1 ? problem(503, "service_unavailable") : HttpResponse.json(authContext());
      }),
    );
    renderGate();
    expect(await screen.findByText(/serviço está indisponível/)).toBeTruthy();
    expect(screen.queryByText("conteúdo protegido")).toBeNull();
    expect(navigations()).toEqual([]);

    fireEvent.click(screen.getByRole("button", { name: "Tentar de novo" }));
    expect(await screen.findByText("conteúdo protegido")).toBeTruthy();
  });
});
