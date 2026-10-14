import { act, render, screen, waitFor } from "@testing-library/react";
import { useEffect } from "react";
import { beforeEach, describe, expect, it } from "vitest";

import { navigations, nextNavigationMock, resetNavigation, router, setLocation } from "./navigation-mock";

const { usePathname, useRouter, useSearchParams } = nextNavigationMock;

beforeEach(() => resetNavigation());

let renders = 0;
let seenParams: URLSearchParams[] = [];

beforeEach(() => {
  renders = 0;
  seenParams = [];
});

/** Lê a URL como uma tela faria e registra cada renderização aplicada. */
function Probe() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  useEffect(() => {
    renders += 1;
    seenParams.push(searchParams);
  });
  return (
    <>
      <p data-testid="pathname">{pathname}</p>
      <p data-testid="tipo">{searchParams.get("tipo") ?? "(sem tipo)"}</p>
      <p data-testid="query">{searchParams.toString()}</p>
    </>
  );
}

const pathname = () => screen.getByTestId("pathname").textContent;
const tipo = () => screen.getByTestId("tipo").textContent;
const query = () => screen.getByTestId("query").textContent;

describe("leitura da URL atual", () => {
  it("devolve o caminho e os parâmetros definidos antes da renderização", () => {
    setLocation("/financeiro/lancamentos?tipo=RECEITA&status=PAGO");
    render(<Probe />);
    expect(pathname()).toBe("/financeiro/lancamentos");
    expect(query()).toBe("tipo=RECEITA&status=PAGO");
  });
});

// DSP-05 AC2 e AC3: quem escreve na URL é renderizado de novo com o valor novo.
describe("reação ao histórico", () => {
  it("replaceState atualiza useSearchParams na renderização seguinte", () => {
    setLocation("/financeiro/lancamentos");
    render(<Probe />);
    expect(tipo()).toBe("(sem tipo)");

    act(() => window.history.replaceState(null, "", "?tipo=DESPESA"));

    expect(tipo()).toBe("DESPESA");
    expect(pathname()).toBe("/financeiro/lancamentos");
  });

  it("pushState atualiza useSearchParams e usePathname na renderização seguinte", () => {
    setLocation("/financeiro/lancamentos");
    render(<Probe />);

    act(() => window.history.pushState(null, "", "/estoque/produtos?tipo=RECEITA"));

    expect(tipo()).toBe("RECEITA");
    expect(pathname()).toBe("/estoque/produtos");
  });

  it("setLocation atualiza um componente já montado", () => {
    setLocation("/financeiro/lancamentos");
    render(<Probe />);

    setLocation("/financeiro/contas?tipo=RECEITA");

    expect(tipo()).toBe("RECEITA");
    expect(pathname()).toBe("/financeiro/contas");
  });

  it("voltar no histórico (popstate) devolve o valor anterior", async () => {
    setLocation("/financeiro/lancamentos?tipo=RECEITA");
    render(<Probe />);
    act(() => window.history.pushState(null, "", "?tipo=DESPESA"));
    expect(tipo()).toBe("DESPESA");

    window.history.back();

    await waitFor(() => expect(tipo()).toBe("RECEITA"));
  });

  it("um componente que escreve na URL num efeito vê o valor que escreveu", () => {
    function Writer() {
      const searchParams = useSearchParams();
      useEffect(() => {
        if (!searchParams.has("aba")) window.history.replaceState(null, "", "?aba=dados");
      }, [searchParams]);
      return <p data-testid="aba">{searchParams.get("aba") ?? "(sem aba)"}</p>;
    }
    setLocation("/estoque/produtos/1");
    render(<Writer />);
    expect(screen.getByTestId("aba").textContent).toBe("dados");
  });
});

