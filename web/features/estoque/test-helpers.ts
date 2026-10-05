import { fireEvent, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";

import { apiUrl } from "@/lib/session/test-utils";
import type { Movimentacao, Produto } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

// Ajudantes dos testes do estoque.

export const PERM = {
  produtoRead: "estoque:produto:read",
  produtoCreate: "estoque:produto:create",
  movRead: "estoque:movimentacao:read",
  movCreate: "estoque:movimentacao:create",
  movAdjust: "estoque:movimentacao:adjust",
  saldoRead: "estoque:saldo:read",
} as const;

/** Somente leitura (diretoria, conselho fiscal). */
export const READ_ONLY = [PERM.produtoRead, PERM.movRead, PERM.saldoRead];
/** Responsável de estoque (ESTOQUE_LOJA). */
export const ALL = Object.values(PERM);

export type Recorded = { body: unknown; csrf: string | null };

/** GET /produtos com a lista dada; devolve o contador de chamadas. */
export function mockProdutos(items: Produto[]) {
  const calls = { count: 0 };
  server.use(
    http.get(apiUrl("/api/v1/estoque/produtos"), () => {
      calls.count += 1;
      return HttpResponse.json({ items });
    }),
  );
  return calls;
}

/** GET /produtos/{id}/saldo com saldos que podem mudar; devolve o estado. */
export function mockSaldo(initial: number) {
  const state = { saldo: initial, calls: 0 };
  server.use(
    http.get(apiUrl("/api/v1/estoque/produtos/:id/saldo"), () => {
      state.calls += 1;
      return HttpResponse.json({ saldo: state.saldo });
    }),
  );
  return state;
}

/** GET /produtos/{id}/movimentacoes com a lista (mutável) dada. */
export function mockMovimentacoes(initial: Movimentacao[]) {
  const state = { items: [...initial], calls: 0 };
  server.use(
    http.get(apiUrl("/api/v1/estoque/produtos/:id/movimentacoes"), () => {
      state.calls += 1;
      return HttpResponse.json({ items: state.items });
    }),
  );
  return state;
}

/** Registra corpo e CSRF de cada POST no caminho e responde com `respond`. */
export function recordPost(path: string, respond: (body: unknown) => Response) {
  const requests: Recorded[] = [];
  server.use(
    http.post(apiUrl(path), async ({ request }) => {
      const body = await request.json();
      requests.push({ body, csrf: request.headers.get("X-CSRF-Token") });
      return respond(body);
    }),
  );
  return requests;
}

export function describedByText(el: HTMLElement): string[] {
  const ids = (el.getAttribute("aria-describedby") ?? "").split(" ").filter(Boolean);
  return ids.map((id) => document.getElementById(id)?.textContent ?? "");
}

export function isInvalid(el: HTMLElement): boolean {
  return el.getAttribute("aria-invalid") === "true";
}

export function type(label: string, value: string) {
  const el = screen.getByLabelText(label);
  fireEvent.change(el, { target: { value } });
  fireEvent.blur(el);
}
