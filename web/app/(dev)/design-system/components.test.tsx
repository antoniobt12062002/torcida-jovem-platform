import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { mockViewport } from "@/test/media";

vi.mock("next/navigation", async () => ({
  ...(await import("@/lib/session/navigation-mock")).nextNavigationMock,
  notFound: vi.fn(),
}));

import { resetNavigation } from "@/lib/session/navigation-mock";

import DesignSystemPage from "./page";

// DSF-06 AC1 a AC4 e DSC-03 AC9: componentes, estados, padrões e shell do
// espécime, todos reais, com dados de exemplo e sem chamada à API.

beforeEach(() => resetNavigation());
afterEach(() => vi.restoreAllMocks());

function renderPage() {
  const view = render(<DesignSystemPage />);
  const section = (id: string) => view.container.querySelector(`section#${id}`) as HTMLElement;
  return { ...view, section };
}

describe("sem API e sem os componentes adiados", () => {
  it("renderizar o espécime não faz nenhuma requisição", () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    renderPage();
    expect(fetchSpy).not.toHaveBeenCalled();
  });

  it("Radio, Switch, Tabs e Tooltip não aparecem (nem o Select do kit, sem uso na V1)", () => {
    const { container } = renderPage();
    for (const role of ["radio", "radiogroup", "switch", "tab", "tablist", "tabpanel", "tooltip"]) {
      expect(screen.queryAllByRole(role), role).toEqual([]);
    }
    expect(container.querySelector('[data-slot^="radio"], [data-slot^="switch"], [data-slot^="tabs"], [data-slot^="tooltip"], [data-slot^="select"]')).toBeNull();
  });
});

describe("ações", () => {
  it("mostra cada variante nos três tamanhos, em repouso, desabilitada e em andamento", () => {
    const { section } = renderPage();
    const acoes = section("acoes");
    for (const label of ["Salvar lançamento", "Exportar", "Limpar filtros", "Excluir rascunho", "Ver histórico"]) {
      const buttons = within(acoes).getAllByRole("button", { name: label });
      expect(buttons, label).toHaveLength(6);
      expect(buttons.filter((button) => (button as HTMLButtonElement).disabled)).toHaveLength(3);
    }
    const busy = acoes.querySelectorAll('[aria-busy="true"]');
    expect(busy).toHaveLength(15);
    for (const button of busy) expect(button.textContent).toBe("Salvando…");
  });

  it("mostra IconButton com nome acessível e o link com aparência de botão", () => {
    const { section } = renderPage();
    const acoes = section("acoes");
    expect(within(acoes).getAllByRole("button", { name: /^Mais ações/ })).toHaveLength(6);
    expect(within(acoes).getAllByRole("button", { name: "Buscar (desabilitado)" })).toHaveLength(2);
    const link = within(acoes).getByRole("link", { name: "Voltar para a lista" });
    expect(link.getAttribute("data-slot")).toBeNull();
    expect(link.className).toContain("h-control-md");
  });
});

describe("campos", () => {
  it("mostra cada controle em repouso, inválido com mensagem, desabilitado, somente leitura e opcional", () => {
    const { section } = renderPage();
    const campos = section("campos");
    expect(within(campos).getByRole("textbox", { name: "Descrição" })).toBeTruthy();
    expect(within(campos).getByText("Informe um e-mail válido.")).toBeTruthy();
    expect(within(campos).getByText("Escolha a conta.")).toBeTruthy();
    expect(within(campos).getByText("É preciso confirmar.")).toBeTruthy();
    expect(within(campos).getByText("Informe o valor.")).toBeTruthy();
    expect(within(campos).getByText("Informe um número diferente de zero.")).toBeTruthy();
    expect((within(campos).getByRole("textbox", { name: "Código" }) as HTMLInputElement).readOnly).toBe(true);
    expect(within(campos).getAllByText("(opcional)").length).toBeGreaterThanOrEqual(3);
    expect(within(campos).getAllByRole("combobox").length).toBeGreaterThanOrEqual(4);
    expect(within(campos).getAllByRole("checkbox")).toHaveLength(4);
    expect(campos.querySelectorAll("input:disabled, textarea:disabled, select:disabled").length).toBeGreaterThanOrEqual(5);
    expect(within(campos).getAllByText("R$").length).toBeGreaterThanOrEqual(3);
  });
});

