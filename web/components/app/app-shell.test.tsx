import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { navigations, resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { AppShell } from "./app-shell";
import type { NavEntry } from "./nav";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const items: NavEntry[] = [
  { label: "Início", href: "/inicio" },
  { label: "Usuários", href: "/usuarios", permission: "identity.users.read" },
  {
    label: "Financeiro",
    items: [
      { label: "Contas", href: "/financeiro/contas", permission: "financeiro.contas.read" },
      { label: "Lançamentos", href: "/financeiro/lancamentos", permission: "financeiro.lancamentos.read" },
    ],
  },
];

async function renderShell(permissions: string[], mustChangePassword = false) {
  mockMe(authContext({ permissions, must_change_password: mustChangePassword }));
  const view = renderWithSession(
    <AppShell items={items}>
      <p>conteúdo da tela</p>
    </AppShell>,
  );
  await screen.findByRole("button", { name: /Pessoa Teste/ });
  return view;
}

function linkNames(container: HTMLElement) {
  return within(container)
    .queryAllByRole("link")
    .map((link) => [link.textContent, link.getAttribute("href")]);
}

async function openAccountMenu() {
  fireEvent.click(screen.getByRole("button", { name: /Pessoa Teste/ }));
  return screen.findByRole("menu");
}

// FND-05 AC1.
describe("navegação por permissão", () => {
  it("mostra o nome da pessoa, o conteúdo e só os itens com permissão (conjunto 1)", async () => {
    await renderShell(["financeiro.contas.read"]);
    expect(screen.getByText("conteúdo da tela")).toBeTruthy();
    const nav = screen.getByRole("navigation", { name: "Navegação principal" });
    expect(linkNames(nav)).toEqual([
      ["Início", "/inicio"],
      ["Contas", "/financeiro/contas"],
    ]);
    expect(within(nav).getByText("Financeiro")).toBeTruthy();
  });

  it("mostra só os itens com permissão (conjunto 2)", async () => {
    await renderShell(["identity.users.read", "financeiro.lancamentos.read"]);
    const nav = screen.getByRole("navigation", { name: "Navegação principal" });
    expect(linkNames(nav)).toEqual([
      ["Início", "/inicio"],
      ["Usuários", "/usuarios"],
      ["Lançamentos", "/financeiro/lancamentos"],
    ]);
  });

  it("marca o item da página atual", async () => {
    setLocation("/financeiro/contas/123");
    await renderShell(["financeiro.contas.read"]);
    const nav = screen.getByRole("navigation", { name: "Navegação principal" });
    expect(within(nav).getByRole("link", { name: "Contas" }).getAttribute("aria-current")).toBe("page");
    expect(within(nav).getByRole("link", { name: "Início" }).getAttribute("aria-current")).toBeNull();
  });
});

// FND-05 AC2.
describe("menu recolhível", () => {
  it("abre um menu com os mesmos itens visíveis e fecha ao escolher um item", async () => {
    await renderShell(["financeiro.contas.read"]);
    fireEvent.click(screen.getByRole("button", { name: "Abrir menu" }));
    const sheet = await screen.findByRole("dialog", { name: "Menu" });
    expect(linkNames(sheet)).toEqual([
      ["Início", "/inicio"],
      ["Contas", "/financeiro/contas"],
    ]);
    fireEvent.click(within(sheet).getByRole("link", { name: "Contas" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });
});

// FND-05 AC1: menu da pessoa.
describe("menu da pessoa", () => {
  it("oferece Trocar senha, que leva a /conta/senha", async () => {
    await renderShell([]);
    const menu = await openAccountMenu();
    expect(within(menu).getAllByRole("menuitem").map((i) => i.textContent)).toEqual(["Trocar senha", "Sair"]);
    fireEvent.click(within(menu).getByRole("menuitem", { name: "Trocar senha" }));
    await waitFor(() => expect(navigations()).toEqual(["/conta/senha"]));
  });

  it("Sair chama POST /auth/logout e leva a /entrar", async () => {
    let logoutCalls = 0;
    server.use(
      http.post(apiUrl("/api/v1/auth/logout"), () => {
        logoutCalls += 1;
        return new HttpResponse(null, { status: 204 });
      }),
    );
    await renderShell([]);
    const menu = await openAccountMenu();
    fireEvent.click(within(menu).getByRole("menuitem", { name: "Sair" }));
    await waitFor(() => expect(navigations()).toEqual(["/entrar"]));
    expect(logoutCalls).toBe(1);
  });
});

// FND-05 AC8.
describe("com troca de senha obrigatória", () => {
  it("esconde a navegação e mantém só Sair", async () => {
    setLocation("/conta/senha");
    await renderShell(["identity.users.read", "financeiro.contas.read"], true);
    expect(screen.getByText("conteúdo da tela")).toBeTruthy();
    expect(screen.queryByRole("navigation")).toBeNull();
    expect(screen.queryByRole("button", { name: "Abrir menu" })).toBeNull();
    expect(screen.queryAllByRole("link")).toEqual([]);
    const menu = await openAccountMenu();
    expect(within(menu).getAllByRole("menuitem").map((i) => i.textContent)).toEqual(["Sair"]);
  });
});
