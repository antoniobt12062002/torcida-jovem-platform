"use client";

import { useCallback, useSyncExternalStore } from "react";

import { BREAKPOINT_MD } from "./breakpoints";

/**
 * A tela de referência: o que o servidor assume, e o que vale onde não há
 * `matchMedia`. Uma tela larga com mouse.
 */
export const DESKTOP = { width: 1280, pointer: "fine" } as const;

const FEATURE = /^\(\s*([a-z-]+)\s*:\s*([^)]+?)\s*\)$/;

function featureMatchesOnDesktop(feature: string) {
  const [, name, value] = FEATURE.exec(feature.trim()) ?? [];
  if (name === "pointer") return value === DESKTOP.pointer;
  const px = /^(\d+(?:\.\d+)?)px$/.exec(value ?? "");
  if (!px) return false;
  if (name === "min-width") return DESKTOP.width >= Number(px[1]);
  if (name === "max-width") return DESKTOP.width <= Number(px[1]);
  return false;
}

/**
 * Resultado da consulta na tela de referência. Entende `(min-width: Npx)`,
 * `(max-width: Npx)`, `(pointer: coarse|fine)` e combinações com `and`.
 * Qualquer outra consulta não casa.
 */
function matchesOnDesktop(query: string) {
  return query.split(/\s+and\s+/).every(featureMatchesOnDesktop);
}

const hasMatchMedia = () => typeof window.matchMedia === "function";

/**
 * Resultado de uma consulta de mídia, atualizado quando a tela muda.
 *
 * Seguro para SSR (DD-2): no servidor e na primeira renderização da hidratação
 * devolve sempre o resultado da tela `DESKTOP`, de modo que o HTML do servidor
 * e o do cliente coincidem; logo depois de hidratar, passa ao resultado real.
 * Numa renderização só de cliente, lê a tela já na primeira renderização. Sem
 * `matchMedia`, devolve o resultado de `DESKTOP`.
 *
 * Reservado à troca entre tabela e lista; o resto do responsivo é CSS.
 */
export function useMediaQuery(query: string): boolean {
  const subscribe = useCallback(
    (onChange: () => void) => {
      if (!hasMatchMedia()) return () => {};
      const list = window.matchMedia(query);
      list.addEventListener("change", onChange);
      return () => list.removeEventListener("change", onChange);
    },
    [query],
  );

  return useSyncExternalStore(
    subscribe,
    () => (hasMatchMedia() ? window.matchMedia(query).matches : matchesOnDesktop(query)),
    () => matchesOnDesktop(query),
  );
}

/** Tela abaixo de 768 px: a listagem vira lista empilhada. No servidor e sem `matchMedia`: `false`. */
export function useIsNarrow(): boolean {
  return !useMediaQuery(`(min-width: ${BREAKPOINT_MD}px)`);
}

/** Ponteiro principal de toque. No servidor e sem `matchMedia`: `false`. */
export function useIsCoarsePointer(): boolean {
  return useMediaQuery("(pointer: coarse)");
}
