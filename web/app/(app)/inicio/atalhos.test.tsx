import { screen, within } from "@testing-library/react";
import { Receipt } from "lucide-react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext } from "@/test/msw/fixtures";

import InicioPage from "./page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

// O módulo fornece ícone e descrição num item e não no outro: o atalho mostra
// o que houver e não quebra sem eles (DSS-06 AC2).
vi.mock("@/features/financeiro/nav", () => ({
  financeiroNav: {
    label: "Financeiro",
    items: [
      {
        label: "Lançamentos",
        href: "/financeiro/lancamentos",
        permission: "financeiro:lancamento:read",
        icon: Receipt,
        description: "Receitas e despesas da torcida.",
      },
      { label: "Contas", href: "/financeiro/contas", permission: "financeiro:conta:read" },
    ],
  },
}));

beforeEach(() => {
  resetNavigation();
  setLocation("/inicio");
});

async function renderInicio(permissions: string[]) {
  mockMe(authContext({ permissions }));
  renderWithSession(<InicioPage />);
  screen.getByRole("heading", { level: 1, name: "Início" });
  // A saudação só aparece com a sessão carregada.
  await screen.findByText("Olá, Pessoa.");
}

function cardOf(link: HTMLElement) {
  const card = link.closest<HTMLElement>("[data-slot='card']");
  expect(card).not.toBeNull();
  return card!;
}

describe("atalhos da página inicial (DSS-06)", () => {
  it("cada atalho é um link em cartão, com ícone, nome e descrição quando o módulo os fornece", async () => {
    await renderInicio(["financeiro:lancamento:read", "financeiro:conta:read"]);
    const regiao = screen.getByRole("navigation", { name: "Atalhos das áreas" });

    const lancamentos = within(regiao).getByRole("link", { name: "Lançamentos" });
    expect(lancamentos.getAttribute("href")).toBe("/financeiro/lancamentos");
    const card = cardOf(lancamentos);
    expect(card.querySelector("svg.lucide-receipt")).not.toBeNull();
    expect(within(card).getByText("Receitas e despesas da torcida.")).toBeTruthy();

    // Sem ícone nem descrição: só o nome, sem elemento vazio.
    const contas = cardOf(within(regiao).getByRole("link", { name: "Contas" }));
    expect(contas.querySelector("svg")).toBeNull();
    expect(contas.querySelector("p")).toBeNull();
  });

  it("respeita as permissões: área sem permissão não tem atalho", async () => {
    await renderInicio(["financeiro:conta:read"]);
    const regiao = screen.getByRole("navigation", { name: "Atalhos das áreas" });
    expect(within(regiao).queryByRole("link", { name: "Lançamentos" })).toBeNull();
    expect(within(regiao).getByRole("link", { name: "Contas" })).toBeTruthy();
  });

  it("os atalhos ficam em uma coluna abaixo de 768 px e se dividem a partir daí", async () => {
    await renderInicio(["financeiro:lancamento:read", "financeiro:conta:read"]);
    const grade = document.querySelector<HTMLElement>("[data-slot='atalhos-grade']");
    expect(grade).not.toBeNull();
    const classes = grade!.className.split(/\s+/);
    expect(classes).toContain("grid-cols-1");
    expect(classes).toContain("md:grid-cols-2");
    // Nenhuma divisão antes de 768 px.
    expect(classes.filter((name) => /^(?:sm:)?grid-cols-(?!1$)/.test(name))).toEqual([]);
  });

  it("cabeçalho com o título Início e a saudação pelo primeiro nome, sem indicadores", async () => {
    await renderInicio(["financeiro:conta:read"]);
    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    expect(screen.getByText("Olá, Pessoa.")).toBeTruthy();
    expect(screen.getAllByRole("main")).toHaveLength(1);
    expect(screen.queryByRole("table")).toBeNull();
    expect(screen.queryByRole("img")).toBeNull();
  });

  it("conta sem nenhuma área vê o estado vazio com a explicação", async () => {
    await renderInicio([]);
    expect(screen.queryByRole("navigation", { name: "Atalhos das áreas" })).toBeNull();
    const vazio = document.querySelector<HTMLElement>("[data-slot='empty-state']");
    expect(vazio).not.toBeNull();
    expect(within(vazio!).getByRole("heading", { name: "Sua conta ainda não tem acesso a nenhum módulo." })).toBeTruthy();
    expect(within(vazio!).getByText(/fale com a administração/)).toBeTruthy();
  });
});