describe("status e valores", () => {
  it("mostra os status de lançamento, ativo, inativo e saldo negativo, cada um com ícone e rótulo", () => {
    const { section } = renderPage();
    const status = section("status-e-valores");
    for (const label of ["Criada", "Recebida", "Paga", "Cancelada", "Ativo", "Inativo", "Saldo negativo"]) {
      const badge = within(status).getByText(label, { selector: "[data-slot=badge], [data-tone]" });
      expect(badge.querySelector("svg"), label).not.toBeNull();
    }
    const cancelada = within(status).getByText("Cancelada", { selector: "[data-tone]" });
    expect(cancelada.getAttribute("data-tone")).toBe("neutral");
  });

  it("mostra o MoneyValue nas quatro variantes e o RoleBadge", () => {
    const { section } = renderPage();
    const status = section("status-e-valores");
    const variants = [...status.querySelectorAll("[data-slot=money-value]")].map((node) => node.getAttribute("data-variant"));
    expect(new Set(variants)).toEqual(new Set(["neutral", "income", "expense", "balance"]));
    expect(within(status).getAllByText("saldo negativo").length).toBeGreaterThanOrEqual(3);
    expect(within(status).getByText("Tesouraria")).toBeTruthy();
  });

  it("mostra o saldo em Barlow Condensed e em Geist lado a lado, com o mesmo componente (DD-4)", () => {
    const { section } = renderPage();
    const status = section("status-e-valores");
    const display = [...status.querySelectorAll<HTMLElement>("[data-slot=money-value]")].filter((node) =>
      node.className.includes("text-4xl"),
    );
    const barlow = display.filter((node) => node.className.includes("font-money-display"));
    const geist = display.filter((node) => node.className.includes("font-sans"));
    expect(barlow.length).toBe(5);
    expect(geist.length).toBe(5);
    expect(geist.every((node) => !node.className.includes("font-money-display"))).toBe(true);
    expect(barlow.map((node) => node.textContent)).toEqual(geist.map((node) => node.textContent));
  });
});

describe("tabela e lista", () => {
  it("mostra a listagem como tabela, forçada em tabela e em lista, e a tabela pequena", () => {
    const { section } = renderPage();
    const listagem = section("tabela-e-lista");
    const auto = within(listagem).getByRole("table", { name: "Lançamentos de exemplo (automático)" });
    expect(within(auto).getAllByRole("row")).toHaveLength(9);
    const table = within(listagem).getByRole("table", { name: "Lançamentos de exemplo (tabela)" });
    expect(within(table).getByRole("columnheader", { name: "Conta" }).className).toContain("hidden lg:table-cell");
    const list = within(listagem).getByRole("list", { name: "Lançamentos de exemplo (lista)" });
    expect(within(list).getAllByRole("listitem")).toHaveLength(8);
    expect(within(listagem).getByRole("table", { name: "Comprovantes de exemplo" })).toBeTruthy();
    expect(within(listagem).getAllByRole("button", { name: "Detalhes lançamento Mensalidades de março" })).toHaveLength(3);
  });

  it("abaixo de 768 px a listagem automática vira lista e a tabela pequena continua tabela", () => {
    mockViewport({ width: 390, coarse: true });
    const { section } = renderPage();
    const listagem = section("tabela-e-lista");
    expect(within(listagem).getByRole("list", { name: "Lançamentos de exemplo (automático)" })).toBeTruthy();
    expect(within(listagem).queryByRole("table", { name: "Lançamentos de exemplo (automático)" })).toBeNull();
    expect(within(listagem).getByRole("table", { name: "Comprovantes de exemplo" })).toBeTruthy();
    expect(within(listagem).getByRole("table", { name: "Lançamentos de exemplo (tabela)" })).toBeTruthy();
  });

  it("Mostrar mais acrescenta as linhas restantes", () => {
    const { section } = renderPage();
    const listagem = section("tabela-e-lista");
    expect(within(listagem).getByText("Mostrando 8 de 12 lançamentos")).toBeTruthy();
    fireEvent.click(within(listagem).getByRole("button", { name: "Mostrar mais" }));
    const auto = within(listagem).getByRole("table", { name: "Lançamentos de exemplo (automático)" });
    expect(within(auto).getAllByRole("row")).toHaveLength(13);
    expect(within(listagem).queryByRole("button", { name: "Mostrar mais" })).toBeNull();
  });
});

describe("estados de tela", () => {
  it("mostra carregando nas quatro formas, vazio, sem resultado, erro e o acesso ao sem acesso", () => {
    const { section } = renderPage();
    const estados = section("estados-de-tela");
    expect(within(estados).getAllByRole("status")).toHaveLength(4);
    expect(within(estados).getByRole("heading", { level: 3, name: "Nenhum lançamento ainda" })).toBeTruthy();
    expect(within(estados).getByRole("heading", { level: 3, name: "Nenhum resultado para os filtros" })).toBeTruthy();
    expect(within(estados).getByRole("button", { name: "Limpar filtros" })).toBeTruthy();
    expect(within(estados).getByRole("alert").textContent).toContain("Não foi possível carregar os dados.");
    expect(within(estados).getByRole("button", { name: "Tentar de novo" })).toBeTruthy();
    expect(within(estados).getByRole("link", { name: "Abrir o estado sem acesso" }).getAttribute("href")).toBe(
      "/design-system/sem-acesso",
    );
  });
});

