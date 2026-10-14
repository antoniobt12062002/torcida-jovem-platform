import { act, render, screen } from "@testing-library/react";
import { hydrateRoot } from "react-dom/client";
import { renderToString } from "react-dom/server";
import { afterEach, describe, expect, it, vi } from "vitest";

import { mockViewport } from "@/test/media";

import { BREAKPOINT_LG, BREAKPOINT_MD, useIsCoarsePointer, useIsNarrow, useMediaQuery } from "./index";

const DESKTOP = "layout=tabela ponteiro=fino largo=sim";
const NARROW = "layout=lista ponteiro=grosso largo=nao";

// Cada renderização é anotada, para ver o que a primeira do cliente mostrou.
let renders: string[] = [];

function Probe() {
  const narrow = useIsNarrow();
  const coarse = useIsCoarsePointer();
  const wide = useMediaQuery(`(min-width: ${BREAKPOINT_LG}px)`);
  const text = `layout=${narrow ? "lista" : "tabela"} ponteiro=${coarse ? "grosso" : "fino"} largo=${wide ? "sim" : "nao"}`;
  renders.push(text);
  return <p>{text}</p>;
}

// Controle: lê a tela direto na renderização, sem retrato de servidor.
function NaiveProbe() {
  const narrow = typeof window.matchMedia === "function" && !window.matchMedia("(min-width: 768px)").matches;
  return <p>layout={narrow ? "lista" : "tabela"}</p>;
}

function Query({ query }: { query: string }) {
  return <p>{useMediaQuery(query) ? "casa" : "não casa"}</p>;
}

afterEach(() => {
  renders = [];
  vi.restoreAllMocks();
});

describe("pontos de quebra", () => {
  it("são 768 e 1024 px", () => {
    expect(BREAKPOINT_MD).toBe(768);
    expect(BREAKPOINT_LG).toBe(1024);
  });
});

// DD-2 (WDS-D-026): servidor e primeira renderização do cliente coincidem.
describe("renderização no servidor e hidratação", () => {
  it("no servidor devolve o valor de desktop", () => {
    expect(renderToString(<Probe />)).toBe(`<p>${DESKTOP}</p>`);
  });

  it("no servidor devolve desktop mesmo com uma tela estreita disponível", () => {
    mockViewport({ width: 390, coarse: true });

    expect(renderToString(<Probe />)).toBe(`<p>${DESKTOP}</p>`);
  });

  it("hidratado em tela estreita, parte do valor do servidor e atualiza sem aviso de hidratação", async () => {
    const container = document.createElement("div");
    document.body.appendChild(container);
    container.innerHTML = renderToString(<Probe />);
    const serverNode = container.firstElementChild;
    expect(container.textContent).toBe(DESKTOP);

    mockViewport({ width: 390, coarse: true });
    renders = [];
    const consoleError = vi.spyOn(console, "error");
    const consoleWarn = vi.spyOn(console, "warn");
    const onRecoverableError = vi.fn();

    const root = await act(async () => hydrateRoot(container, <Probe />, { onRecoverableError }));

    expect(onRecoverableError).not.toHaveBeenCalled();
    expect(consoleError).not.toHaveBeenCalled();
    expect(consoleWarn).not.toHaveBeenCalled();
    // A primeira renderização do cliente repete o HTML do servidor.
    expect(renders[0]).toBe(DESKTOP);
    // Depois da hidratação, vale a tela de verdade.
    expect(renders.at(-1)).toBe(NARROW);
    expect(container.textContent).toBe(NARROW);
    // O nó do servidor foi aproveitado, não recriado.
    expect(container.firstElementChild).toBe(serverNode);

    act(() => root.unmount());
    container.remove();
  });

  it("hidratado em tela de desktop, não renderiza de novo com outro valor", async () => {
    const container = document.createElement("div");
    container.innerHTML = renderToString(<Probe />);
    mockViewport({ width: 1440 });
    renders = [];
    const consoleError = vi.spyOn(console, "error");
    const onRecoverableError = vi.fn();

    const root = await act(async () => hydrateRoot(container, <Probe />, { onRecoverableError }));

    expect(onRecoverableError).not.toHaveBeenCalled();
    expect(consoleError).not.toHaveBeenCalled();
    expect(new Set(renders)).toEqual(new Set([DESKTOP]));
    expect(container.textContent).toBe(DESKTOP);

    act(() => root.unmount());
  });

  it("controle: ler matchMedia direto na renderização diverge e é acusado na hidratação", async () => {
    const container = document.createElement("div");
    container.innerHTML = renderToString(<NaiveProbe />);
    expect(container.textContent).toBe("layout=tabela");

    mockViewport({ width: 390 });
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
    const onRecoverableError = vi.fn();

    const root = await act(async () => hydrateRoot(container, <NaiveProbe />, { onRecoverableError }));

    expect(onRecoverableError.mock.calls.length + consoleError.mock.calls.length).toBeGreaterThan(0);

    act(() => root.unmount());
  });
});

