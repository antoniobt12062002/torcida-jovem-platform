import { fireEvent, screen, waitFor } from "@testing-library/react";
import { HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, problem, type User, user } from "@/test/msw/fixtures";

import {
  describedByText,
  isInvalid,
  mockListUsers,
  mockWrite,
  type,
} from "../test-helpers";
import { UsersPage } from "../users-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const NAME = "Nome";
const EMAIL = "E-mail";
const PASSWORD = "Senha inicial";

const created = user({
  id: "00000000-0000-4000-8000-0000000000d4",
  name: "Daniela Nova",
  email: "daniela@example.com",
  must_change_password: true,
});

async function renderPage(permissions: string[]) {
  mockMe(authContext({ permissions }));
  renderWithSession(<UsersPage />);
  await screen.findByRole("heading", { name: "Usuários" });
}

async function openCreate() {
  fireEvent.click(await screen.findByRole("button", { name: "Novo usuário" }));
  await screen.findByRole("dialog", { name: "Novo usuário" });
}

function fill() {
  type(NAME, "Daniela Nova");
  type(EMAIL, "daniela@example.com");
  type(PASSWORD, "senha-inicial-123");
}

function submit() {
  fireEvent.click(screen.getByRole("button", { name: "Criar usuário" }));
}

// USR-02 AC4.
describe("visibilidade", () => {
  it("sem identity:user:create, não mostra Novo usuário", async () => {
    mockListUsers(() => ({ items: [], next_cursor: null }));
    await renderPage(["identity:user:read"]);
    await screen.findByText("Nenhum usuário encontrado.");
    expect(screen.queryByRole("button", { name: "Novo usuário" })).toBeNull();
  });
});

// USR-02 AC1.
describe("criação com sucesso", () => {
  it("envia nome, e-mail e senha, avisa da troca no primeiro acesso e mostra a pessoa na lista", async () => {
    let items: User[] = [];
    mockListUsers(() => ({ items, next_cursor: null }));
    const requests = mockWrite("post", "/api/v1/users", () => {
      items = [created];
      return HttpResponse.json(created, { status: 201 });
    });
    await renderPage(["identity:user:read", "identity:user:create"]);
    await screen.findByText("Nenhum usuário encontrado.");

    await openCreate();
    fill();
    submit();

    expect(
      await screen.findByText("Usuário criado. A pessoa terá de trocar a senha no primeiro acesso."),
    ).toBeTruthy();
    expect(await screen.findByText("daniela@example.com")).toBeTruthy();
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Novo usuário" })).toBeNull());
    expect(requests).toEqual([
      {
        body: { name: "Daniela Nova", email: "daniela@example.com", password: "senha-inicial-123" },
        csrf: "csrf-token-de-teste",
      },
    ]);
  });
});

// USR-02 AC2 e AC3.
describe("erros de campo", () => {
  it.each([
    [409, "email_taken", EMAIL, "Já existe um usuário com este e-mail."],
    [422, "invalid_email", EMAIL, "Informe um e-mail válido."],
    [422, "invalid_name", NAME, "Informe um nome válido."],
    [422, "password_too_short", PASSWORD, "A senha é curta demais. Mínimo de 8 caracteres."],
    [422, "password_too_long", PASSWORD, "A senha é longa demais."],
    [
      422,
      "password_compromised",
      PASSWORD,
      "Esta senha aparece em vazamentos conhecidos. Escolha outra.",
    ],
  ])("%i %s no campo certo", async (status, code, label, message) => {
    mockListUsers(() => ({ items: [], next_cursor: null }));
    mockWrite("post", "/api/v1/users", () => problem(status, code));
    await renderPage(["identity:user:read", "identity:user:create"]);
    await openCreate();
    fill();
    submit();

    expect(await screen.findByText(message)).toBeTruthy();
    const field = screen.getByLabelText(label);
    expect(isInvalid(field)).toBe(true);
    expect(describedByText(field)).toContain(message);
    for (const other of [NAME, EMAIL, PASSWORD].filter((l) => l !== label)) {
      expect(isInvalid(screen.getByLabelText(other))).toBe(false);
    }
    // O diálogo continua aberto, sem aviso de sucesso.
    expect(screen.getByRole("dialog", { name: "Novo usuário" })).toBeTruthy();
    expect(screen.queryByText(/Usuário criado/)).toBeNull();
  });
});
