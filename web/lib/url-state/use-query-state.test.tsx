import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";

import { codecs, SEARCH_DEBOUNCE_MS, useQueryState } from "./index";

vi.mock("next/navigation", async () => (await import("@/lib/session/navigation-mock")).nextNavigationMock);

const TIPOS = ["RECEITA", "DESPESA"] as const;
const tipoCodec = codecs.oneOf(TIPOS, "");
const ABAS = ["dados", "historico"] as const;

function TipoFilter({ history }: { history?: "replace" | "push" }) {
  const [tipo, setTipo] = useQueryState("tipo", tipoCodec, { history });
  return (
    <label>
      Tipo
      <select value={tipo} onChange={(e) => setTipo(e.target.value as (typeof TIPOS)[number] | "")}>
        <option value="">Todos</option>
        <option value="RECEITA">Receita</option>
        <option value="DESPESA">Despesa</option>
      </select>
    </label>
  );
}

function StatusFilter() {
  const [status, setStatus] = useQueryState("status", codecs.oneOf(["ATIVO", "CANCELADO"] as const, ""));
  return (
    <label>
      Status
      <select value={status} onChange={(e) => setStatus(e.target.value as "ATIVO" | "CANCELADO" | "")}>
        <option value="">Todos</option>
        <option value="ATIVO">Ativo</option>
        <option value="CANCELADO">Cancelado</option>
      </select>
    </label>
  );
}

function DateFilter() {
  const [de, setDe] = useQueryState("de", codecs.isoDate());
  return (
    <label>
      De
      <input value={de} onChange={(e) => setDe(e.target.value)} />
    </label>
  );
}

function Tabs() {
  const [aba, setAba] = useQueryState("aba", codecs.oneOf(ABAS, "dados"), { history: "push" });
  return (
    <div>
      <p>Aba atual: {aba}</p>
      <button onClick={() => setAba("dados")}>Dados</button>
      <button onClick={() => setAba("historico")}>Histórico</button>
    </div>
  );
}

function Search({ history }: { history?: "replace" | "push" }) {
  const [busca, setBusca] = useQueryState("busca", codecs.string(), { history, debounceMs: SEARCH_DEBOUNCE_MS });
  return (
    <label>
      Busca
      <input value={busca} onChange={(e) => setBusca(e.target.value)} />
    </label>
  );
}

const select = (name: string) => screen.getByRole("combobox", { name }) as HTMLSelectElement;
const input = (name: string) => screen.getByRole("textbox", { name }) as HTMLInputElement;
const choose = (name: string, value: string) => fireEvent.change(select(name), { target: { value } });
const type = (name: string, value: string) => fireEvent.change(input(name), { target: { value } });
const url = () => window.location.pathname + window.location.search + window.location.hash;

// Espiões instalados depois da renderização, por cima da escuta do mock.
function spyHistory() {
  return {
    replace: vi.spyOn(window.history, "replaceState"),
    push: vi.spyOn(window.history, "pushState"),
  };
}

beforeEach(() => resetNavigation());
afterEach(() => {
  vi.restoreAllMocks();
  vi.useRealTimers();
});