describe("feedback", () => {
  it("mostra alertas nos quatro tons e um botão para cada toast", () => {
    const { section } = renderPage();
    const feedback = section("feedback");
    expect(feedback.querySelectorAll("[data-slot=alert]")).toHaveLength(4);
    for (const name of ["Toast neutro", "Toast de sucesso", "Toast de aviso", "Toast de erro"]) {
      expect(within(feedback).getByRole("button", { name })).toBeTruthy();
    }
  });

  it("a confirmação pede motivo obrigatório e nomeia o registro e a consequência", async () => {
    const { section } = renderPage();
    fireEvent.click(within(section("feedback")).getByRole("button", { name: "Cancelar lançamento" }));
    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText("Receita fb6e3d8f · Mensalidades de março")).toBeTruthy();
    expect(within(dialog).getByText(/sai do saldo/)).toBeTruthy();
    const confirm = within(dialog).getByRole("button", { name: "Cancelar lançamento" }) as HTMLButtonElement;
    expect(confirm.disabled).toBe(true);
    fireEvent.change(within(dialog).getByLabelText("Motivo"), { target: { value: "Duplicado" } });
    expect(confirm.disabled).toBe(false);
  });
});

describe("sobreposições", () => {
  it("o diálogo abre com o formulário em FormLayout e acusa os campos ao salvar vazio", async () => {
    const { section } = renderPage();
    fireEvent.click(within(section("sobreposicoes")).getByRole("button", { name: "Novo lançamento" }));
    const dialog = await screen.findByRole("dialog", { name: "Novo lançamento" });
    expect(dialog.querySelector("[data-slot=form-layout]")).not.toBeNull();
    fireEvent.click(within(dialog).getByRole("button", { name: "Salvar" }));
    expect(await within(dialog).findByText("Informe a descrição.")).toBeTruthy();
    expect(within(dialog).getByText("Corrija os campos destacados.")).toBeTruthy();
  });

  it("a gaveta inferior e o menu de ações de linha abrem", async () => {
    const { section } = renderPage();
    const sobreposicoes = section("sobreposicoes");
    fireEvent.click(within(sobreposicoes).getByRole("button", { name: "Abrir gaveta" }));
    const sheet = await screen.findByRole("dialog", { name: "Resumo do mês" });
    expect(sheet.getAttribute("data-side")).toBe("bottom");
    fireEvent.keyDown(sheet, { key: "Escape" });
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Resumo do mês" })).toBeNull());
    fireEvent.click(within(sobreposicoes).getByRole("button", { name: "Ações de lançamento Venda de bonés" }));
    const menu = await screen.findByRole("menu");
    expect(within(menu).getAllByRole("menuitem").map((item) => item.textContent)).toEqual([
      "Editar",
      "Anexar comprovante",
      "Cancelar lançamento",
    ]);
  });
});

describe("shell", () => {
  it("mostra navegação com ícones, item ativo e menu da pessoa, sem criar main nem h1", async () => {
    const { section } = renderPage();
    const shell = section("shell");
    const nav = within(shell).getByRole("navigation", { name: "Navegação principal" });
    const active = within(nav).getByRole("link", { name: "Lançamentos" });
    expect(active.getAttribute("aria-current")).toBe("page");
    for (const label of ["Início", "Lançamentos", "Contas", "Saldo", "Produtos", "Usuários"]) {
      expect(within(nav).getByRole("link", { name: label }).querySelector("svg"), label).not.toBeNull();
    }
    expect(shell.querySelector("main, h1")).toBeNull();
    fireEvent.click(within(shell).getByRole("button", { name: "Pessoa de Exemplo" }));
    const menu = await screen.findByRole("menu");
    expect(within(menu).getByText("pessoa@exemplo.com")).toBeTruthy();
    expect(within(menu).getAllByRole("menuitem").map((item) => item.textContent)).toEqual(["Trocar senha", "Sair"]);
  });

  it("acesso e página de exemplo completa abrem em páginas próprias do espécime", () => {
    const { section } = renderPage();
    expect(within(section("acesso")).getByRole("link", { name: "Abrir a tela de acesso de exemplo" }).getAttribute("href")).toBe(
      "/design-system/acesso",
    );
    expect(
      within(section("pagina-de-exemplo")).getByRole("link", { name: "Abrir a página de exemplo completa" }).getAttribute("href"),
    ).toBe("/design-system/pagina-de-exemplo");
  });
});

describe("página de exemplo", () => {
  it("a busca e os filtros filtram a listagem de exemplo e ficam na URL", async () => {
    const { section } = renderPage();
    const exemplo = section("pagina-de-exemplo");
    expect(within(exemplo).getByText("12 lançamentos")).toBeTruthy();
    fireEvent.change(within(exemplo).getByRole("combobox", { name: "Tipo" }), { target: { value: "DESPESA" } });
    expect(within(exemplo).getByText("6 lançamentos")).toBeTruthy();
    expect(window.location.search).toContain("tipo=DESPESA");
    fireEvent.change(within(exemplo).getByRole("searchbox", { name: "Buscar" }), { target: { value: "onibus" } });
    expect(within(exemplo).getByText("1 lançamento")).toBeTruthy();
    fireEvent.change(within(exemplo).getByRole("combobox", { name: "Status" }), { target: { value: "PAGA" } });
    expect(within(exemplo).getByRole("heading", { level: 3, name: "Nenhum resultado para os filtros" })).toBeTruthy();
    act(() => {
      fireEvent.click(within(exemplo).getAllByRole("button", { name: "Limpar filtros" })[0]);
    });
    expect(within(exemplo).getByText("12 lançamentos")).toBeTruthy();
  });
});
