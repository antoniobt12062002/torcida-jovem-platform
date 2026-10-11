import { screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, saldoFinanceiro } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import FinanceiroSaldoPage from "./page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

function mockSaldo() {
  const calls = { count: 0 };
  server.use(
    http.get(apiUrl("/api/v1/financeiro/saldo"), () => {
      calls.count += 1;
      return HttpResponse.json(saldoFinanceiro(4200));
    }),
  );
  return calls;
}

// FWB-02 AC1 e FND-04 AC4: /financeiro/saldo exige financeiro:saldo:read.
describe("/financeiro/saldo", () => {
  it("com financeiro:saldo:read, chama GET /saldo e mostra o valor", async () => {
    const calls = mockSaldo();
    mockMe(authContext({ permissions: ["financeiro:saldo:read"] }));
    renderWithSession(<FinanceiroSaldoPage />);
    expect(await screen.findByText("R$ 42,00")).toBeTruthy();
    expect(calls.count).toBe(1);
  });

  it("sem financeiro:saldo:read, mostra Sem acesso e não chama a API", async () => {
    const calls = mockSaldo();
    mockMe(authContext({ permissions: ["financeiro:conta:read"] }));
    renderWithSession(<FinanceiroSaldoPage />);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(calls.count).toBe(0);
    expect(screen.queryByRole("heading", { name: "Saldo" })).toBeNull();
  });
});
