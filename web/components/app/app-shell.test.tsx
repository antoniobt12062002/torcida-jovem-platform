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

  // WDS-D-019: o ponto de quebra da navegação fixa passou de 768 (md) para
  // 1024 px (lg); abaixo dele vale o botão de menu.
  it("a navegação fixa e o botão de menu trocam em 1024 px (lg)", async () => {
    await renderShell(["financeiro.contas.read"]);
    const rail = screen.getByRole("navigation", { name: "Navegação principal" }).closest("[data-slot=shell-nav]");
    const railClasses = rail?.getAttribute("class")?.split(/\s+/) ?? [];
    expect(railClasses).toEqual(expect.arrayContaining(["hidden", "lg:block"]));
    expect(railClasses.filter((name) => name.startsWith("md:"))).toEqual([]);
    const buttonClasses = screen.getByRole("button", { name: "Abrir menu" }).className.split(/\s+/);
    expect(buttonClasses).toContain("lg:hidden");
    expect(buttonClasses.filter((name) => name.startsWith("md:"))).toEqual([]);
  });
});

// WDS-D-020: no shell a marca é "Torcida Jovem", com o escudo; "TJ Platform"
// fica só no título do documento.
describe("marca", () => {
  it("exibe o escudo e Torcida Jovem, com link para o início, e não exibe TJ Platform", async () => {
    const { container } = await renderShell([]);
    const brand = within(screen.getByRole("banner")).getByRole("link", { name: "Torcida Jovem" });
    expect(brand.getAttribute("href")).toBe("/inicio");
    expect(within(brand).getByRole("img", { name: "Torcida Jovem" })).toBeTruthy();
    expect(container.textContent).not.toContain("TJ Platform");
  });
});

// DSS-01 AC5.
describe("link de pular", () => {
  it("é o primeiro link da página e leva o foco ao conteúdo", async () => {
    await renderShell([]);
    const skip = screen.getAllByRole("link")[0];
    expect(skip.textContent).toBe("Pular para o conteúdo");
    fireEvent.click(skip);
    expect(document.activeElement?.textContent).toBe("conteúdo da tela");
  });
});

// FND-05 AC1: menu da pessoa.
describe("menu da pessoa", () => {
  // DSS-01 AC7: "Trocar senha" navega, então é um link para /conta/senha (antes
  // era um item com `onClick` e `router.push`); "Sair" é um botão.
  it("oferece Trocar senha, que é um link para /conta/senha, e Sair, que é um botão", async () => {
    await renderShell([]);
    const menu = await openAccountMenu();
    expect(within(menu).getAllByRole("menuitem").map((i) => i.textContent)).toEqual(["Trocar senha", "Sair"]);
    const changePassword = within(menu).getByRole("menuitem", { name: "Trocar senha" });
    expect(changePassword.tagName).toBe("A");
    expect(changePassword.getAttribute("href")).toBe("/conta/senha");
    expect(within(menu).getByRole("menuitem", { name: "Sair" }).tagName).toBe("BUTTON");
    // Abrir o menu não navega: quem navega é o link.
    expect(navigations()).toEqual([]);
  });

  it("mostra o nome completo e o e-mail da pessoa no topo do menu", async () => {
    await renderShell([]);
    const menu = await openAccountMenu();
    expect(within(menu).getByText("Pessoa Teste")).toBeTruthy();
    expect(within(menu).getByText("pessoa.teste@example.com")).toBeTruthy();
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
