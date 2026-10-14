"use client";

import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";

import type { Codec } from "./codecs";

/** Espera da busca antes de escrever na URL (o Safari recusa mais de 100 chamadas de histórico em 30 s). */
export const SEARCH_DEBOUNCE_MS = 300;

export type QueryStateOptions = {
  /** `"replace"` (padrão: filtros e busca) troca a entrada atual; `"push"` (aba) cria uma entrada. */
  history?: "replace" | "push";
  /**
   * Adia a escrita na URL: várias mudanças dentro do prazo viram uma só
   * escrita, com o último valor. O valor devolvido muda na hora. Para busca,
   * use `SEARCH_DEBOUNCE_MS`. Padrão: 0 (escreve na hora).
   */
  debounceMs?: number;
};

// Valor escolhido pela pessoa que a URL lida pelo React ainda não mostra:
// durante a espera (`settledFrom` indefinido) ou entre a escrita e a
// renderização com a URL nova (`settledFrom` = o que a URL tinha antes).
type Draft<T> = { value: T; settledFrom?: string | null };

const currentRaw = (key: string) => new URLSearchParams(window.location.search).get(key);

/**
 * Valor guardado em um parâmetro da query string (DSP-05 AC2 a AC4).
 *
 * Lê por `useSearchParams` e escreve por `window.history.replaceState` ou
 * `pushState`, que o Next integra ao roteador. Cada escrita parte da URL do
 * momento e mexe só na própria chave: os outros parâmetros, o caminho e a
 * âncora ficam como estão.
 *
 * Em página pré-renderizada, quem usa o hook fica dentro de `<Suspense>`
 * (exigência de `useSearchParams`).
 */
export function useQueryState<T>(
  key: string,
  codec: Codec<T>,
  options: QueryStateOptions = {},
): [value: T, setValue: (next: T) => void] {
  const { history = "replace", debounceMs = 0 } = options;
  const raw = useSearchParams().get(key);
  const [draft, setDraft] = useState<Draft<T> | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  // A URL já mudou depois da escrita: ela volta a ser a fonte do valor.
  const settled = draft !== null && draft.settledFrom !== undefined && draft.settledFrom !== raw;
  if (settled) setDraft(null);

  // O codec costuma ser criado a cada renderização; a escrita usa o mais novo.
  const latest = useRef({ key, codec, history, debounceMs });
  useEffect(() => {
    latest.current = { key, codec, history, debounceMs };
  });

  useEffect(() => () => clearTimeout(timer.current), []);

  const setValue = useCallback((next: T) => {
    const { key, codec, history, debounceMs } = latest.current;

    const write = () => {
      const url = new URL(window.location.href);
      const before = url.searchParams.get(key);
      const serialized = codec.serialize(next);
      if (serialized === null) url.searchParams.delete(key);
      else url.searchParams.set(key, serialized);

      // Nada a escrever: não gasta chamada nem cria entrada de histórico.
      if (serialized === before) return setDraft(null);

      const href = url.pathname + url.search + url.hash;
      if (history === "push") window.history.pushState(null, "", href);
      else window.history.replaceState(null, "", href);
      setDraft(currentRaw(key) === before ? null : { value: next, settledFrom: before });
    };

    clearTimeout(timer.current);
    if (debounceMs > 0) {
      setDraft({ value: next });
      timer.current = setTimeout(write, debounceMs);
    } else {
      write();
    }
  }, []);

  return [draft !== null && !settled ? draft.value : codec.parse(raw), setValue];
}
