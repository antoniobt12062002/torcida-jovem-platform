import { screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, conta } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import FinanceiroContasPage from "./page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

function mockContas() {
  const calls = { count: 0 };
  server.use(
    http.get(apiUrl("/api/v1/financeiro/contas"), () => {
      calls.count += 1;
      return HttpResponse.json({ items: [conta({ nome: "Mensalidades" })] });
    }),
  );
  return calls;
}

// FWB-01 AC1 e FND-04 AC4: /financeiro/contas exige financeiro:conta:read.
describe("/financeiro/contas", () => {
  it("com financeiro:conta:read, chama GET /contas e mostra o plano", async () => {
    const calls = mockContas();
    mockMe(authContext({ permissions: ["financeiro:conta:read"] }));
    renderWithSession(<FinanceiroContasPage />);
    expect(await screen.findByText("Mensalidades")).toBeTruthy();
    expect(screen.getByRole("heading", { name: "Plano de contas" })).toBeTruthy();
    expect(calls.count).toBe(1);
  });

  it("sem financeiro:conta:read, mostra Sem acesso e não chama a API", async () => {
    const calls = mockContas();
    mockMe(authContext({ permissions: ["financeiro:saldo:read"] }));
    renderWithSession(<FinanceiroContasPage />);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(calls.count).toBe(0);
    expect(screen.queryByRole("heading", { name: "Plano de contas" })).toBeNull();
  });
});
