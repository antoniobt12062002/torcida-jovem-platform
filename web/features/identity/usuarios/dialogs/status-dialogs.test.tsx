import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, ids, problem, type User, user } from "@/test/msw/fixtures";

import { cellsOf, chooseAction, mockListUsers, mockWrite, noContent, openRowMenu } from "../test-helpers";
import { UsersPage } from "../users-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const UPDATER = ["identity:user:read", "identity:user:update"];

const ativa = user({ name: "Eva Ativa", email: "eva@example.com", active: true });
const inativa = user({
  id: "00000000-0000-4000-8000-0000000000f6",
  name: "Fabio Inativo",
  email: "fabio@example.com",
  active: false,
});
const eu = user({ id: ids.user, name: "Pessoa Teste", email: "pessoa.teste@example.com" });

async function renderPage(permissions: string[], items: () => User[]) {
  mockListUsers(() => ({ items: items(), next_cursor: null }));
  mockMe(authContext({ permissions }));
  renderWithSession(<UsersPage />);
  await screen.findByText(items()[0].email);
}

function situation(email: string): string {
  return cellsOf(email)[2];
}

// Caso de borda e USR-03: ações conforme permissão, estado e própria linha.
describe("menu de ações da linha", () => {
  it("só leitura: nenhuma ação em nenhuma linha", async () => {
    await renderPage(["identity:user:read"], () => [ativa, inativa]);
    expect(screen.queryByRole("button", { name: /^Ações de / })).toBeNull();
  });

  it("com identity:user:update: Desativar para ativo e Reativar para inativo", async () => {
    await renderPage(UPDATER, () => [ativa, inativa]);
    expect(await openRowMenu("Eva Ativa")).toEqual(["Desativar"]);
    fireEvent.keyDown(screen.getByRole("menu"), { key: "Escape" });
    await waitFor(() => expect(screen.queryByRole("menu")).toBeNull());
    expect(await openRowMenu("Fabio Inativo")).toEqual(["Reativar"]);
  });

  it("na própria linha, nenhuma ação, mesmo com todas as permissões", async () => {
    await renderPage(
      [
        "identity:user:read",
        "identity:user:update",
        "identity:role:assign",
        "identity:admin:grant",
        "identity:admin:revoke",
        "identity:user:reset_password",
      ],
      () => [eu, ativa],
    );
    expect(screen.queryByRole("button", { name: "Ações de Pessoa Teste" })).toBeNull();
    expect(screen.getByRole("button", { name: "Ações de Eva Ativa" })).toBeTruthy();
  });
});

// USR-03 AC1 e AC3.
describe("desativar", () => {
  it("pede confirmação avisando do fim das sessões e, em 204, mostra inativo", async () => {
    let current = ativa;
    const requests = mockWrite("post", `/api/v1/users/${ativa.id}/deactivate`, () => {
      current = { ...ativa, active: false };
      return noContent();
    });
    await renderPage(UPDATER, () => [current]);
    expect(situation("eva@example.com")).toBe("Ativo");

    await chooseAction("Eva Ativa", "Desativar");
    const dialog = await screen.findByRole("alertdialog", { name: "Desativar Eva Ativa?" });
    expect(dialog.textContent).toContain("as sessões dela serão encerradas");
    // Nada é chamado antes de confirmar.
    expect(requests).toEqual([]);

    fireEvent.click(screen.getByRole("button", { name: "Desativar" }));

    await waitFor(() => expect(situation("eva@example.com")).toBe("Inativo"));
    expect(await screen.findByText("Usuário desativado.")).toBeTruthy();
    expect(requests).toEqual([{ body: null, csrf: "csrf-token-de-teste" }]);
  });
});

// USR-03 AC2.
describe("reativar", () => {
  it("em 204, mostra ativo e avisa que papéis e vínculo não são restaurados", async () => {
    let current = inativa;
    const requests = mockWrite("post", `/api/v1/users/${inativa.id}/reactivate`, () => {
      current = { ...inativa, active: true };
      return noContent();
    });
    await renderPage(UPDATER, () => [current]);

    await chooseAction("Fabio Inativo", "Reativar");
    await screen.findByRole("alertdialog", { name: "Reativar Fabio Inativo?" });
    fireEvent.click(screen.getByRole("button", { name: "Reativar" }));

    await waitFor(() => expect(situation("fabio@example.com")).toBe("Ativo"));
    expect(
      await screen.findByText(
        "Usuário reativado. Papéis administrativos e vínculo administrativo não são restaurados.",
      ),
    ).toBeTruthy();
    expect(requests).toEqual([{ body: null, csrf: "csrf-token-de-teste" }]);
  });
});

// USR-03 AC4.
describe("erros da API mantêm o estado exibido", () => {
  it.each([
    [409, "last_admin", "Não é possível: o sistema ficaria sem ninguém que possa conceder acesso."],
    [403, "self_change_forbidden", "Você não pode fazer isso com a sua própria conta."],
    [403, "privilege_escalation", "Você não pode conceder ou alterar permissões que não tem."],
  ])("%i %s", async (status, code, message) => {
    mockWrite("post", `/api/v1/users/${ativa.id}/deactivate`, () => problem(status, code));
    await renderPage(UPDATER, () => [ativa]);

    await chooseAction("Eva Ativa", "Desativar");
    await screen.findByRole("alertdialog");
    fireEvent.click(screen.getByRole("button", { name: "Desativar" }));

    expect(await screen.findByText(message)).toBeTruthy();
    expect(situation("eva@example.com")).toBe("Ativo");
    expect(screen.queryByText("Usuário desativado.")).toBeNull();
  });

  it("409 last_admin na reativação também mostra a mensagem", async () => {
    mockWrite("post", `/api/v1/users/${inativa.id}/reactivate`, () => problem(409, "last_admin"));
    await renderPage(UPDATER, () => [inativa]);
    await chooseAction("Fabio Inativo", "Reativar");
    await screen.findByRole("alertdialog");
    fireEvent.click(screen.getByRole("button", { name: "Reativar" }));
    expect(
      await screen.findByText("Não é possível: o sistema ficaria sem ninguém que possa conceder acesso."),
    ).toBeTruthy();
    expect(situation("fabio@example.com")).toBe("Inativo");
  });
});
