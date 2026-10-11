import { fireEvent, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, problem, saldoFinanceiro } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { SaldoPage } from "./saldo-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

// FWB-02: saldo.

const SALDO = "/api/v1/financeiro/saldo";

function mockSaldo(respond: () => Response) {
  const calls = { count: 0 };
  server.use(
    http.get(apiUrl(SALDO), () => {
      calls.count += 1;
      return respond();
    }),
  );
  return calls;
}

function renderPage(permissions = ["financeiro:saldo:read"]) {
  mockMe(authContext({ permissions }));
  renderWithSession(<SaldoPage />);
}

// FWB-02 AC1.
describe("valor do saldo", () => {
  it.each([
    ["positivo", 123456, "R$ 1.234,56"],
    ["zero", 0, "R$ 0,00"],
    ["negativo", -98765, "-R$ 987,65"],
  ])("%s: chama GET /saldo e mostra saldo_cents com formatBRL", async (_caso, cents, texto) => {
    const calls = mockSaldo(() => HttpResponse.json(saldoFinanceiro(cents)));
    renderPage();
    expect(await screen.findByText(texto)).toBeTruthy();
    expect(screen.getByRole("heading", { name: "Saldo" })).toBeTruthy();
    expect(calls.count).toBe(1);
  });
});

// FWB-02 AC2.
it("explica que o saldo é recebidos menos pagos, pelo status atual", async () => {
  mockSaldo(() => HttpResponse.json(saldoFinanceiro(500)));
  renderPage();
  await screen.findByText("R$ 5,00");
  expect(
    screen.getByText(
      "O saldo é a soma dos lançamentos recebidos menos a soma dos lançamentos pagos, pelo status atual de cada lançamento.",
    ),
  ).toBeTruthy();
});

// FND-05 AC3 e FND-06 AC5.
it("falha ao carregar mostra o erro e tenta de novo", async () => {
  let fail = true;
  mockSaldo(() => (fail ? problem(503, "service_unavailable") : HttpResponse.json(saldoFinanceiro(100))));
  renderPage();
  expect(
    await screen.findByText("O serviço está indisponível no momento. Tente de novo em instantes."),
  ).toBeTruthy();
  fail = false;
  fireEvent.click(screen.getByRole("button", { name: "Tentar de novo" }));
  expect(await screen.findByText("R$ 1,00")).toBeTruthy();
});
