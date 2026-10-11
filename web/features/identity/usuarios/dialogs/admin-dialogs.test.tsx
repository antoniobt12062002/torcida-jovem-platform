import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, problem, type User, user } from "@/test/msw/fixtures";

import {
  cellsOf,
  chooseAction,
  mockListUsers,
  mockWrite,
  noContent,
  openRowMenu,
  type,
} from "../test-helpers";
import { UsersPage } from "../users-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const ADMIN = [
  "identity:user:read",
  "identity:admin:grant",
  "identity:role:assign",
  "identity:admin:revoke",
];

const MEMBERSHIP = { reason: "Assumiu a tesouraria", granted_at: "2026-10-01T12:00:00Z" };

const associado = user({ name: "Gabi Associada", email: "gabi@example.com" });
const tesoureira = user({
  id: "00000000-0000-4000-8000-0000000000a7",
  name: "Hana Tesoureira",
  email: "hana@example.com",
  roles: ["ASSOCIADO", "TESOURARIA"],
  admin_membership: MEMBERSHIP,
});

const ADMIN_ROLE_LABELS = [
  "Presidente",
  "Diretoria",
  "Tesouraria",
  "Estoque e loja",
  "Eventos",
  "Conselho Fiscal",
  "Administração do sistema",
];

async function renderPage(permissions: string[], items: () => User[]) {
  mockListUsers(() => ({ items: items(), next_cursor: null }));
  mockMe(authContext({ permissions }));
  renderWithSession(<UsersPage />);
  await screen.findByText(items()[0].email);
}

function checkbox(label: string) {
  return screen.getByRole("checkbox", { name: label });
}

function isChecked(label: string) {
  return checkbox(label).getAttribute("aria-checked") === "true";
}

function button(name: string) {
  return screen.getByRole("button", { name }) as HTMLButtonElement;
}

// USR-04 AC6.
describe("ação compatível com o estado", () => {
  it("sem vínculo: só Promover; com vínculo: Editar papéis e Retirar acesso", async () => {
    await renderPage(ADMIN, () => [associado, tesoureira]);
    expect(await openRowMenu("Gabi Associada")).toEqual(["Promover a administrativo"]);
    fireEvent.keyDown(screen.getByRole("menu"), { key: "Escape" });
    await waitFor(() => expect(screen.queryByRole("menu")).toBeNull());
    expect(await openRowMenu("Hana Tesoureira")).toEqual([
      "Editar papéis",
      "Retirar acesso administrativo",
    ]);
  });

  it("cada ação exige a sua permissão", async () => {
    await renderPage(["identity:user:read", "identity:admin:grant"], () => [associado, tesoureira]);
    expect(await openRowMenu("Gabi Associada")).toEqual(["Promover a administrativo"]);
    expect(screen.queryByRole("button", { name: "Ações de Hana Tesoureira" })).toBeNull();
  });
});

