import { fireEvent, screen, waitFor } from "@testing-library/react";
import { HttpResponse } from "msw";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, problem, type User, user } from "@/test/msw/fixtures";

import { cellsOf, chooseAction, mockListUsers, mockWrite, openRowMenu, type } from "../test-helpers";
import { UsersPage } from "../users-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

const CONSOLE_METHODS = ["log", "info", "warn", "error", "debug", "trace"] as const;
const consoleSpies: ReturnType<typeof vi.spyOn>[] = [];

beforeEach(() => {
  resetNavigation();
  localStorage.clear();
  sessionStorage.clear();
  for (const method of CONSOLE_METHODS) consoleSpies.push(vi.spyOn(console, method));
});

afterEach(() => {
  for (const spy of consoleSpies.splice(0)) spy.mockRestore();
});

const RESETTER = ["identity:user:read", "identity:user:reset_password"];
const SECRET = "Temp-Xk7-Senha-Unica-93";
const SHOWN_ONCE =
  "Esta senha é mostrada uma única vez. Entregue-a à pessoa por um canal seguro: ela terá de trocá-la no primeiro acesso.";

const pessoa = user({ name: "Iris Esquecida", email: "iris@example.com" });

async function renderPage(items: () => User[], permissions = RESETTER) {
  mockListUsers(() => ({ items: items(), next_cursor: null }));
  mockMe(authContext({ permissions }));
  const rendered = renderWithSession(<UsersPage />);
  await screen.findByText(items()[0].email);
  return rendered;
}

async function openReset() {
  await chooseAction("Iris Esquecida", "Gerar senha temporária");
  await screen.findByRole("dialog", { name: "Senha temporária para Iris Esquecida" });
}

function generateButton() {
  return screen.getByRole("button", { name: "Gerar senha temporária" }) as HTMLButtonElement;
}

function storageText(): string {
  const entries: string[] = [];
  for (const storage of [localStorage, sessionStorage]) {
    for (let i = 0; i < storage.length; i++) {
      const key = storage.key(i) ?? "";
      entries.push(key, storage.getItem(key) ?? "");
    }
  }
  return entries.join("\n");
}

describe("visibilidade", () => {
  it("sem identity:user:reset_password, não oferece a senha temporária", async () => {
    await renderPage(() => [pessoa], ["identity:user:read", "identity:user:update"]);
    expect(await openRowMenu("Iris Esquecida")).toEqual(["Desativar"]);
  });
});

// USR-05 AC1.
describe("motivo mínimo", () => {
  it("mantém o envio desabilitado com menos de 10 caracteres úteis", async () => {
    const requests = mockWrite("post", `/api/v1/users/${pessoa.id}/password-reset`, () =>
      HttpResponse.json({ temporary_password: SECRET }),
    );
    await renderPage(() => [pessoa]);
    await openReset();
    expect(generateButton().disabled).toBe(true);
    type("Motivo", "    123456789    ");
    expect(generateButton().disabled).toBe(true);
    fireEvent.click(generateButton());
    type("Motivo", " 1234567890 ");
    expect(generateButton().disabled).toBe(false);
    expect(requests).toEqual([]);
  });
});

// USR-05 AC1 e AC2.
describe("geração e descarte", () => {
  it("mostra a senha uma vez, copia, e ao fechar some do DOM, do cache, do armazenamento e do log", async () => {
    let current = pessoa;
    const requests = mockWrite("post", `/api/v1/users/${pessoa.id}/password-reset`, () => {
      current = { ...pessoa, must_change_password: true };
      return HttpResponse.json(
        { temporary_password: SECRET },
        { headers: { "Cache-Control": "no-store" } },
      );
    });
    const writeText = vi.fn<(text: string) => Promise<void>>().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", { value: { writeText }, configurable: true });

    const { queryClient } = await renderPage(() => [current]);
    await openReset();
    type("Motivo", "  Perdeu o acesso ao e-mail  ");
    fireEvent.click(generateButton());

    expect(await screen.findByText(SECRET)).toBeTruthy();
    expect(screen.getByText(SHOWN_ONCE)).toBeTruthy();
    expect(requests).toEqual([
      { body: { reason: "Perdeu o acesso ao e-mail" }, csrf: "csrf-token-de-teste" },
    ]);

    fireEvent.click(screen.getByRole("button", { name: "Copiar" }));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith(SECRET));

    // A lista reflete a troca de senha obrigatória.
    await waitFor(() => expect(cellsOf("iris@example.com")[5]).toBe("Pendente"));

    fireEvent.click(screen.getByRole("button", { name: "Fechar" }));
    await waitFor(() =>
      expect(screen.queryByRole("dialog", { name: "Senha temporária para Iris Esquecida" })).toBeNull(),
    );

    expect(document.body.innerHTML).not.toContain(SECRET);
    const queryData = queryClient
      .getQueryCache()
      .getAll()
      .map((query) => query.state.data);
    const mutationState = queryClient
      .getMutationCache()
      .getAll()
      .map((mutation) => ({ data: mutation.state.data, variables: mutation.state.variables }));
    expect(JSON.stringify(queryData)).not.toContain(SECRET);
    expect(JSON.stringify(mutationState)).not.toContain(SECRET);
    expect(storageText()).not.toContain(SECRET);
    for (const spy of consoleSpies) {
      expect(JSON.stringify(spy.mock.calls.map((args: unknown[]) => args.map(String)))).not.toContain(SECRET);
    }

    // Reabrir não mostra a senha de novo.
    await openReset();
    expect(screen.queryByText(SECRET)).toBeNull();
    expect(generateButton()).toBeTruthy();
  });
});

// USR-05 AC3.
describe("erros", () => {
  it.each([
    [422, "reason_required", "Informe um motivo com pelo menos 10 caracteres."],
    [403, "self_change_forbidden", "Você não pode fazer isso com a sua própria conta."],
    [403, "privilege_escalation", "Você não pode conceder ou alterar permissões que não tem."],
    [409, "user_inactive", "Esta pessoa está desativada. Reative a conta antes."],
  ])("%i %s mostra a mensagem", async (status, code, message) => {
    mockWrite("post", `/api/v1/users/${pessoa.id}/password-reset`, () => problem(status, code));
    await renderPage(() => [pessoa]);
    await openReset();
    type("Motivo", "Perdeu o acesso ao e-mail");
    fireEvent.click(generateButton());
    expect(await screen.findByText(message)).toBeTruthy();
    expect(screen.queryByText(SHOWN_ONCE)).toBeNull();
  });
});
