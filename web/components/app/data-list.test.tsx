import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { hydrateRoot } from "react-dom/client";
import { renderToString } from "react-dom/server";
import { afterEach, describe, expect, it, vi } from "vitest";

import { mockViewport } from "@/test/media";

import { type Column, DataList, type DataListProps } from "./data-list";
import { RowActionButton, RowActions } from "./row-actions";

type Lancamento = {
  id: string;
  tipo: string;
  descricao: string;
  conta: string;
  data: string;
  responsavel: string;
  status: string;
  valor: string;
};

const rows: Lancamento[] = [
  {
    id: "fb6e3d8f",
    tipo: "Receita",
    descricao: "Bilheteria do jogo de sábado",
    conta: "Bilheteria",
    data: "09/10/2026",
    responsavel: "Ana Souza",
    status: "Recebida",
    valor: "R$ 198,50",
  },
  {
    id: "0a1b2c3d",
    tipo: "Despesa",
    descricao: "Faixas novas",
    conta: "Materiais",
    data: "10/10/2026",
    responsavel: "Bruno Lima",
    status: "Criada",
    valor: "R$ 10,50",
  },
];

const columns: Column<Lancamento>[] = [
  { id: "descricao", header: "Descrição", priority: 1, inList: "title", cell: (row) => row.descricao },
  { id: "status", header: "Status", priority: 1, inList: "meta", cell: (row) => <span>{row.status}</span> },
  { id: "conta", header: "Conta", priority: 2, inList: "field", wrap: "truncate", cell: (row) => row.conta },
  { id: "data", header: "Data", priority: 2, inList: "field", wrap: "nowrap", cell: (row) => row.data },
  { id: "tipo", header: "Tipo", priority: 2, inList: "field", cell: (row) => row.tipo },
  { id: "id", header: "Código", priority: 3, inList: "field", cell: (row) => row.id },
  { id: "responsavel", header: "Responsável", priority: 3, inList: false, cell: (row) => row.responsavel },
  { id: "valor", header: "Valor", priority: 1, numeric: true, cell: (row) => row.valor },
];

const rowLabel = (row: Lancamento) => `lançamento ${row.tipo} ${row.id}`;

function Lista(props: Partial<DataListProps<Lancamento>>) {
  return (
    <DataList
      label="Lançamentos"
      columns={columns}
      rows={rows}
      rowKey={(row) => row.id}
      rowLabel={rowLabel}
      primaryAction={(row) => <RowActionButton label="Detalhes" rowLabel={rowLabel(row)} href={`/l/${row.id}`} />}
      moreActions={() => [{ label: "Editar" }, { label: "Cancelar lançamento", destructive: true }]}
      {...props}
    />
  );
}

function classes(el: Element) {
  return el.getAttribute("class")?.split(/\s+/) ?? [];
}

afterEach(() => vi.restoreAllMocks());

