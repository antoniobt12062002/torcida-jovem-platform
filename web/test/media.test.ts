import { beforeEach, describe, expect, it, vi } from "vitest";

import { mockViewport } from "./media";

// É como os hooks de mídia detectam a ausência (o ambiente de teste declara a
// chave em `window`, sem valor).
const hasMatchMedia = () => typeof window.matchMedia === "function";
const matches = (query: string) => window.matchMedia(query).matches;

// DSP-05 AC5 a AC8: sem o apoio, o ambiente continua como sempre foi, e os
// hooks de mídia tratam a ausência de matchMedia como desktop.
describe("ambiente padrão", () => {
  it("o jsdom continua sem matchMedia", () => {
    expect(hasMatchMedia()).toBe(false);
  });
});

describe("consultas de largura", () => {
  it("min-width casa a partir da largura informada, inclusive", () => {
    mockViewport({ width: 768 });
    expect(matches("(min-width: 768px)")).toBe(true);
    expect(matches("(min-width: 640px)")).toBe(true);
    expect(matches("(min-width: 769px)")).toBe(false);
    expect(matches("(min-width: 1024px)")).toBe(false);
  });

  it("max-width casa até a largura informada, inclusive", () => {
    mockViewport({ width: 390 });
    expect(matches("(max-width: 390px)")).toBe(true);
    expect(matches("(max-width: 767px)")).toBe(true);
    expect(matches("(max-width: 389px)")).toBe(false);
  });

  it("devolve a consulta pedida em media", () => {
    mockViewport({ width: 390 });
    expect(window.matchMedia("(min-width: 768px)").media).toBe("(min-width: 768px)");
  });
});

describe("tipo de ponteiro", () => {
  it("pointer: coarse casa só quando o ponteiro é grosso", () => {
    mockViewport({ width: 390, coarse: true });
    expect(matches("(pointer: coarse)")).toBe(true);
    expect(matches("(pointer: fine)")).toBe(false);
  });

  it("sem coarse, o ponteiro é fino", () => {
    mockViewport({ width: 1440 });
    expect(matches("(pointer: coarse)")).toBe(false);
    expect(matches("(pointer: fine)")).toBe(true);
  });

  it("consultas combinadas com and exigem todas as condições", () => {
    mockViewport({ width: 820, coarse: true });
    expect(matches("(min-width: 768px) and (pointer: coarse)")).toBe(true);
    expect(matches("(min-width: 1024px) and (pointer: coarse)")).toBe(false);
    expect(matches("(min-width: 768px) and (max-width: 1023px)")).toBe(true);
  });
});

describe("consulta fora do que o apoio simula", () => {
  it("não casa", () => {
    mockViewport({ width: 1440 });
    expect(matches("(prefers-reduced-motion: reduce)")).toBe(false);
    expect(matches("(min-width: 768px) and (orientation: landscape)")).toBe(false);
  });
});

