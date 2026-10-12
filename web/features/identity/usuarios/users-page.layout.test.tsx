import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { mockViewport } from "@/test/media";
import { authContext, ids, user } from "@/test/msw/fixtures";

import { menuItemName, mockListUsers } from "./test-helpers";
import { UsersPage } from "./users-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => {
  resetNavigation();
  setLocation("/admin/usuarios");
});

// DSM-02 AC1 a AC3 e DSP-05 AC1 a AC4, AC7 e AC10 na tela de usuários: filtros
// na URL (ainda enviados à API como active e role), estados de lista e lista
// empilhada em tela estreita com as ações de cada pessoa.

const READ = "identity:user:read";
const WRITES = [
  "identity:user:update",
  "identity:admin:grant",
  "identity:role:assign",
  "identity:admin:revoke",
  "identity:user:reset_password",
];

const MEMBERSHIP = { reason: "Assumiu a tesouraria", granted_at: "2026-10-01T12:00:00Z" };

const eu = user({
  id: ids.user,
  name: "Pessoa Teste",
  email: "pessoa.teste@example.com",
  roles: ["ASSOCIADO", "PRESIDENTE"],
  admin_membership: MEMBERSHIP,
});
const ana = user({
  id: "00000000-0000-4000-8000-0000000000a1",
  name: "Ana Tesoureira",
  email: "ana@example.com",
  roles: ["ASSOCIADO", "TESOURARIA"],
  admin_membership: MEMBERSHIP,
});
const bruno = user({
  id: "00000000-0000-4000-8000-0000000000b2",
  name: "Bruno Associado",
  email: "bruno@example.com",
  active: false,
});

function renderPage(permissions: readonly string[] = [READ]) {
  mockMe(authContext({ permissions: [...permissions] }));
  renderWithSession(<UsersPage />);
}

const params = () => Object.fromEntries(new URLSearchParams(window.location.search));
const select = (label: string) => screen.getByLabelText(label) as HTMLSelectElement;

describe("filtros na URL (DSM-02 AC2, DSP-05 AC2 a AC4)", () => {
  it("mudar papel e situação escreve papel e situacao na URL e consulta a API com role e active", async () => {
    const requests = mockListUsers(() => ({ items: [ana, bruno], next_cursor: null }));
    renderPage();
    await screen.findByText("ana@example.com");

    fireEvent.change(select("Papel"), { target: { value: "TESOURARIA" } });
    await waitFor(() => expect(params()).toEqual({ papel: "TESOURARIA" }));
    fireEvent.change(select("Situação"), { target: { value: "inativos" } });
    await waitFor(() => expect(params()).toEqual({ papel: "TESOURARIA", situacao: "inativos" }));

    await waitFor(() => expect(requests).toHaveLength(3));
    expect(requests).toEqual([
      { limit: "50" },
      { limit: "50", role: "TESOURARIA" },
      { limit: "50", active: "false", role: "TESOURARIA" },
    ]);
    expect(window.location.pathname).toBe("/admin/usuarios");
  });

  it("voltar a Todos tira o parâmetro da URL e o filtro da consulta", async () => {
    setLocation("/admin/usuarios?situacao=ativos");
    const requests = mockListUsers(() => ({ items: [ana], next_cursor: null }));
    renderPage();
    await screen.findByText("ana@example.com");

    fireEvent.change(select("Situação"), { target: { value: "" } });

    await waitFor(() => expect(requests).toHaveLength(2));
    expect(window.location.search).toBe("");
    expect(requests).toEqual([{ limit: "50", active: "true" }, { limit: "50" }]);
  });

  it("abrir a URL com filtros restaura os campos e a consulta", async () => {
    setLocation("/admin/usuarios?situacao=inativos&papel=TESOURARIA");
    const requests = mockListUsers(() => ({ items: [bruno], next_cursor: null }));
    renderPage();
    await screen.findByText("bruno@example.com");

    expect(select("Situação").value).toBe("inativos");
    expect(select("Papel").value).toBe("TESOURARIA");
    expect(requests).toEqual([{ limit: "50", active: "false", role: "TESOURARIA" }]);
  });

  it("valor inválido na URL é ignorado: vale Todos, sem erro e sem filtro na consulta", async () => {
    setLocation("/admin/usuarios?situacao=talvez&papel=REI");
    const requests = mockListUsers(() => ({ items: [ana], next_cursor: null }));
    renderPage();
    await screen.findByText("ana@example.com");

    expect(select("Situação").value).toBe("");
    expect(select("Papel").value).toBe("");
    expect(requests).toEqual([{ limit: "50" }]);
    expect(screen.queryByRole("button", { name: "Limpar filtros" })).toBeNull();
  });
});