// DSP-05 AC5, DSC-05 AC1: sem `matchMedia` (padrão dos testes) vale o desktop.
describe("tabela", () => {
  it("renderiza uma tabela com nome acessível e um cabeçalho de coluna por coluna", () => {
    render(<Lista />);
    const table = screen.getByRole("table", { name: "Lançamentos" });
    const headers = within(table).getAllByRole("columnheader");
    expect(headers.map((th) => th.textContent)).toEqual([
      "Descrição",
      "Status",
      "Conta",
      "Data",
      "Tipo",
      "Código",
      "Responsável",
      "Valor",
      "Ações",
    ]);
    for (const th of headers) expect(th.getAttribute("scope")).toBe("col");
    expect(screen.queryByRole("list", { name: "Lançamentos" })).toBeNull();
  });

  it("mostra uma linha por registro, com o conteúdo de cada coluna", () => {
    render(<Lista />);
    const [, first, second] = screen.getAllByRole("row");
    expect(within(first).getAllByRole("cell").map((td) => td.textContent).slice(0, 8)).toEqual([
      "Bilheteria do jogo de sábado",
      "Recebida",
      "Bilheteria",
      "09/10/2026",
      "Receita",
      "fb6e3d8f",
      "Ana Souza",
      "R$ 198,50",
    ]);
    expect(within(second).getByText("Faixas novas")).toBeTruthy();
    expect(screen.getAllByRole("row")).toHaveLength(3);
  });

  it("não cria contêiner de rolagem horizontal", () => {
    const { container } = render(<Lista />);
    expect(container.querySelector("[data-slot=table-container]")).toBeNull();
  });

  it("a coluna numérica fica à direita com algarismos tabulares, no cabeçalho e nas células", () => {
    render(<Lista />);
    const header = screen.getByRole("columnheader", { name: "Valor" });
    expect(classes(header)).toEqual(expect.arrayContaining(["text-right", "tabular-nums"]));
    const cell = screen.getByText("R$ 198,50");
    expect(classes(cell)).toEqual(expect.arrayContaining(["text-right", "tabular-nums"]));
    expect(classes(screen.getByRole("columnheader", { name: "Conta" }))).not.toContain("text-right");
    // O valor não quebra no meio.
    expect(classes(cell)).toContain("whitespace-nowrap");
  });

  it("a coluna declara o tratamento de texto longo", () => {
    render(<Lista />);
    expect(classes(screen.getByText("Bilheteria"))).toContain("truncate");
    expect(classes(screen.getByText("09/10/2026"))).toContain("whitespace-nowrap");
  });

  // DSP-05 AC6: prioridade só por CSS, na mesma tabela.
  it("coluna de prioridade 3 fica oculta abaixo de 1024 px, no cabeçalho e nas células", () => {
    render(<Lista />);
    for (const el of [
      screen.getByRole("columnheader", { name: "Código" }),
      screen.getByRole("columnheader", { name: "Responsável" }),
      screen.getByText("fb6e3d8f"),
      screen.getByText("Ana Souza"),
    ]) {
      expect(classes(el)).toEqual(expect.arrayContaining(["hidden", "lg:table-cell"]));
    }
  });

  // A prioridade 3 depende também da largura do contêiner, não só da tela: ao
  // lado da navegação fixa, em 1024 px, sobram cerca de 720 px para a tabela.
  it("coluna de prioridade 3 também fica oculta quando o contêiner tem menos de 56rem", () => {
    render(<Lista />);
    for (const el of [
      screen.getByRole("columnheader", { name: "Código" }),
      screen.getByRole("columnheader", { name: "Responsável" }),
      screen.getByText("fb6e3d8f"),
      screen.getByText("Ana Souza"),
    ]) {
      expect(classes(el)).toContain("@max-4xl:hidden");
    }
  });

  it("a tabela fica num contêiner de consulta que encolhe com a área de conteúdo", () => {
    render(<Lista />);
    const wrapper = screen.getByRole("table", { name: "Lançamentos" }).parentElement!;
    expect(wrapper.getAttribute("data-slot")).toBe("data-list-table");
    expect(classes(wrapper)).toEqual(expect.arrayContaining(["@container", "w-full", "min-w-0"]));
  });

  it("no layout automático, a tabela que não couber rola no próprio contêiner, nunca no documento", () => {
    render(<Lista />);
    const wrapper = screen.getByRole("table", { name: "Lançamentos" }).parentElement!;
    expect(classes(wrapper)).toContain("overflow-x-auto");
    // Não é uma região rolável nomeada: a rolagem é só uma salvaguarda.
    expect(wrapper.getAttribute("role")).toBeNull();
    expect(wrapper.getAttribute("tabindex")).toBeNull();
  });

  it("na tabela forçada, a rolagem fica com a região nomeada de quem usa", () => {
    render(<Lista layout="table" />);
    const wrapper = screen.getByRole("table", { name: "Lançamentos" }).parentElement!;
    expect(classes(wrapper)).toEqual(expect.arrayContaining(["@container", "min-w-0"]));
    expect(classes(wrapper)).not.toContain("overflow-x-auto");
  });

  it("a coluna de ações nunca recebe classe de ocultação", () => {
    render(<Lista />);
    const header = screen.getByRole("columnheader", { name: "Ações" });
    expect(classes(header).filter((c) => c === "hidden" || c.endsWith(":hidden"))).toEqual([]);
    const button = screen.getByRole("link", { name: "Detalhes lançamento Receita fb6e3d8f" });
    expect(classes(button.closest("td")!).filter((c) => c === "hidden" || c.endsWith(":hidden"))).toEqual([]);
  });

  it("colunas de prioridade 1 e 2 nunca são ocultadas", () => {
    render(<Lista />);
    for (const name of ["Descrição", "Status", "Valor", "Conta", "Data", "Tipo"]) {
      const header = screen.getByRole("columnheader", { name });
      expect(classes(header).filter((c) => c === "hidden" || c.endsWith(":hidden"))).toEqual([]);
    }
    expect(classes(screen.getByText("Bilheteria do jogo de sábado"))).not.toContain("hidden");
  });

  it("sem ações, não cria a coluna de ações", () => {
    render(<Lista primaryAction={undefined} moreActions={undefined} />);
    expect(screen.queryByRole("columnheader", { name: "Ações" })).toBeNull();
    expect(screen.getAllByRole("columnheader")).toHaveLength(8);
  });
});

