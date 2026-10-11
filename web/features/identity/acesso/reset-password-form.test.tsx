import { fireEvent, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { afterEach, beforeEach, describe, expect, it, type MockInstance, vi } from "vitest";

import { queryKeys } from "@/lib/api/query-keys";
import { navigations, resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, problem } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { ResetPasswordForm } from "./reset-password-form";
import { describedByText, isInvalid, type } from "./test-helpers";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

const TOKEN = "tok_Reset-0123456789abcdef";
const NEW = "Nova senha";
const CONFIRM = "Confirmar nova senha";
const SUBMIT = "Redefinir senha";

let consoleSpies: MockInstance[] = [];

beforeEach(() => {
  resetNavigation();
  window.localStorage.clear();
  window.sessionStorage.clear();
  consoleSpies = (["log", "info", "warn", "error", "debug"] as const).map((m) =>
    vi.spyOn(console, m),
  );
});

afterEach(() => {
  for (const spy of consoleSpies) spy.mockRestore();
});

function mockConfirm(respond: () => Response) {
  const requests: { url: string; body: unknown }[] = [];
  server.use(
    http.post(apiUrl("/api/v1/auth/password-reset/confirm"), async ({ request }) => {
      requests.push({ url: request.url, body: await request.json() });
      return respond();
    }),
  );
  return requests;
}

async function openLink(hash: string, { session = null as ReturnType<typeof authContext> | null } = {}) {
  setLocation(`/redefinir-senha${hash}`);
  mockMe(session);
  return renderWithSession(<ResetPasswordForm />);
}

function fill(next: string, confirm = next) {
  type(NEW, next);
  type(CONFIRM, confirm);
}

function submit() {
  fireEvent.click(screen.getByRole("button", { name: SUBMIT }));
}

/** O token não pode estar na URL, no armazenamento do navegador nem em log. */
function expectTokenNotLeaked() {
  expect(window.location.href).not.toContain(TOKEN);
  expect(window.location.hash).toBe("");
  const stored = [window.localStorage, window.sessionStorage].flatMap((s) =>
    Object.keys(s).flatMap((k) => [k, s.getItem(k) ?? ""]),
  );
  expect(stored.join(" ")).not.toContain(TOKEN);
  const logged = consoleSpies.flatMap((spy) => spy.mock.calls.flat().map((a) => String(a)));
  expect(logged.join(" ")).not.toContain(TOKEN);
}

// ACS-03 AC1.
describe("abrir o link com token", () => {
  it("remove o fragmento da barra de endereço sem recarregar e mostra o formulário", async () => {
    await openLink(`#token=${TOKEN}`);
    expect(await screen.findByRole("button", { name: SUBMIT })).toBeTruthy();
    expect(window.location.pathname).toBe("/redefinir-senha");
    expect(window.location.hash).toBe("");
    // Remoção com history.replaceState: nenhuma navegação do roteador.
    expect(navigations()).toEqual([]);
    expectTokenNotLeaked();
  });
});

// ACS-03 AC2.
describe("abrir sem token", () => {
  it.each(["", "#token=", "#outro=1"])("informa link inválido e oferece pedir outro (%s)", async (hash) => {
    await openLink(hash);
    expect(await screen.findByText("O link de recuperação é inválido.")).toBeTruthy();
    expect(screen.getByRole("link", { name: "Pedir um novo link" }).getAttribute("href")).toBe(
      "/recuperar-acesso",
    );
    expect(screen.queryByRole("button", { name: SUBMIT })).toBeNull();
  });
});

// ACS-03 AC3 e AC6.
describe("redefinição com sucesso", () => {
  it("envia token e nova senha no corpo e leva a /entrar com aviso", async () => {
    const requests = mockConfirm(() => new HttpResponse(null, { status: 204 }));
    await openLink(`#token=${TOKEN}`);
    await screen.findByRole("button", { name: SUBMIT });
    fill("senha-nova-123");
    submit();
    await waitFor(() => expect(navigations()).toEqual(["/entrar"]));
    expect(requests).toEqual([
      {
        url: apiUrl("/api/v1/auth/password-reset/confirm"),
        body: { token: TOKEN, new_password: "senha-nova-123" },
      },
    ]);
    expect(new URL(requests[0].url).search).toBe("");
    expect(await screen.findByText("Senha redefinida. Entre com a nova senha.")).toBeTruthy();
    expectTokenNotLeaked();
  });

  it("com sessão aberta, redefine normalmente, esquece a sessão e leva a /entrar", async () => {
    mockConfirm(() => new HttpResponse(null, { status: 204 }));
    const { queryClient } = await openLink(`#token=${TOKEN}`, { session: authContext() });
    await waitFor(() => expect(queryClient.getQueryData(queryKeys.session.me)).toBeTruthy());
    fill("senha-nova-123");
    submit();
    await waitFor(() => expect(navigations()).toEqual(["/entrar"]));
    // A API encerrou todas as sessões: o front não continua como autenticado.
    expect(queryClient.getQueryData(queryKeys.session.me)).toBeNull();
  });
});

// ACS-03 AC3 (confirmação só no front).
describe("confirmação diferente da nova senha", () => {
  it("mostra o erro no campo de confirmação e não chama a API", async () => {
    const requests = mockConfirm(() => new HttpResponse(null, { status: 204 }));
    await openLink(`#token=${TOKEN}`);
    await screen.findByRole("button", { name: SUBMIT });
    fill("senha-nova-123", "senha-nova-321");
    submit();
    const message = "A confirmação não confere com a nova senha.";
    expect(await screen.findByText(message)).toBeTruthy();
    expect(isInvalid(screen.getByLabelText(CONFIRM))).toBe(true);
    expect(describedByText(screen.getByLabelText(CONFIRM))).toContain(message);
    expect(requests).toEqual([]);
  });
});

// ACS-03 AC4.
describe("token expirado ou usado", () => {
  it("informa que o link expirou ou já foi usado e oferece pedir outro", async () => {
    mockConfirm(() => problem(400, "invalid_reset_token"));
    await openLink(`#token=${TOKEN}`);
    await screen.findByRole("button", { name: SUBMIT });
    fill("senha-nova-123");
    submit();
    expect(await screen.findByText("O link expirou ou já foi usado.")).toBeTruthy();
    expect(screen.getByRole("link", { name: "Pedir um novo link" }).getAttribute("href")).toBe(
      "/recuperar-acesso",
    );
    expect(navigations()).toEqual([]);
    expectTokenNotLeaked();
  });
});

// ACS-03 AC5.
describe("erros de senha da API", () => {
  it.each([
    [
      "password_too_short",
      "A senha é curta demais. Mínimo de 8 caracteres (10 para quem tem vínculo administrativo).",
    ],
    ["password_too_long", "A senha é longa demais."],
    ["password_compromised", "Esta senha aparece em vazamentos conhecidos. Escolha outra."],
  ])("%s fica no campo de nova senha e o token vale para nova tentativa", async (code, message) => {
    let attempt = 0;
    const requests = mockConfirm(() =>
      attempt++ === 0 ? problem(422, code) : new HttpResponse(null, { status: 204 }),
    );
    await openLink(`#token=${TOKEN}`);
    await screen.findByRole("button", { name: SUBMIT });
    fill("senha-ruim");
    submit();
    expect(await screen.findByText(message)).toBeTruthy();
    const field = screen.getByLabelText(NEW);
    expect(isInvalid(field)).toBe(true);
    expect(describedByText(field)).toContain(message);
    expect(navigations()).toEqual([]);

    fill("senha-nova-melhor-123");
    submit();
    await waitFor(() => expect(navigations()).toEqual(["/entrar"]));
    expect(requests.map((r) => r.body)).toEqual([
      { token: TOKEN, new_password: "senha-ruim" },
      { token: TOKEN, new_password: "senha-nova-melhor-123" },
    ]);
    expectTokenNotLeaked();
  });
});