describe("mudança de largura", () => {
  it("avisa quem ouve change, com o novo resultado", () => {
    const viewport = mockViewport({ width: 390 });
    const query = window.matchMedia("(min-width: 768px)");
    const events: { matches: boolean; media: string; type: string }[] = [];
    query.addEventListener("change", (event) => {
      const { matches, media, type } = event as MediaQueryListEvent;
      events.push({ matches, media, type });
    });

    viewport.resize(1024);

    expect(query.matches).toBe(true);
    expect(events).toEqual([{ matches: true, media: "(min-width: 768px)", type: "change" }]);

    viewport.resize(390);

    expect(query.matches).toBe(false);
    expect(events).toHaveLength(2);
    expect(events[1].matches).toBe(false);
  });

  it("avisa também os ouvintes legados (addListener) e onchange", () => {
    const viewport = mockViewport({ width: 390 });
    const query = window.matchMedia("(min-width: 768px)");
    const legacy = vi.fn();
    const onchange = vi.fn();
    query.addListener(legacy);
    query.onchange = onchange;

    viewport.resize(768);

    expect(legacy).toHaveBeenCalledTimes(1);
    expect(legacy.mock.calls[0][0].matches).toBe(true);
    expect(onchange).toHaveBeenCalledTimes(1);
  });

  it("não avisa quando o resultado da consulta não muda", () => {
    const viewport = mockViewport({ width: 390 });
    const listener = vi.fn();
    window.matchMedia("(min-width: 768px)").addEventListener("change", listener);

    viewport.resize(640);

    expect(listener).not.toHaveBeenCalled();
  });

  it("para de avisar quem removeu o ouvinte", () => {
    const viewport = mockViewport({ width: 390 });
    const query = window.matchMedia("(min-width: 768px)");
    const modern = vi.fn();
    const legacy = vi.fn();
    query.addEventListener("change", modern);
    query.addListener(legacy);
    query.removeEventListener("change", modern);
    query.removeListener(legacy);

    viewport.resize(1024);

    expect(modern).not.toHaveBeenCalled();
    expect(legacy).not.toHaveBeenCalled();
  });

  it("avisa cada consulta criada, de forma independente", () => {
    const viewport = mockViewport({ width: 390 });
    const medium = vi.fn();
    const large = vi.fn();
    window.matchMedia("(min-width: 768px)").addEventListener("change", medium);
    window.matchMedia("(min-width: 1024px)").addEventListener("change", large);

    viewport.resize(820);

    expect(medium).toHaveBeenCalledTimes(1);
    expect(large).not.toHaveBeenCalled();
  });
});

describe("restauração", () => {
  it("restore devolve o ambiente sem matchMedia e pode ser chamado de novo", () => {
    const viewport = mockViewport({ width: 390 });
    expect(hasMatchMedia()).toBe(true);

    viewport.restore();
    viewport.restore();

    expect(hasMatchMedia()).toBe(false);
  });

  it("restore devolve o matchMedia que já existia", () => {
    const original = Object.getOwnPropertyDescriptor(window, "matchMedia");
    const previous = vi.fn();
    Object.defineProperty(window, "matchMedia", { configurable: true, writable: true, value: previous });
    try {
      const viewport = mockViewport({ width: 390 });
      expect(window.matchMedia).not.toBe(previous);
      viewport.restore();
      expect(window.matchMedia).toBe(previous);
    } finally {
      if (original) Object.defineProperty(window, "matchMedia", original);
      else Reflect.deleteProperty(window, "matchMedia");
    }
  });

  it("uma segunda chamada no mesmo teste substitui a primeira", () => {
    const first = mockViewport({ width: 390 });
    mockViewport({ width: 1440 });
    expect(matches("(min-width: 768px)")).toBe(true);

    // A restauração da chamada antiga não desfaz a nova.
    first.restore();
    expect(matches("(min-width: 768px)")).toBe(true);
  });

  it("depois de restaurar, resize não avisa mais ninguém", () => {
    const viewport = mockViewport({ width: 390 });
    const listener = vi.fn();
    window.matchMedia("(min-width: 768px)").addEventListener("change", listener);
    viewport.restore();

    viewport.resize(1024);

    expect(listener).not.toHaveBeenCalled();
  });
});

// Os testes abaixo dependem da ordem: cada um confere o que o anterior deixou.
describe("restauração automática entre testes", () => {
  it("instala num teste sem restaurar", () => {
    mockViewport({ width: 390, coarse: true });
    expect(hasMatchMedia()).toBe(true);
  });

  it("o teste seguinte encontra o ambiente padrão", () => {
    expect(hasMatchMedia()).toBe(false);
  });

  describe("instalado em beforeEach", () => {
    beforeEach(() => {
      mockViewport({ width: 390 });
    });

    it("vale para o teste", () => {
      expect(matches("(max-width: 767px)")).toBe(true);
    });
  });

  it("e é restaurado ao sair do bloco", () => {
    expect(hasMatchMedia()).toBe(false);
  });
});