// DSP-05 AC7: abaixo de 768 px, item empilhado.
describe("lista empilhada", () => {
  it("em tela estreita renderiza lista com nome acessível e nenhuma tabela", () => {
    mockViewport({ width: 390, coarse: true });
    render(<Lista />);
    const list = screen.getByRole("list", { name: "Lançamentos" });
    expect(list.tagName).toBe("UL");
    expect(within(list).getAllByRole("listitem")).toHaveLength(2);
    expect(screen.queryByRole("table")).toBeNull();
    expect(document.querySelector("table")).toBeNull();
  });

  it("a lista declara role=list, para não perder o papel sem marcadores", () => {
    mockViewport({ width: 390, coarse: true });
    render(<Lista />);
    expect(screen.getByRole("list", { name: "Lançamentos" }).getAttribute("role")).toBe("list");
  });

  it("cada item tem título, até três campos, status, valor e ações", () => {
    mockViewport({ width: 390 });
    render(<Lista />);
    const [item] = screen.getAllByRole("listitem");
    const slot = (name: string) => item.querySelector(`[data-slot=data-list-${name}]`) as HTMLElement;

    expect(slot("title").textContent).toBe("Bilheteria do jogo de sábado");
    expect(classes(slot("title"))).toContain("font-semibold");
    expect(slot("meta").textContent).toBe("Recebida");
    expect(slot("value").textContent).toBe("Valor: R$ 198,50");
    expect(classes(slot("value"))).toContain("tabular-nums");

    // Quatro colunas pedem "field"; só as três primeiras entram.
    const fields = [...slot("fields").querySelectorAll("[data-slot=data-list-field]")];
    expect(fields.map((field) => field.textContent)).toEqual(["Conta: Bilheteria", "Data: 09/10/2026", "Tipo: Receita"]);
    expect(within(item).queryByText("fb6e3d8f")).toBeNull();
    // `inList: false` fica fora do item.
    expect(within(item).queryByText("Ana Souza")).toBeNull();

    expect(within(item).getByRole("link", { name: "Detalhes lançamento Receita fb6e3d8f" })).toBeTruthy();
    expect(within(item).getByRole("button", { name: "Ações de lançamento Receita fb6e3d8f" })).toBeTruthy();
  });

  it("o rótulo de cada campo existe para leitor de tela, sem ocupar a tela", () => {
    mockViewport({ width: 390 });
    render(<Lista />);
    const label = within(screen.getAllByRole("listitem")[0]).getByText("Conta:");
    expect(classes(label)).toContain("sr-only");
  });

  it("texto longo é limitado a três linhas, no título e nos campos", () => {
    mockViewport({ width: 390 });
    render(<Lista />);
    const [item] = screen.getAllByRole("listitem");
    expect(classes(item.querySelector("[data-slot=data-list-title]") as Element)).toContain("line-clamp-3");
    expect(classes(item.querySelector("[data-slot=data-list-fields]") as Element)).toContain("line-clamp-3");
  });

  it("sem coluna marcada como título, a primeira coluna não numérica é o título", () => {
    mockViewport({ width: 390 });
    render(
      <DataList
        label="Contas"
        columns={[
          { id: "saldo", header: "Saldo", priority: 1, numeric: true, cell: (row) => row.valor },
          { id: "nome", header: "Nome", priority: 1, cell: (row) => row.conta },
        ]}
        rows={rows}
        rowKey={(row) => row.id}
        rowLabel={(row) => row.conta}
      />,
    );
    const [item] = screen.getAllByRole("listitem");
    expect(item.querySelector("[data-slot=data-list-title]")?.textContent).toBe("Bilheteria");
    expect(item.querySelector("[data-slot=data-list-value]")?.textContent).toBe("Saldo: R$ 198,50");
  });

  it("a troca segue a largura: 767 px é lista, 768 px é tabela", () => {
    const viewport = mockViewport({ width: 767 });
    render(<Lista />);
    expect(screen.getByRole("list", { name: "Lançamentos" })).toBeTruthy();
    expect(screen.queryByRole("table")).toBeNull();

    viewport.resize(768);
    expect(screen.getByRole("table", { name: "Lançamentos" })).toBeTruthy();
    expect(screen.queryByRole("list", { name: "Lançamentos" })).toBeNull();
  });
});