describe("barra de filtros e estados (DSP-05 AC1, DSP-02 AC2 e AC3)", () => {
  it("mostra a contagem de usuários carregados e Limpar filtros só com filtro ativo", async () => {
    mockListUsers((query) =>
      query.get("role") === "TESOURARIA"
        ? { items: [ana], next_cursor: null }
        : { items: [ana, bruno], next_cursor: null },
    );
    renderPage();
    await screen.findByText("bruno@example.com");
    expect(screen.getByText("2 usuários")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Limpar filtros" })).toBeNull();

    fireEvent.change(select("Papel"), { target: { value: "TESOURARIA" } });
    expect(await screen.findByText("1 usuário")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "Limpar filtros" }));
    await waitFor(() => expect(window.location.search).toBe(""));
    expect(await screen.findByText("2 usuários")).toBeTruthy();
    expect(select("Papel").value).toBe("");
  });

  it("filtro sem resultado mostra o estado de sem resultado, cuja ação limpa os filtros", async () => {
    setLocation("/admin/usuarios?papel=EVENTOS");
    const requests = mockListUsers((query) =>
      query.get("role") === "EVENTOS" ? { items: [], next_cursor: null } : { items: [ana], next_cursor: null },
    );
    renderPage();

    const heading = await screen.findByRole("heading", { name: "Nenhum resultado para os filtros" });
    expect(screen.queryByText("Nenhum usuário encontrado.")).toBeNull();
    const empty = heading.closest("[data-slot=empty-state]") as HTMLElement;
    fireEvent.click(within(empty).getByRole("button", { name: "Limpar filtros" }));

    expect(await screen.findByText("ana@example.com")).toBeTruthy();
    expect(window.location.search).toBe("");
    expect(requests).toEqual([{ limit: "50", role: "EVENTOS" }, { limit: "50" }]);
  });

  it("sem nenhum usuário e sem filtro, o estado vazio orienta e não oferece Limpar filtros", async () => {
    mockListUsers(() => ({ items: [], next_cursor: null }));
    renderPage();
    expect(await screen.findByRole("heading", { name: "Nenhum usuário encontrado." })).toBeTruthy();
    expect(screen.getByText(/As pessoas cadastradas aparecem aqui/)).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Limpar filtros" })).toBeNull();
  });

  it("a página tem um único h1, o do cabeçalho", async () => {
    mockListUsers(() => ({ items: [ana], next_cursor: null }));
    renderPage();
    await screen.findByText("ana@example.com");
    expect(screen.getAllByRole("heading", { level: 1 }).map((h) => h.textContent)).toEqual(["Usuários"]);
  });
});

describe("tela estreita (DSP-05 AC7 e AC10, DSM-02 AC3)", () => {
  it("mostra cada pessoa como item empilhado, com nome, e-mail, papéis, situação e as ações no menu", async () => {
    mockViewport({ width: 390, coarse: true });
    mockListUsers(() => ({ items: [eu, ana, bruno], next_cursor: null }));
    renderPage([READ, ...WRITES]);

    const list = await screen.findByRole("list", { name: "Usuários" });
    expect(screen.queryByRole("table")).toBeNull();
    const items = within(list).getAllByRole("listitem");
    expect(items).toHaveLength(3);

    const anaItem = items[1];
    expect(within(anaItem).getByText("Ana Tesoureira")).toBeTruthy();
    expect(within(anaItem).getByText("ana@example.com")).toBeTruthy();
    expect(within(anaItem).getByText("Tesouraria")).toBeTruthy();
    expect(within(anaItem).getByText("Ativo")).toBeTruthy();
    expect(within(items[2]).getByText("Inativo")).toBeTruthy();

    // A própria linha continua sem ações.
    expect(within(items[0]).queryByRole("button", { name: /^Ações de / })).toBeNull();

    fireEvent.click(within(anaItem).getByRole("button", { name: "Ações de Ana Tesoureira" }));
    const menu = await screen.findByRole("menu");
    expect(within(menu).getAllByRole("menuitem").map((item) => item.getAttribute("aria-label"))).toEqual([
      menuItemName("Editar papéis", "Ana Tesoureira"),
      menuItemName("Retirar acesso administrativo", "Ana Tesoureira"),
      menuItemName("Gerar senha temporária", "Ana Tesoureira"),
      menuItemName("Desativar", "Ana Tesoureira"),
    ]);

    fireEvent.click(within(menu).getByRole("menuitem", { name: menuItemName("Desativar", "Ana Tesoureira") }));
    expect(await screen.findByRole("alertdialog", { name: "Desativar Ana Tesoureira?" })).toBeTruthy();
  });

  it("os filtros ficam na gaveta e continuam escrevendo na URL", async () => {
    mockViewport({ width: 390, coarse: true });
    const requests = mockListUsers(() => ({ items: [ana], next_cursor: null }));
    renderPage();
    await screen.findByRole("list", { name: "Usuários" });

    fireEvent.click(screen.getByRole("button", { name: "Filtros" }));
    fireEvent.change(await screen.findByLabelText("Papel"), { target: { value: "TESOURARIA" } });

    await waitFor(() => expect(params()).toEqual({ papel: "TESOURARIA" }));
    await waitFor(() => expect(requests).toEqual([{ limit: "50" }, { limit: "50", role: "TESOURARIA" }]));
  });
});