describe("sem matchMedia", () => {
  it("o ambiente de teste não tem matchMedia", () => {
    expect(typeof window.matchMedia).toBe("undefined");
  });

  it("os helpers devolvem desktop", () => {
    render(<Probe />);

    expect(screen.getByText(DESKTOP)).toBeTruthy();
  });

  it.each([
    ["(min-width: 768px)", "casa"],
    ["(min-width: 1024px)", "casa"],
    ["(max-width: 767px)", "não casa"],
    ["(pointer: coarse)", "não casa"],
    ["(pointer: fine)", "casa"],
    ["(min-width: 768px) and (max-width: 1023px)", "não casa"],
    ["(min-width: 768px) and (pointer: fine)", "casa"],
    ["(min-width: 768px) and (pointer: coarse)", "não casa"],
    ["(prefers-reduced-motion: reduce)", "não casa"],
  ])("useMediaQuery(%s) em desktop: %s", (query, expected) => {
    render(<Query query={query} />);

    expect(screen.getByText(expected)).toBeTruthy();
  });
});

// DSP-05 AC5 a AC8: a troca de layout segue a largura da tela.
describe("com matchMedia", () => {
  it("lê a tela já na primeira renderização do cliente", () => {
    mockViewport({ width: 390, coarse: true });

    render(<Probe />);

    expect(renders).toEqual([NARROW]);
  });

  it.each([
    [390, true],
    [767, true],
    [768, false],
    [820, false],
    [1024, false],
    [1440, false],
  ])("em %i px, useIsNarrow devolve %s", (width, narrow) => {
    mockViewport({ width });
    function Narrow() {
      return <p>{String(useIsNarrow())}</p>;
    }

    render(<Narrow />);

    expect(screen.getByText(String(narrow))).toBeTruthy();
  });

  it("useIsCoarsePointer segue o tipo de ponteiro, não a largura", () => {
    mockViewport({ width: 1440, coarse: true });

    render(<Probe />);

    expect(screen.getByText("layout=tabela ponteiro=grosso largo=sim")).toBeTruthy();
  });

  it("mudar a largura atualiza o componente, nos dois sentidos", () => {
    const viewport = mockViewport({ width: 1440 });
    render(<Probe />);
    expect(screen.getByText(DESKTOP)).toBeTruthy();

    viewport.resize(820);
    expect(screen.getByText("layout=tabela ponteiro=fino largo=nao")).toBeTruthy();

    viewport.resize(390);
    expect(screen.getByText("layout=lista ponteiro=fino largo=nao")).toBeTruthy();

    viewport.resize(1024);
    expect(screen.getByText(DESKTOP)).toBeTruthy();
  });

  it("trocar a consulta passa a seguir a consulta nova", () => {
    const viewport = mockViewport({ width: 800 });
    const view = render(<Query query="(min-width: 768px)" />);
    expect(screen.getByText("casa")).toBeTruthy();

    view.rerender(<Query query="(min-width: 1024px)" />);
    expect(screen.getByText("não casa")).toBeTruthy();

    viewport.resize(1024);
    expect(screen.getByText("casa")).toBeTruthy();
  });

  it("depois de desmontar, não reage mais à mudança de largura", () => {
    const viewport = mockViewport({ width: 1440 });
    const view = render(<Probe />);
    view.unmount();
    renders = [];

    viewport.resize(390);

    expect(renders).toEqual([]);
  });
});