// DSP-05 AC8.
describe("layout forçado", () => {
  it.each([320, 390, 767])('layout="table" mantém a tabela em %i px', (width) => {
    mockViewport({ width });
    render(<Lista layout="table" />);
    expect(screen.getByRole("table", { name: "Lançamentos" })).toBeTruthy();
    expect(screen.queryByRole("list", { name: "Lançamentos" })).toBeNull();
  });

  it('layout="list" mantém a lista em tela larga', () => {
    mockViewport({ width: 1440 });
    render(<Lista layout="list" />);
    expect(screen.getByRole("list", { name: "Lançamentos" })).toBeTruthy();
    expect(screen.queryByRole("table")).toBeNull();
  });
});

// DSP-05 AC10 e AC11; WCAG 2.5.3.
describe("ações de linha", () => {
  it.each([
    ["tabela", 1280],
    ["lista", 390],
  ])("na %s, há uma ação visível e um menu por linha, com nomes únicos que identificam a linha", (_, width) => {
    mockViewport({ width });
    render(<Lista />);
    const names = [...screen.getAllByRole("link"), ...screen.getAllByRole("button")].map((el) =>
      el.getAttribute("aria-label"),
    );
    expect(names).toEqual([
      "Detalhes lançamento Receita fb6e3d8f",
      "Detalhes lançamento Despesa 0a1b2c3d",
      "Ações de lançamento Receita fb6e3d8f",
      "Ações de lançamento Despesa 0a1b2c3d",
    ]);
    expect(new Set(names).size).toBe(names.length);
  });

  it("o nome acessível da ação visível contém o texto visível", () => {
    render(<Lista />);
    const action = screen.getByRole("link", { name: "Detalhes lançamento Receita fb6e3d8f" });
    expect(action.textContent).toBe("Detalhes");
    expect(action.getAttribute("href")).toBe("/l/fb6e3d8f");
  });

  it("o menu reúne as demais ações, com o texto visível e o nome que identifica a linha", async () => {
    render(<Lista />);
    fireEvent.click(screen.getByRole("button", { name: "Ações de lançamento Despesa 0a1b2c3d" }));
    const menu = await screen.findByRole("menu");
    const items = within(menu).getAllByRole("menuitem");
    expect(items.map((item) => item.textContent)).toEqual(["Editar", "Cancelar lançamento"]);
    expect(items.map((item) => item.getAttribute("aria-label"))).toEqual([
      "Editar lançamento Despesa 0a1b2c3d",
      "Cancelar lançamento lançamento Despesa 0a1b2c3d",
    ]);
  });

  it("linha sem ações de menu não mostra o botão de menu", () => {
    render(<Lista moreActions={(row) => (row.id === "fb6e3d8f" ? [{ label: "Editar" }] : [])} />);
    expect(screen.getByRole("button", { name: "Ações de lançamento Receita fb6e3d8f" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Ações de lançamento Despesa 0a1b2c3d" })).toBeNull();
  });
});

describe("RowActions", () => {
  it("escolher um item chama a ação e fecha o menu", async () => {
    const onEdit = vi.fn();
    render(<RowActions rowLabel="Ana Souza" actions={[{ label: "Editar papéis", onSelect: onEdit }]} />);
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana Souza" }));
    fireEvent.click(await screen.findByRole("menuitem", { name: "Editar papéis Ana Souza" }));
    expect(onEdit).toHaveBeenCalledTimes(1);
    await waitFor(() => expect(screen.queryByRole("menu")).toBeNull());
  });

  it("item com destino é um link", async () => {
    render(<RowActions rowLabel="Ana Souza" actions={[{ label: "Ver detalhes", href: "/admin/usuarios/1" }]} />);
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana Souza" }));
    const item = await screen.findByRole("menuitem", { name: "Ver detalhes Ana Souza" });
    expect(item.tagName).toBe("A");
    expect(item.getAttribute("href")).toBe("/admin/usuarios/1");
  });

  it("a ação destrutiva vai por último, depois de um separador, com a variante destrutiva", async () => {
    render(
      <RowActions
        rowLabel="Ana Souza"
        actions={[{ label: "Desativar", destructive: true }, { label: "Editar papéis" }]}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana Souza" }));
    const menu = await screen.findByRole("menu");
    const items = within(menu).getAllByRole("menuitem");
    expect(items.map((item) => item.textContent)).toEqual(["Editar papéis", "Desativar"]);
    expect(items[1].getAttribute("data-variant")).toBe("destructive");
    const separator = within(menu).getByRole("separator");
    expect(items[0].compareDocumentPosition(separator) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(separator.compareDocumentPosition(items[1]) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it("item desabilitado não chama a ação", async () => {
    const onSelect = vi.fn();
    render(<RowActions rowLabel="Ana Souza" actions={[{ label: "Editar papéis", onSelect, disabled: true }]} />);
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana Souza" }));
    const item = await screen.findByRole("menuitem", { name: "Editar papéis Ana Souza" });
    expect(item.getAttribute("aria-disabled")).toBe("true");
    fireEvent.click(item);
    expect(onSelect).not.toHaveBeenCalled();
  });

  it("sem ação visível e sem itens, não renderiza nada", () => {
    const { container } = render(<RowActions rowLabel="Ana Souza" actions={[]} />);
    expect(container.textContent).toBe("");
    expect(screen.queryByRole("button")).toBeNull();
  });

  it("RowActionButton sem destino é um botão que chama a ação", () => {
    const onClick = vi.fn();
    render(<RowActionButton label="Devolver" rowLabel="saída de 3 unidades em 09/10/2026" onClick={onClick} />);
    const button = screen.getByRole("button", { name: "Devolver saída de 3 unidades em 09/10/2026" });
    expect(button.textContent).toBe("Devolver");
    fireEvent.click(button);
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});

// DD-2 (WDS-D-026): servidor e primeira renderização do cliente coincidem.
describe("renderização no servidor e hidratação", () => {
  it("no servidor renderiza a tabela, mesmo com uma tela estreita disponível", () => {
    mockViewport({ width: 390 });
    const html = renderToString(<Lista />);
    expect(html).toContain("<table");
    expect(html).not.toContain("<ul");
  });

  it("hidratada em tela estreita, não gera erro nem aviso de hidratação e termina em lista", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);
    container.innerHTML = renderToString(<Lista />);
    expect(container.querySelector("table")).not.toBeNull();

    mockViewport({ width: 390, coarse: true });
    const consoleError = vi.spyOn(console, "error");
    const consoleWarn = vi.spyOn(console, "warn");
    const onRecoverableError = vi.fn();

    const root = await act(async () => hydrateRoot(container, <Lista />, { onRecoverableError }));

    expect(onRecoverableError).not.toHaveBeenCalled();
    expect(consoleError).not.toHaveBeenCalled();
    expect(consoleWarn).not.toHaveBeenCalled();
    expect(container.querySelector("table")).toBeNull();
    expect(container.querySelector("ul")?.getAttribute("aria-label")).toBe("Lançamentos");
    expect(container.querySelectorAll("li")).toHaveLength(2);

    act(() => root.unmount());
    container.remove();
  });

  it("hidratada em tela larga, mantém a tabela do servidor sem erro", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);
    container.innerHTML = renderToString(<Lista />);
    const serverTable = container.querySelector("table");

    mockViewport({ width: 1440 });
    const consoleError = vi.spyOn(console, "error");
    const onRecoverableError = vi.fn();

    const root = await act(async () => hydrateRoot(container, <Lista />, { onRecoverableError }));

    expect(onRecoverableError).not.toHaveBeenCalled();
    expect(consoleError).not.toHaveBeenCalled();
    expect(container.querySelector("table")).toBe(serverTable);

    act(() => root.unmount());
    container.remove();
  });
});