describe("estabilidade entre renderizações", () => {
  it("devolve o mesmo objeto enquanto a query não muda", () => {
    setLocation("/financeiro/lancamentos?tipo=RECEITA");
    const { rerender } = render(<Probe />);
    rerender(<Probe />);
    // Mudança só de fragmento: a query é a mesma.
    act(() => window.history.replaceState(null, "", "?tipo=RECEITA#topo"));
    rerender(<Probe />);

    expect(seenParams.length).toBeGreaterThanOrEqual(3);
    for (const params of seenParams) expect(params).toBe(seenParams[0]);
  });

  it("escrita que não muda caminho nem query não renderiza de novo", () => {
    setLocation("/financeiro/lancamentos?tipo=RECEITA");
    render(<Probe />);
    const before = renders;

    act(() => window.history.replaceState(null, "", "/financeiro/lancamentos?tipo=RECEITA"));

    expect(renders).toBe(before);
  });

  it("um efeito que depende de useSearchParams não entra em laço", () => {
    let effects = 0;
    function Counter() {
      const searchParams = useSearchParams();
      useEffect(() => {
        effects += 1;
      }, [searchParams]);
      return null;
    }
    setLocation("/financeiro/lancamentos?tipo=RECEITA");
    const { rerender } = render(<Counter />);
    rerender(<Counter />);
    expect(effects).toBe(1);

    act(() => window.history.replaceState(null, "", "?tipo=DESPESA"));
    expect(effects).toBe(2);
  });
});

describe("resetNavigation", () => {
  it("limpa as chamadas registradas do roteador e volta a /", () => {
    setLocation("/financeiro/contas?tipo=RECEITA");
    router.push("/a");
    router.replace("/b");
    router.refresh();
    expect(navigations()).toEqual(["/a", "/b"]);

    resetNavigation();

    expect(navigations()).toEqual([]);
    expect(router.push).not.toHaveBeenCalled();
    expect(router.replace).not.toHaveBeenCalled();
    expect(router.refresh).not.toHaveBeenCalled();
    expect(window.location.pathname).toBe("/");
    expect(window.location.search).toBe("");
  });

  it("limpa as assinaturas: quem ficou montado deixa de ser avisado", () => {
    setLocation("/financeiro/lancamentos?tipo=RECEITA");
    render(<Probe />);
    resetNavigation();
    const before = renders;

    act(() => window.history.pushState(null, "", "/estoque/produtos?tipo=DESPESA"));

    expect(renders).toBe(before);
    expect(tipo()).toBe("RECEITA");
  });

  it("depois de limpar, um componente novo volta a reagir", () => {
    render(<Probe />);
    resetNavigation();
    const { getAllByTestId } = render(<Probe />);

    act(() => window.history.replaceState(null, "", "?tipo=DESPESA"));

    expect(getAllByTestId("tipo").at(-1)?.textContent).toBe("DESPESA");
  });
});

// O mock não pode deixar o histórico do jsdom alterado para o teste seguinte.
describe("histórico nativo", () => {
  const isNative = () =>
    window.history.pushState === History.prototype.pushState &&
    window.history.replaceState === History.prototype.replaceState;

  it("volta ao original quando o último componente desmonta", () => {
    const first = render(<Probe />);
    const second = render(<Probe />);
    first.unmount();
    act(() => window.history.replaceState(null, "", "?tipo=DESPESA"));
    expect(second.getByTestId("tipo").textContent).toBe("DESPESA");

    second.unmount();

    expect(isNative()).toBe(true);
  });

  it("volta ao original em resetNavigation, mesmo com componente montado", () => {
    render(<Probe />);
    resetNavigation();
    expect(isNative()).toBe(true);
  });
});

describe("API pública", () => {
  it("useRouter devolve o roteador espião e navigations lista push e replace na ordem", () => {
    expect(useRouter()).toBe(router);
    router.replace("/primeiro");
    router.push("/segundo");
    router.replace("/terceiro");
    expect(navigations()).toEqual(["/primeiro", "/segundo", "/terceiro"]);
  });
});
