import { fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, problem, type User, user } from "@/test/msw/fixtures";

import { chooseAction, mockListUsers, mockWrite, type } from "../test-helpers";
import { UsersPage } from "../users-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

// DSP-03 AC2, AC3 e AC5 e DSP-04 AC1, AC4 e AC7 nos diálogos de usuários: a
// confirmação nomeia a pessoa, diz a consequência e repete o verbo no botão;
// o erro da ação aparece dentro dela; os formulários usam o layout padrão.

const ALL = [
  "identity:user:read",
  "identity:user:create",
  "identity:user:update",
  "identity:admin:grant",
  "identity:role:assign",
  "identity:admin:revoke",
  "identity:user:reset_password",
];

const MEMBERSHIP = { reason: "Assumiu a tesouraria", granted_at: "2026-10-01T12:00:00Z" };

const eva = user({ name: "Eva Ativa", email: "eva@example.com" });
const fabio = user({
  id: "00000000-0000-4000-8000-0000000000f6",
  name: "Fabio Inativo",
  email: "fabio@example.com",
  active: false,
});
const hana = user({
  id: "00000000-0000-4000-8000-0000000000a7",
  name: "Hana Tesoureira",
  email: "hana@example.com",
  roles: ["ASSOCIADO", "TESOURARIA"],
  admin_membership: MEMBERSHIP,
});

async function renderPage(items: User[]) {
  mockListUsers(() => ({ items, next_cursor: null }));
  mockMe(authContext({ permissions: ALL }));
  renderWithSession(<UsersPage />);
  await screen.findByText(items[0].email);
}

type Confirmation = {
  who: User;
  action: string;
  title: string;
  consequence: string;
  confirm: string;
  destructive: boolean;
  path: string;
  reason?: string;
};

const CONFIRMATIONS: Confirmation[] = [
  {
    who: eva,
    action: "Desativar",
    title: "Desativar Eva Ativa?",
    consequence: "A pessoa não poderá mais entrar, e as sessões dela serão encerradas.",
    confirm: "Desativar",
    destructive: true,
    path: `/api/v1/users/${eva.id}/deactivate`,
  },
  {
    who: fabio,
    action: "Reativar",
    title: "Reativar Fabio Inativo?",
    consequence:
      "A pessoa volta a poder entrar só como Associado: papéis administrativos e vínculo administrativo não são restaurados.",
    confirm: "Reativar",
    destructive: false,
    path: `/api/v1/users/${fabio.id}/reactivate`,
  },
  {
    who: hana,
    action: "Retirar acesso administrativo",
    title: "Retirar o acesso administrativo de Hana Tesoureira?",
    consequence: "A pessoa fica só com o papel Associado e as sessões dela serão encerradas.",
    confirm: "Retirar acesso",
    destructive: true,
    path: `/api/v1/users/${hana.id}/admin-membership/revoke`,
    reason: "Deixou a tesouraria",
  },
];

describe("confirmações (DSP-03 AC2, AC3 e AC5)", () => {
  it.each(CONFIRMATIONS.map((c) => ({ ...c, name: c.who.name })))(
    "$action $name: nomeia a pessoa, diz a consequência e o botão repete o verbo",
    async ({ who, action, title, consequence, confirm, destructive }) => {
      await renderPage([who]);
      await chooseAction(who.name, action);
      const dialog = await screen.findByRole("alertdialog", { name: title });

      expect(within(dialog).getByText(`${who.name} · ${who.email}`).tagName).toBe("STRONG");
      expect(within(dialog).getByText(consequence)).toBeTruthy();
      const button = within(dialog).getByRole("button", { name: confirm });
      expect(button.className.includes("bg-destructive")).toBe(destructive);
      expect(within(dialog).queryByRole("button", { name: /^(Confirmar|OK)$/ })).toBeNull();
    },
  );

  it.each(CONFIRMATIONS.map((c) => ({ ...c, name: c.who.name })))(
    "$action $name: o erro da API aparece dentro da confirmação, que continua aberta",
    async ({ who, action, title, confirm, path, reason }) => {
      mockWrite("post", path, () => problem(409, "last_admin"));
      await renderPage([who]);
      await chooseAction(who.name, action);
      const dialog = await screen.findByRole("alertdialog", { name: title });
      if (reason) type("Motivo", reason);

      fireEvent.click(within(dialog).getByRole("button", { name: confirm }));

      const message = "Não é possível: o sistema ficaria sem ninguém que possa conceder acesso.";
      expect(await within(dialog).findByText(message)).toBeTruthy();
      expect(screen.getByRole("alertdialog", { name: title })).toBe(dialog);
    },
  );
});

describe("formulários (DSP-04 AC1, AC4 e AC7)", () => {
  it("Novo usuário: formulário em FormLayout, e-mail sem correção ortográfica, Cancelar secundário", async () => {
    await renderPage([eva]);
    fireEvent.click(screen.getByRole("button", { name: "Novo usuário" }));
    const dialog = await screen.findByRole("dialog", { name: "Novo usuário" });

    expect(dialog.querySelector("form[data-slot=form-layout]")).not.toBeNull();
    expect(within(dialog).getByLabelText("E-mail").getAttribute("spellcheck")).toBe("false");
    const footer = dialog.querySelector("[data-slot=form-footer]") as HTMLElement;
    expect(within(footer).getAllByRole("button").map((b) => b.textContent)).toEqual(["Cancelar", "Criar usuário"]);
  });

  it("erro da API sem campo vai para o alerta no topo do formulário", async () => {
    mockWrite("post", "/api/v1/users", () => problem(503, "service_unavailable"));
    await renderPage([eva]);
    fireEvent.click(screen.getByRole("button", { name: "Novo usuário" }));
    const dialog = await screen.findByRole("dialog", { name: "Novo usuário" });
    type("Nome", "Daniela Nova");
    type("E-mail", "daniela@example.com");
    type("Senha inicial", "senha-inicial-123");
    fireEvent.click(within(dialog).getByRole("button", { name: "Criar usuário" }));

    const alert = await within(dialog).findByRole("alert");
    expect(alert.textContent).toContain("O serviço está indisponível no momento. Tente de novo em instantes.");
    const form = dialog.querySelector("form[data-slot=form-layout]") as HTMLElement;
    expect(form.firstElementChild).toBe(alert);
  });

  it.each([
    { who: eva, action: "Promover a administrativo", title: "Promover Eva Ativa a administrativo", submit: "Promover" },
    { who: hana, action: "Editar papéis", title: "Editar papéis de Hana Tesoureira", submit: "Salvar papéis" },
    { who: eva, action: "Gerar senha temporária", title: "Senha temporária para Eva Ativa", submit: "Gerar senha temporária" },
  ])("$title: FormLayout com Cancelar e a ação principal por último", async ({ who, action, title, submit }) => {
    await renderPage([who]);
    await chooseAction(who.name, action);
    const dialog = await screen.findByRole("dialog", { name: title });
    const footer = dialog.querySelector("form[data-slot=form-layout] [data-slot=form-footer]") as HTMLElement;
    expect(within(footer).getAllByRole("button").map((b) => b.textContent)).toEqual(["Cancelar", submit]);
  });
});
