import { fireEvent, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { navigations, resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { server } from "@/test/msw/server";

import { RequestResetForm } from "./request-reset-form";
import { type } from "./test-helpers";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const NEUTRAL = "Se existir uma conta vinculada ao e-mail informado, enviaremos instruções.";

/** A API responde 202 com a mesma mensagem neutra para qualquer e-mail. */
function mockRequest() {
  const bodies: unknown[] = [];
  server.use(
    http.post(apiUrl("/api/v1/auth/password-reset/request"), async ({ request }) => {
      bodies.push(await request.json());
      return HttpResponse.json({ message: NEUTRAL }, { status: 202 });
    }),
  );
  return bodies;
}

async function requestFor(email: string) {
  setLocation("/recuperar-acesso");
  mockMe(null);
  const view = renderWithSession(<RequestResetForm />);
  await screen.findByRole("button", { name: "Enviar link" });
  type("E-mail", email);
  fireEvent.click(screen.getByRole("button", { name: "Enviar link" }));
  await screen.findByText(NEUTRAL);
  const text = view.container.textContent;
  view.unmount();
  return text;
}

// ACS-02 AC1-2.
describe("pedido de recuperação", () => {
  it("envia o e-mail e mostra a mensagem neutra devolvida pela API", async () => {
    const bodies = mockRequest();
    await requestFor("pessoa.teste@example.com");
    expect(bodies).toEqual([{ email: "pessoa.teste@example.com" }]);
    expect(navigations()).toEqual([]);
  });

  it("mostra a mesma tela para qualquer e-mail, sem indicar se existe conta", async () => {
    const bodies = mockRequest();
    const first = await requestFor("pessoa.teste@example.com");
    const second = await requestFor("ninguem@example.com");
    expect(bodies).toEqual([{ email: "pessoa.teste@example.com" }, { email: "ninguem@example.com" }]);
    expect(second).toBe(first);
    expect(first).toContain(NEUTRAL);
    await waitFor(() => expect(navigations()).toEqual([]));
  });

  it("mostra a mensagem devolvida pela API, não um texto fixo do front", async () => {
    server.use(
      http.post(apiUrl("/api/v1/auth/password-reset/request"), () =>
        HttpResponse.json({ message: "Mensagem neutra da API." }, { status: 202 }),
      ),
    );
    setLocation("/recuperar-acesso");
    mockMe(null);
    renderWithSession(<RequestResetForm />);
    type("E-mail", "pessoa.teste@example.com");
    fireEvent.click(await screen.findByRole("button", { name: "Enviar link" }));
    expect(await screen.findByText("Mensagem neutra da API.")).toBeTruthy();
  });
});
