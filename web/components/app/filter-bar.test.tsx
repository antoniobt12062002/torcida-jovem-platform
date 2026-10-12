import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { mockViewport } from "@/test/media";

import { FilterBar, type FilterBarProps } from "./filter-bar";

function Barra(props: Partial<FilterBarProps>) {
  return (
    <FilterBar
      activeCount={0}
      resultCount={12}
      resultNoun={["lançamento", "lançamentos"]}
      onClear={() => {}}
      search={<input aria-label="Busca" />}
      {...props}
    >
      <select aria-label="Tipo">
        <option>Todos</option>
      </select>
      <select aria-label="Status">
        <option>Todos</option>
      </select>
    </FilterBar>
  );
}

function classes(el: Element) {
  return el.getAttribute("class")?.split(/\s+/) ?? [];
}

// DSP-05 AC1: sem `matchMedia` (padrão dos testes) vale o desktop.
describe("barra em tela larga", () => {
  it("mostra a busca e os filtros na própria barra, numa região nomeada", () => {
    render(<Barra />);
    const bar = screen.getByRole("group", { name: "Filtros" });
    expect(within(bar).getByRole("textbox", { name: "Busca" })).toBeTruthy();
    expect(within(bar).getByRole("combobox", { name: "Tipo" })).toBeTruthy();
    expect(within(bar).getByRole("combobox", { name: "Status" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: /^Filtros/ })).toBeNull();
  });

  it("sem filtro ativo, não mostra Limpar filtros", () => {
    render(<Barra activeCount={0} />);
    expect(screen.queryByRole("button", { name: "Limpar filtros" })).toBeNull();
  });

  it("com filtro ativo, mostra Limpar filtros, que chama onClear", () => {
    const onClear = vi.fn();
    render(<Barra activeCount={2} onClear={onClear} />);
    fireEvent.click(screen.getByRole("button", { name: "Limpar filtros" }));
    expect(onClear).toHaveBeenCalledTimes(1);
  });

  it("entre 768 e 1023 px a barra é uma grade de três colunas; a partir de 1024, uma linha", () => {
    render(<Barra />);
    expect(classes(screen.getByRole("group", { name: "Filtros" }))).toEqual(
      expect.arrayContaining(["bg-muted", "grid", "grid-cols-3", "lg:flex"]),
    );
  });

  it("sem busca, mostra só os filtros", () => {
    render(<Barra search={undefined} />);
    expect(screen.queryByRole("textbox")).toBeNull();
    expect(screen.getByRole("combobox", { name: "Tipo" })).toBeTruthy();
  });
});

// DSP-05 AC1: a contagem anuncia o efeito do filtro.
describe("contagem de resultados", () => {
  it.each([
    [0, "0 lançamentos"],
    [1, "1 lançamento"],
    [2, "2 lançamentos"],
    [12, "12 lançamentos"],
    [1234, "1.234 lançamentos"],
  ])("com %i resultados, mostra %s", (resultCount, text) => {
    render(<Barra resultCount={resultCount} />);
    const count = screen.getByText(text);
    expect(count.getAttribute("aria-live")).toBe("polite");
  });

  it("a região viva é a mesma quando a contagem muda", () => {
    const view = render(<Barra resultCount={12} />);
    const region = screen.getByText("12 lançamentos");
    view.rerender(<Barra resultCount={1} activeCount={1} />);
    expect(region.textContent).toBe("1 lançamento");
    expect(region.isConnected).toBe(true);
  });

  it("fica depois da barra, fora dela", () => {
    render(<Barra />);
    const bar = screen.getByRole("group", { name: "Filtros" });
    const count = screen.getByText("12 lançamentos");
    expect(bar.contains(count)).toBe(false);
    expect(bar.compareDocumentPosition(count) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it("continua visível em tela estreita", () => {
    mockViewport({ width: 390 });
    render(<Barra resultCount={1} />);
    expect(screen.getByText("1 lançamento").getAttribute("aria-live")).toBe("polite");
  });
});

// Abaixo de 768 px: busca visível, demais filtros na gaveta inferior.
describe("forma recolhida em tela estreita", () => {
  it("a busca fica visível e os demais filtros saem da tela", () => {
    mockViewport({ width: 390, coarse: true });
    render(<Barra />);
    expect(screen.getByRole("textbox", { name: "Busca" })).toBeTruthy();
    expect(screen.queryByRole("combobox")).toBeNull();
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("o botão mostra a quantidade de filtros ativos", () => {
    mockViewport({ width: 390 });
    const view = render(<Barra activeCount={2} />);
    expect(screen.getByRole("button", { name: "Filtros (2)" })).toBeTruthy();
    view.rerender(<Barra activeCount={1} />);
    expect(screen.getByRole("button", { name: "Filtros (1)" })).toBeTruthy();
  });

  it("sem filtro ativo, o botão não mostra quantidade", () => {
    mockViewport({ width: 390 });
    render(<Barra activeCount={0} />);
    expect(screen.getByRole("button", { name: "Filtros" })).toBeTruthy();
  });

  it("o botão abre uma gaveta inferior com os filtros", async () => {
    mockViewport({ width: 390 });
    render(<Barra activeCount={2} />);
    fireEvent.click(screen.getByRole("button", { name: "Filtros (2)" }));
    const sheet = await screen.findByRole("dialog", { name: "Filtros" });
    expect(sheet.getAttribute("data-side")).toBe("bottom");
    expect(within(sheet).getByRole("combobox", { name: "Tipo" })).toBeTruthy();
    expect(within(sheet).getByRole("combobox", { name: "Status" })).toBeTruthy();
    // A busca não é duplicada dentro da gaveta.
    expect(within(sheet).queryByRole("textbox", { name: "Busca" })).toBeNull();
  });

  it("Aplicar fecha a gaveta e devolve o foco ao botão", async () => {
    mockViewport({ width: 390 });
    render(<Barra activeCount={2} />);
    const trigger = screen.getByRole("button", { name: "Filtros (2)" });
    trigger.focus();
    fireEvent.click(trigger);
    const sheet = await screen.findByRole("dialog", { name: "Filtros" });
    fireEvent.click(within(sheet).getByRole("button", { name: "Aplicar" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
    await waitFor(() => expect(document.activeElement).toBe(trigger));
  });

  it("na gaveta, Limpar filtros só aparece com filtro ativo e chama onClear", async () => {
    mockViewport({ width: 390 });
    const onClear = vi.fn();
    const view = render(<Barra activeCount={1} onClear={onClear} />);
    fireEvent.click(screen.getByRole("button", { name: "Filtros (1)" }));
    const sheet = await screen.findByRole("dialog", { name: "Filtros" });
    fireEvent.click(within(sheet).getByRole("button", { name: "Limpar filtros" }));
    expect(onClear).toHaveBeenCalledTimes(1);

    view.rerender(<Barra activeCount={0} onClear={onClear} />);
    expect(within(sheet).queryByRole("button", { name: "Limpar filtros" })).toBeNull();
    expect(within(sheet).getByRole("button", { name: "Aplicar" })).toBeTruthy();
  });

  it("ao alargar a tela, os filtros voltam para a barra", () => {
    const viewport = mockViewport({ width: 390 });
    render(<Barra activeCount={1} />);
    expect(screen.queryByRole("combobox")).toBeNull();
    viewport.resize(768);
    expect(screen.getByRole("combobox", { name: "Tipo" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: /^Filtros/ })).toBeNull();
  });
});
