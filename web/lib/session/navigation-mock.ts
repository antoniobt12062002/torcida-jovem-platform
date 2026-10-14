import { act } from "@testing-library/react";
import { useSyncExternalStore } from "react";
import { vi } from "vitest";

// Substituto de `next/navigation` para os testes (fora do App Router não há
// roteador montado). Uso, no topo do arquivo de teste:
//
//   vi.mock("next/navigation", async () =>
//     (await import("@/lib/session/navigation-mock")).nextNavigationMock);
//
// O caminho atual é o do jsdom (`window.location`), ajustado com setLocation.
// Como no Next, `usePathname` e `useSearchParams` acompanham
// `history.pushState`, `history.replaceState` e a volta no histórico: quem
// escreve na URL é renderizado de novo com o valor novo.

export const router = {
  push: vi.fn<(href: string) => void>(),
  replace: vi.fn<(href: string) => void>(),
  back: vi.fn(),
  forward: vi.fn(),
  refresh: vi.fn(),
  prefetch: vi.fn(),
};

const HISTORY_METHODS = ["pushState", "replaceState"] as const;
type HistoryMethod = (typeof HISTORY_METHODS)[number];

const listeners = new Set<() => void>();
// Interceptações instaladas no histórico do jsdom, só enquanto há assinante.
const intercepted = new Map<HistoryMethod, { wrapper: History[HistoryMethod]; restore: () => void }>();

function notify() {
  for (const listener of [...listeners]) listener();
}

function watchHistory() {
  if (intercepted.size > 0) return;
  for (const method of HISTORY_METHODS) {
    const own = Object.getOwnPropertyDescriptor(window.history, method);
    const original = window.history[method];
    const wrapper: History[HistoryMethod] = (...args) => {
      original.apply(window.history, args);
      notify();
    };
    window.history[method] = wrapper;
    intercepted.set(method, {
      wrapper,
      restore: () => {
        if (own) Object.defineProperty(window.history, method, own);
        else Reflect.deleteProperty(window.history, method);
      },
    });
  }
  window.addEventListener("popstate", notify);
}

function unwatchHistory() {
  for (const [method, { wrapper, restore }] of intercepted) {
    // Se um teste espionou o método por cima, quem desfaz é o próprio espião.
    if (window.history[method] === wrapper) restore();
  }
  intercepted.clear();
  window.removeEventListener("popstate", notify);
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  watchHistory();
  return () => {
    listeners.delete(listener);
    if (listeners.size === 0) unwatchHistory();
  };
}

const getPathname = () => window.location.pathname;

// O mesmo objeto enquanto a query não muda, como no Next: um efeito que
// depende de `useSearchParams` não dispara a cada renderização.
let searchParams = { search: "", value: new URLSearchParams() };

function getSearchParams() {
  const { search } = window.location;
  if (search !== searchParams.search) searchParams = { search, value: new URLSearchParams(search) };
  return searchParams.value;
}

export const nextNavigationMock = {
  useRouter: () => router,
  usePathname: () => useSyncExternalStore(subscribe, getPathname, getPathname),
  useSearchParams: () => useSyncExternalStore(subscribe, getSearchParams, getSearchParams),
};

/** Muda o caminho atual do jsdom (pathname, query e hash) e avisa os componentes montados. */
export function setLocation(path: string) {
  // Chamado direto pelo teste: `act` aplica a nova renderização antes de voltar.
  act(() => window.history.replaceState(null, "", path));
}

/** Todas as navegações pedidas (push e replace), na ordem. */
export function navigations(): string[] {
  const calls = [
    ...router.push.mock.calls.map((args, i) => ({ href: args[0], order: router.push.mock.invocationCallOrder[i] })),
    ...router.replace.mock.calls.map((args, i) => ({ href: args[0], order: router.replace.mock.invocationCallOrder[i] })),
  ];
  return calls.sort((a, b) => a.order - b.order).map((c) => c.href);
}

/** Limpa chamadas do roteador e assinaturas do histórico, e volta a `/`. */
export function resetNavigation() {
  for (const fn of Object.values(router)) fn.mockReset();
  listeners.clear();
  unwatchHistory();
  setLocation("/");
}
