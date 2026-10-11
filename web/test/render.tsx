import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, type RenderOptions } from "@testing-library/react";
import type { ReactElement, ReactNode } from "react";

export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: Infinity },
      mutations: { retry: false },
    },
  });
}

export type ProvidersOptions = Omit<RenderOptions, "wrapper"> & {
  queryClient?: QueryClient;
  /**
   * Ponto de extensão: envolve a árvore em provedores extras (por exemplo a
   * sessão, acrescentada por F2), por dentro do QueryClientProvider.
   */
  wrap?: (children: ReactNode) => ReactNode;
};

/** Renderiza com um QueryClient novo por chamada, sem repetição de consultas. */
export function renderWithProviders(
  ui: ReactElement,
  { queryClient = createTestQueryClient(), wrap, ...options }: ProvidersOptions = {},
) {
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {wrap ? wrap(children) : children}
      </QueryClientProvider>
    );
  }
  return { queryClient, ...render(ui, { wrapper: Wrapper, ...options }) };
}