// USR-04 AC1 e AC2.
describe("promover", () => {
  async function openGrant() {
    await chooseAction("Gabi Associada", "Promover a administrativo");
    return screen.findByRole("dialog", { name: "Promover Gabi Associada a administrativo" });
  }

  it("oferece os sete papéis administrativos, sem Associado", async () => {
    await renderPage(ADMIN, () => [associado]);
    const dialog = await openGrant();
    const labels = within(dialog)
      .getAllByRole("checkbox")
      .map((box) => document.getElementById(box.getAttribute("aria-labelledby") ?? "")?.textContent);
    expect(labels).toEqual(ADMIN_ROLE_LABELS);
  });

  it("mantém o envio desabilitado sem papel ou com motivo de menos de 10 caracteres úteis", async () => {
    const requests = mockWrite("post", `/api/v1/users/${associado.id}/admin-membership`, noContent);
    await renderPage(ADMIN, () => [associado]);
    await openGrant();

    // Motivo válido, nenhum papel.
    type("Motivo", "Assumiu a tesouraria");
    expect(button("Promover").disabled).toBe(true);

    // Papel escolhido, motivo com 9 caracteres úteis entre espaços.
    fireEvent.click(checkbox("Tesouraria"));
    type("Motivo", "   123456789   ");
    expect(button("Promover").disabled).toBe(true);
    fireEvent.click(button("Promover"));

    // Exatamente 10 caracteres úteis habilita.
    type("Motivo", "  1234567890  ");
    expect(button("Promover").disabled).toBe(false);
    expect(requests).toEqual([]);
  });

  it("envia papéis e motivo e, em 204, avisa do fim das sessões e da troca de senha", async () => {
    let current = associado;
    const requests = mockWrite("post", `/api/v1/users/${associado.id}/admin-membership`, () => {
      current = { ...associado, roles: ["ASSOCIADO", "TESOURARIA", "DIRETORIA"], admin_membership: MEMBERSHIP };
      return noContent();
    });
    await renderPage(ADMIN, () => [current]);
    await openGrant();
    fireEvent.click(checkbox("Tesouraria"));
    fireEvent.click(checkbox("Diretoria"));
    type("Motivo", "  Assumiu a tesouraria  ");
    fireEvent.click(button("Promover"));

    expect(
      await screen.findByText(
        "Acesso administrativo concedido. As sessões da pessoa foram encerradas e ela terá de trocar a senha.",
      ),
    ).toBeTruthy();
    await waitFor(() => expect(cellsOf("gabi@example.com")[4]).toBe("Sim"));
    expect(requests).toEqual([
      {
        body: { roles: ["DIRETORIA", "TESOURARIA"], reason: "Assumiu a tesouraria" },
        csrf: "csrf-token-de-teste",
      },
    ]);
  });
});

// USR-04 AC5.
describe("erros da API em português", () => {
  it.each([
    [409, "already_admin", "Esta pessoa já tem acesso administrativo."],
    [409, "not_admin", "Esta pessoa não tem acesso administrativo."],
    [
      409,
      "admin_membership_required",
      "Só é possível definir papéis administrativos para quem tem acesso administrativo.",
    ],
    [403, "privilege_escalation", "Você não pode conceder ou alterar permissões que não tem."],
    [403, "self_change_forbidden", "Você não pode fazer isso com a sua própria conta."],
    [409, "last_admin", "Não é possível: o sistema ficaria sem ninguém que possa conceder acesso."],
    [409, "user_inactive", "Esta pessoa está desativada. Reative a conta antes."],
    [422, "unknown_role", "Um dos papéis escolhidos não existe."],
    [422, "admin_role_required", "Escolha ao menos um papel administrativo."],
    [422, "reason_required", "Informe um motivo com pelo menos 10 caracteres."],
  ])("promoção com %i %s", async (status, code, message) => {
    mockWrite("post", `/api/v1/users/${associado.id}/admin-membership`, () => problem(status, code));
    await renderPage(ADMIN, () => [associado]);
    await chooseAction("Gabi Associada", "Promover a administrativo");
    await screen.findByRole("dialog");
    fireEvent.click(checkbox("Tesouraria"));
    type("Motivo", "Assumiu a tesouraria");
    fireEvent.click(button("Promover"));

    expect(await screen.findByText(message)).toBeTruthy();
    // O diálogo continua aberto e a lista não muda.
    expect(screen.getByRole("dialog")).toBeTruthy();
    expect(cellsOf("gabi@example.com")[4]).toBe("Não");
  });

  it("edição de papéis com 403 privilege_escalation mostra a mensagem", async () => {
    mockWrite("put", `/api/v1/users/${tesoureira.id}/roles`, () => problem(403, "privilege_escalation"));
    await renderPage(ADMIN, () => [tesoureira]);
    await chooseAction("Hana Tesoureira", "Editar papéis");
    await screen.findByRole("dialog");
    fireEvent.click(checkbox("Presidente"));
    fireEvent.click(button("Salvar papéis"));
    expect(
      await screen.findByText("Você não pode conceder ou alterar permissões que não tem."),
    ).toBeTruthy();
    expect(cellsOf("hana@example.com")[3]).toBe("Associado, Tesouraria");
  });
});

