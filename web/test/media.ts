import { act } from "@testing-library/react";
import { onTestFinished } from "vitest";

// Simula largura de tela e tipo de ponteiro nos testes. O ambiente padrão não
// muda: sem chamar mockViewport, o jsdom continua sem `window.matchMedia`, e
// os hooks de mídia tratam essa ausência como desktop.
//
//   const viewport = mockViewport({ width: 390, coarse: true });
//   render(<Lista />);            // layout de tela estreita
//   viewport.resize(1024);        // avisa os ouvintes de `change`
//
// Chame dentro de um teste ou de um `beforeEach`: a restauração é registrada
// para o fim do teste, sem depender de quem chamou.

export type ViewportOptions = {
  /** Largura da tela em px. Padrão: 1024 (desktop). */
  width?: number;
  /** Ponteiro grosso (toque). Padrão: false (ponteiro fino). */
  coarse?: boolean;
};

export type MockedViewport = {
  /** Muda a largura e avisa os ouvintes das consultas cujo resultado mudou. */
  resize: (width: number) => void;
  /** Devolve `window.matchMedia` ao que era antes. Chamada automática ao fim do teste. */
  restore: () => void;
};

type Viewport = { width: number; coarse: boolean };

const FEATURE = /^\(\s*([a-z-]+)\s*:\s*([^)]+?)\s*\)$/;

function featureMatches(feature: string, viewport: Viewport) {
  const [, name, value] = FEATURE.exec(feature.trim()) ?? [];
  if (name === "pointer") return value === (viewport.coarse ? "coarse" : "fine");
  const px = /^(\d+(?:\.\d+)?)px$/.exec(value ?? "");
  if (!px) return false;
  if (name === "min-width") return viewport.width >= Number(px[1]);
  if (name === "max-width") return viewport.width <= Number(px[1]);
  return false;
}

/** Só `(min-width: Npx)`, `(max-width: Npx)`, `(pointer: coarse|fine)` e combinações com `and`. O resto não casa. */
function evaluate(query: string, viewport: Viewport) {
  return query.split(/\s+and\s+/).every((feature) => featureMatches(feature, viewport));
}

function createQueryList(query: string, read: () => boolean): MediaQueryList {
  const list = new EventTarget() as MediaQueryList;
  list.onchange = null;
  list.addListener = (listener) => {
    if (listener) list.addEventListener("change", listener);
  };
  list.removeListener = (listener) => {
    if (listener) list.removeEventListener("change", listener);
  };
  Object.defineProperties(list, {
    media: { value: query, enumerable: true },
    matches: { get: read, enumerable: true },
  });
  return list;
}

function announce(list: MediaQueryList) {
  const event = new Event("change") as MediaQueryListEvent;
  Object.defineProperties(event, {
    media: { value: list.media, enumerable: true },
    matches: { value: list.matches, enumerable: true },
  });
  list.dispatchEvent(event);
  list.onchange?.(event);
}

// Restauração da instalação ativa (no máximo uma por vez).
let restoreActive: (() => void) | null = null;

export function mockViewport({ width = 1024, coarse = false }: ViewportOptions = {}): MockedViewport {
  restoreActive?.();

  const viewport: Viewport = { width, coarse };
  const lists: MediaQueryList[] = [];
  const previous = Object.getOwnPropertyDescriptor(window, "matchMedia");
  let active = true;

  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    writable: true,
    value: (query: string) => {
      const list = createQueryList(query, () => evaluate(query, viewport));
      lists.push(list);
      return list;
    },
  });

  function restore() {
    if (!active) return;
    active = false;
    restoreActive = null;
    lists.length = 0;
    if (previous) Object.defineProperty(window, "matchMedia", previous);
    else Reflect.deleteProperty(window, "matchMedia");
  }

  function resize(next: number) {
    const before = lists.map((list) => list.matches);
    viewport.width = next;
    // Chamado direto pelo teste: `act` aplica a nova renderização antes de voltar.
    act(() => {
      lists.forEach((list, i) => {
        if (list.matches !== before[i]) announce(list);
      });
    });
  }

  restoreActive = restore;
  onTestFinished(restore);
  return { resize, restore };
}
