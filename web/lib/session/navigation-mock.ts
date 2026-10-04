import { vi } from "vitest";

// Substituto de `next/navigation` para os testes (fora do App Router não há
// roteador montado). Uso, no topo do arquivo de teste:
//
//   vi.mock("next/navigation", async () =>
//     (await import("@/lib/session/navigation-mock")).nextNavigationMock);
//
// O caminho atual é o do jsdom (`window.location`), ajustado com setLocation.

export const router = {
  push: vi.fn<(href: string) => void>(),
  replace: vi.fn<(href: string) => void>(),
  back: vi.fn(),
  forward: vi.fn(),
  refresh: vi.fn(),
  prefetch: vi.fn(),
};

export const nextNavigationMock = {
  useRouter: () => router,
  usePathname: () => window.location.pathname,
  useSearchParams: () => new URLSearchParams(window.location.search),
};

/** Muda o caminho atual do jsdom (pathname, query e hash). */
export function setLocation(path: string) {
  window.history.replaceState(null, "", path);
}

/** Todas as navegações pedidas (push e replace), na ordem. */
export function navigations(): string[] {
  const calls = [
    ...router.push.mock.calls.map((args, i) => ({ href: args[0], order: router.push.mock.invocationCallOrder[i] })),
    ...router.replace.mock.calls.map((args, i) => ({ href: args[0], order: router.replace.mock.invocationCallOrder[i] })),
  ];
  return calls.sort((a, b) => a.order - b.order).map((c) => c.href);
}

export function resetNavigation() {
  for (const fn of Object.values(router)) fn.mockReset();
  setLocation("/");
}