// DSP-05 AC2
describe("escrita na URL", () => {
  it("alterar o valor atualiza a query string e o valor devolvido", () => {
    setLocation("/financeiro/lancamentos");
    render(<TipoFilter />);

    choose("Tipo", "RECEITA");

    expect(url()).toBe("/financeiro/lancamentos?tipo=RECEITA");
    expect(select("Tipo").value).toBe("RECEITA");
  });

  it("trocar de um valor para outro substitui o parâmetro, sem duplicar", () => {
    setLocation("/financeiro/lancamentos?tipo=RECEITA");
    render(<TipoFilter />);

    choose("Tipo", "DESPESA");

    expect(url()).toBe("/financeiro/lancamentos?tipo=DESPESA");
  });

  it("voltar ao padrão remove o parâmetro e preserva os demais, o caminho e a âncora", () => {
    setLocation("/financeiro/lancamentos?conta=7&tipo=RECEITA&status=ATIVO#lista");
    render(<TipoFilter />);

    choose("Tipo", "");

    expect(url()).toBe("/financeiro/lancamentos?conta=7&status=ATIVO#lista");
    expect(select("Tipo").value).toBe("");
  });

  it("remover o único parâmetro deixa a URL limpa, sem interrogação", () => {
    setLocation("/financeiro/lancamentos?tipo=RECEITA");
    render(<TipoFilter />);

    choose("Tipo", "");

    expect(url()).toBe("/financeiro/lancamentos");
  });

  it("escrever preserva os parâmetros de outras chaves", () => {
    setLocation("/financeiro/lancamentos?conta=7&de=2026-01-01");
    render(<TipoFilter />);

    choose("Tipo", "DESPESA");

    const params = new URLSearchParams(window.location.search);
    expect(window.location.pathname).toBe("/financeiro/lancamentos");
    expect(Object.fromEntries(params)).toEqual({ conta: "7", de: "2026-01-01", tipo: "DESPESA" });
  });

  it("valor com espaço e acento vai codificado e volta igual", () => {
    setLocation("/estoque/produtos");
    function Imediata() {
      const [busca, setBusca] = useQueryState("busca", codecs.string());
      return (
        <label>
          Busca
          <input value={busca} onChange={(e) => setBusca(e.target.value)} />
        </label>
      );
    }
    render(<Imediata />);

    type("Busca", "camisa & boné");

    expect(new URLSearchParams(window.location.search).get("busca")).toBe("camisa & boné");
    expect(window.location.search).not.toContain(" ");
    expect(input("Busca").value).toBe("camisa & boné");
  });

  it("dois hooks de chaves diferentes na mesma tela não se sobrescrevem", () => {
    setLocation("/financeiro/lancamentos");
    render(
      <>
        <TipoFilter />
        <StatusFilter />
      </>,
    );

    choose("Tipo", "RECEITA");
    choose("Status", "CANCELADO");
    expect(url()).toBe("/financeiro/lancamentos?tipo=RECEITA&status=CANCELADO");

    choose("Tipo", "");
    expect(url()).toBe("/financeiro/lancamentos?status=CANCELADO");
    expect(select("Status").value).toBe("CANCELADO");
  });

  it("duas escritas no mesmo evento, de chaves diferentes, ficam as duas na URL", () => {
    setLocation("/financeiro/lancamentos?conta=7");
    function Atalho() {
      const [tipo, setTipo] = useQueryState("tipo", tipoCodec);
      const [de, setDe] = useQueryState("de", codecs.isoDate());
      return (
        <button
          onClick={() => {
            setTipo("DESPESA");
            setDe("2026-10-01");
          }}
        >
          {tipo || "todos"} desde {de || "sempre"}
        </button>
      );
    }
    render(<Atalho />);

    fireEvent.click(screen.getByRole("button"));

    expect(url()).toBe("/financeiro/lancamentos?conta=7&tipo=DESPESA&de=2026-10-01");
    expect(screen.getByRole("button").textContent).toBe("DESPESA desde 2026-10-01");
  });
});

