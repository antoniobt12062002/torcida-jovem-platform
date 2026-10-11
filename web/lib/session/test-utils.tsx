import { http, HttpResponse } from "msw";
import type { ReactElement } from "react";

import { Toaster } from "@/components/ui/toast";
import { type AuthContext, problem } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";
import { type ProvidersOptions, renderWithProviders } from "@/test/render";

import { SessionProvider } from "./session-provider";

// Ajudantes de teste da sessão (F2). Os testes que os usam precisam também do
// substituto de next/navigation (ver navigation-mock.ts).

export function apiUrl(path: string) {
  return `${window.location.origin}${path}`;
}

/**
 * Responde GET /auth/me com o contexto dado (`null` = sem sessão, 401) e
 * devolve o contador de chamadas.
 */
export function mockMe(context: AuthContext | null) {
  const calls = { count: 0 };
  server.use(
    http.get(apiUrl("/api/v1/auth/me"), () => {
      calls.count += 1;
      return context ? HttpResponse.json(context) : problem(401, "unauthenticated");
    }),
  );
  return calls;
}

/** renderWithProviders com o SessionProvider e o Toaster da raiz. */
export function renderWithSession(ui: ReactElement, { wrap, ...options }: ProvidersOptions = {}) {
  return renderWithProviders(ui, {
    ...options,
    wrap: (children) => (
      <SessionProvider>
        {wrap ? wrap(children) : children}
        <Toaster />
      </SessionProvider>
    ),
  });
}