// USR-04 AC3.
describe("editar papéis", () => {
  async function openSetRoles() {
    await chooseAction("Hana Tesoureira", "Editar papéis");
    await screen.findByRole("dialog", { name: "Editar papéis de Hana Tesoureira" });
  }

  it("começa com os papéis atuais e envia a lista nova", async () => {
    let current = tesoureira;
    const requests = mockWrite("put", `/api/v1/users/${tesoureira.id}/roles`, () => {
      current = { ...tesoureira, roles: ["ASSOCIADO", "TESOURARIA", "EVENTOS"] };
      return noContent();
    });
    await renderPage(ADMIN, () => [current]);
    await openSetRoles();
    expect(ADMIN_ROLE_LABELS.filter(isChecked)).toEqual(["Tesouraria"]);

    fireEvent.click(checkbox("Eventos"));
    fireEvent.click(button("Salvar papéis"));

    expect(await screen.findByText("Papéis atualizados.")).toBeTruthy();
    await waitFor(() => expect(cellsOf("hana@example.com")[3]).toBe("Associado, Tesouraria, Eventos"));
    expect(requests).toEqual([
      { body: { roles: ["TESOURARIA", "EVENTOS"] }, csrf: "csrf-token-de-teste" },
    ]);
  });

  it("lista vazia: avisa que o vínculo fica dormente antes de enviar", async () => {
    const requests = mockWrite("put", `/api/v1/users/${tesoureira.id}/roles`, noContent);
    await renderPage(ADMIN, () => [tesoureira]);
    await openSetRoles();
    fireEvent.click(checkbox("Tesouraria"));
    fireEvent.click(button("Salvar papéis"));

    expect(await screen.findByText(/o vínculo administrativo fica dormente/)).toBeTruthy();
    expect(requests).toEqual([]);

    fireEvent.click(button("Salvar sem papéis"));
    expect(await screen.findByText("Papéis atualizados.")).toBeTruthy();
    expect(requests).toEqual([{ body: { roles: [] }, csrf: "csrf-token-de-teste" }]);
  });
});

// USR-04 AC4.
describe("retirar acesso", () => {
  it("confirma com motivo e mostra a pessoa só com Associado", async () => {
    let current = tesoureira;
    const requests = mockWrite("post", `/api/v1/users/${tesoureira.id}/admin-membership/revoke`, () => {
      current = { ...tesoureira, roles: ["ASSOCIADO"], admin_membership: null };
      return noContent();
    });
    await renderPage(ADMIN, () => [current]);
    await chooseAction("Hana Tesoureira", "Retirar acesso administrativo");
    await screen.findByRole("alertdialog", {
      name: "Retirar o acesso administrativo de Hana Tesoureira?",
    });
    expect(button("Retirar acesso").disabled).toBe(true);

    type("Motivo", "  Deixou a tesouraria  ");
    fireEvent.click(button("Retirar acesso"));

    expect(
      await screen.findByText("Acesso administrativo retirado. A pessoa fica só com o papel Associado."),
    ).toBeTruthy();
    await waitFor(() => expect(cellsOf("hana@example.com")[3]).toBe("Associado"));
    expect(cellsOf("hana@example.com")[4]).toBe("Não");
    expect(requests).toEqual([{ body: { reason: "Deixou a tesouraria" }, csrf: "csrf-token-de-teste" }]);
  });

  it("409 last_admin mostra a mensagem e mantém o vínculo", async () => {
    mockWrite("post", `/api/v1/users/${tesoureira.id}/admin-membership/revoke`, () =>
      problem(409, "last_admin"),
    );
    await renderPage(ADMIN, () => [tesoureira]);
    await chooseAction("Hana Tesoureira", "Retirar acesso administrativo");
    await screen.findByRole("alertdialog");
    type("Motivo", "Deixou a tesouraria");
    fireEvent.click(button("Retirar acesso"));
    expect(
      await screen.findByText("Não é possível: o sistema ficaria sem ninguém que possa conceder acesso."),
    ).toBeTruthy();
    expect(cellsOf("hana@example.com")[4]).toBe("Sim");
  });
});