// DSP-05 AC3 e AC4
describe("leitura da URL", () => {
  it("abrir com o parâmetro presente restaura o valor", () => {
    setLocation("/financeiro/lancamentos?tipo=DESPESA&de=2026-03-15");
    render(
      <>
        <TipoFilter />
        <DateFilter />
      </>,
    );

    expect(select("Tipo").value).toBe("DESPESA");
    expect(input("De").value).toBe("2026-03-15");
  });

  it("sem o parâmetro, devolve o padrão", () => {
    setLocation("/financeiro/lancamentos?conta=7");
    render(
      <>
        <TipoFilter />
        <Tabs />
      </>,
    );

    expect(select("Tipo").value).toBe("");
    expect(screen.getByText("Aba atual: dados")).toBeTruthy();
  });

  it("valor inválido cai no padrão, sem erro e sem mexer na URL", () => {
    const errors = vi.spyOn(console, "error");
    setLocation("/financeiro/lancamentos?tipo=QUALQUER&de=2026-02-30&aba=nada");
    render(
      <>
        <TipoFilter />
        <DateFilter />
        <Tabs />
      </>,
    );

    expect(select("Tipo").value).toBe("");
    expect(input("De").value).toBe("");
    expect(screen.getByText("Aba atual: dados")).toBeTruthy();
    expect(url()).toBe("/financeiro/lancamentos?tipo=QUALQUER&de=2026-02-30&aba=nada");
    expect(errors).not.toHaveBeenCalled();
  });

  it("acompanha a URL quando ela muda por fora (navegação ou outro componente)", () => {
    setLocation("/financeiro/lancamentos?tipo=RECEITA");
    render(<TipoFilter />);

    act(() => window.history.pushState(null, "", "/financeiro/lancamentos?tipo=DESPESA"));
    expect(select("Tipo").value).toBe("DESPESA");

    setLocation("/financeiro/lancamentos");
    expect(select("Tipo").value).toBe("");
  });

  it("depois de escrever, continua acompanhando a URL", () => {
    setLocation("/financeiro/lancamentos");
    render(<TipoFilter />);
    choose("Tipo", "RECEITA");

    setLocation("/financeiro/lancamentos?tipo=DESPESA");
    expect(select("Tipo").value).toBe("DESPESA");

    setLocation("/financeiro/lancamentos?tipo=RECEITA");
    expect(select("Tipo").value).toBe("RECEITA");

    setLocation("/financeiro/lancamentos");
    expect(select("Tipo").value).toBe("");
  });
});

// Caso de borda da spec: filtros e busca substituem a entrada; aba cria entrada.
describe("histórico", () => {
  it("por padrão usa replaceState e não cria entrada de histórico", () => {
    setLocation("/financeiro/lancamentos");
    render(<TipoFilter />);
    const history = spyHistory();
    const before = window.history.length;

    choose("Tipo", "RECEITA");
    choose("Tipo", "DESPESA");

    expect(window.history.length).toBe(before);
    expect(history.replace).toHaveBeenCalledTimes(2);
    expect(history.push).not.toHaveBeenCalled();
  });

  it('history: "replace" explícito também não cria entrada', () => {
    setLocation("/financeiro/lancamentos");
    render(<TipoFilter history="replace" />);
    const before = window.history.length;

    choose("Tipo", "RECEITA");

    expect(window.history.length).toBe(before);
    expect(url()).toBe("/financeiro/lancamentos?tipo=RECEITA");
  });

  it('history: "push" cria uma entrada por mudança', () => {
    setLocation("/admin/usuarios/3");
    render(<Tabs />);
    const history = spyHistory();
    const before = window.history.length;

    fireEvent.click(screen.getByRole("button", { name: "Histórico" }));

    expect(window.history.length).toBe(before + 1);
    expect(history.push).toHaveBeenCalledTimes(1);
    expect(history.replace).not.toHaveBeenCalled();
    expect(url()).toBe("/admin/usuarios/3?aba=historico");
    expect(screen.getByText("Aba atual: historico")).toBeTruthy();
  });

  it('com "push", escolher o valor que já está na URL não cria entrada', () => {
    setLocation("/admin/usuarios/3?aba=historico");
    render(<Tabs />);
    const before = window.history.length;

    fireEvent.click(screen.getByRole("button", { name: "Histórico" }));

    expect(window.history.length).toBe(before);
  });

  it('com "push", voltar no navegador restaura o valor anterior', async () => {
    setLocation("/admin/usuarios/3");
    render(<Tabs />);
    fireEvent.click(screen.getByRole("button", { name: "Histórico" }));
    expect(screen.getByText("Aba atual: historico")).toBeTruthy();

    window.history.back();

    await waitFor(() => expect(screen.getByText("Aba atual: dados")).toBeTruthy());
    expect(url()).toBe("/admin/usuarios/3");
  });
});

// Caso de borda da spec: a busca muda a cada tecla sem uma escrita por tecla.
describe("escrita adiada (busca)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  const advance = (ms: number) => act(() => vi.advanceTimersByTime(ms));

  it("a espera padrão da busca é de 300 ms", () => {
    expect(SEARCH_DEBOUNCE_MS).toBe(300);
  });

  it("o valor devolvido acompanha a digitação antes de a URL mudar", () => {
    setLocation("/estoque/produtos");
    render(<Search />);

    type("Busca", "c");
    type("Busca", "ca");

    expect(input("Busca").value).toBe("ca");
    expect(url()).toBe("/estoque/produtos");
  });

  it("várias mudanças seguidas geram uma única escrita, com o último valor", () => {
    setLocation("/estoque/produtos?categoria=2");
    render(<Search />);
    const history = spyHistory();
    const before = window.history.length;

    type("Busca", "c");
    advance(100);
    type("Busca", "ca");
    advance(100);
    type("Busca", "cam");
    advance(299);
    expect(history.replace).not.toHaveBeenCalled();
    expect(url()).toBe("/estoque/produtos?categoria=2");

    advance(1);

    expect(history.replace).toHaveBeenCalledTimes(1);
    expect(history.push).not.toHaveBeenCalled();
    expect(url()).toBe("/estoque/produtos?categoria=2&busca=cam");
    expect(window.history.length).toBe(before);
    expect(input("Busca").value).toBe("cam");

    advance(1000);
    expect(history.replace).toHaveBeenCalledTimes(1);
  });

  it('com "push", a rajada cria uma única entrada de histórico', () => {
    setLocation("/estoque/produtos");
    render(<Search history="push" />);
    const history = spyHistory();

    type("Busca", "c");
    type("Busca", "ca");
    type("Busca", "cam");
    advance(300);

    // Contado pela chamada: o comprimento do histórico do jsdom depende das
    // voltas feitas por testes anteriores.
    expect(history.push).toHaveBeenCalledTimes(1);
    expect(history.replace).not.toHaveBeenCalled();
    expect(url()).toBe("/estoque/produtos?busca=cam");
  });

  it("apagar tudo remove o parâmetro depois da espera", () => {
    setLocation("/estoque/produtos?busca=cam&categoria=2");
    render(<Search />);
    expect(input("Busca").value).toBe("cam");

    type("Busca", "");
    expect(input("Busca").value).toBe("");
    advance(300);

    expect(url()).toBe("/estoque/produtos?categoria=2");
    expect(input("Busca").value).toBe("");
  });

  it("digitar e voltar ao valor que já está na URL não escreve nada e não prende o valor", () => {
    setLocation("/estoque/produtos?busca=cam");
    render(<Search />);
    const history = spyHistory();

    type("Busca", "cami");
    type("Busca", "cam");
    advance(300);
    expect(history.replace).not.toHaveBeenCalled();
    expect(input("Busca").value).toBe("cam");

    vi.restoreAllMocks();
    setLocation("/estoque/produtos?busca=bone");
    expect(input("Busca").value).toBe("bone");
  });

  it("a escrita adiada parte da URL do momento da escrita, não da digitação", () => {
    setLocation("/estoque/produtos");
    render(
      <>
        <Search />
        <TipoFilter />
      </>,
    );

    type("Busca", "cam");
    choose("Tipo", "RECEITA");
    expect(url()).toBe("/estoque/produtos?tipo=RECEITA");
    expect(input("Busca").value).toBe("cam");

    advance(300);

    expect(url()).toBe("/estoque/produtos?tipo=RECEITA&busca=cam");
    expect(select("Tipo").value).toBe("RECEITA");
  });

  it("depois da escrita, volta a acompanhar a URL", () => {
    setLocation("/estoque/produtos");
    render(<Search />);
    type("Busca", "cam");
    advance(300);

    setLocation("/estoque/produtos");

    expect(input("Busca").value).toBe("");
  });

  it("desmontar antes do prazo limpa o temporizador e não escreve na URL", () => {
    setLocation("/estoque/produtos");
    const view = render(<Search />);
    const history = spyHistory();

    type("Busca", "cam");
    expect(vi.getTimerCount()).toBe(1);
    view.unmount();

    expect(vi.getTimerCount()).toBe(0);
    advance(1000);
    expect(history.replace).not.toHaveBeenCalled();
    expect(url()).toBe("/estoque/produtos");
  });
});
