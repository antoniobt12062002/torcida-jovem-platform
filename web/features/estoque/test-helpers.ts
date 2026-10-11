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
/** Escritas do módulo. */
export const WRITES = [PERM.produtoCreate, PERM.movCreate, PERM.movAdjust];

/** Todas as permissões do estoque, menos as dadas (negativo cruzado, FND-04 AC3). */
export function allExcept(...excluded: string[]): string[] {
  return ALL.filter((permission) => !excluded.includes(permission));
}

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

type MovBody = {
  tipo: "ENTRADA" | "SAIDA" | "DEVOLUCAO";
  produto_id: string;
  quantidade: number;
  origem: Movimentacao["origem"];
  movimentacao_de_id?: string;
};
type AjusteBody = { produto_id: string; quantidade: number; motivo: string };

/**
 * Estoque falso de um produto: histórico, saldo (soma das movimentações) e
 * os POST de movimentação e ajuste, com as regras de sinal e de saldo
 * insuficiente da API. Registra corpo e CSRF de cada POST.
 */
export function mockEstoque(initial: Movimentacao[] = []) {
  const state = {
    items: [...initial],
    saldoCalls: 0,
    movCalls: 0,
    movRequests: [] as Recorded[],
    ajusteRequests: [] as Recorded[],
    nextId: 900,
  };
  const saldo = () => state.items.reduce((sum, m) => sum + m.quantidade, 0);
  const newId = () => `00000000-0000-4000-8000-000000000${state.nextId++}`;
  server.use(
    http.get(apiUrl("/api/v1/estoque/produtos/:id/movimentacoes"), () => {
      state.movCalls += 1;
      return HttpResponse.json({ items: state.items });
    }),
    http.get(apiUrl("/api/v1/estoque/produtos/:id/saldo"), () => {
      state.saldoCalls += 1;
      return HttpResponse.json({ saldo: saldo() });
    }),
    http.post(apiUrl("/api/v1/estoque/movimentacoes"), async ({ request }) => {
      const body = (await request.json()) as MovBody;
      state.movRequests.push({ body, csrf: request.headers.get("X-CSRF-Token") });
      let signed = body.tipo === "SAIDA" ? -body.quantidade : body.quantidade;
      if (body.tipo === "DEVOLUCAO") {
        const ref = state.items.find((m) => m.id === body.movimentacao_de_id);
        if (!ref || (ref.tipo !== "ENTRADA" && ref.tipo !== "SAIDA")) {
          return HttpResponse.json(
            { type: "about:blank", title: "x", status: 422, code: "devolucao_invalida" },
            { status: 422, headers: { "Content-Type": "application/problem+json" } },
          );
        }
        signed = ref.tipo === "SAIDA" ? body.quantidade : -body.quantidade;
      }
      if (signed < 0 && saldo() + signed < 0) {
        return HttpResponse.json(
          { type: "about:blank", title: "x", status: 409, code: "saldo_insuficiente" },
          { status: 409, headers: { "Content-Type": "application/problem+json" } },
        );
      }
      const created: Movimentacao = {
        id: newId(),
        produto_id: body.produto_id,
        tipo: body.tipo,
        quantidade: signed,
        origem: body.origem,
        motivo: null,
        movimentacao_de_id: body.movimentacao_de_id ?? null,
        responsavel_id: "00000000-0000-4000-8000-000000000001",
        criado_em: "2026-10-04T15:00:00Z",
      };
      state.items = [...state.items, created];
      return HttpResponse.json(created, { status: 201 });
    }),
    http.post(apiUrl("/api/v1/estoque/ajustes"), async ({ request }) => {
      const body = (await request.json()) as AjusteBody;
      state.ajusteRequests.push({ body, csrf: request.headers.get("X-CSRF-Token") });
      const created: Movimentacao = {
        id: newId(),
        produto_id: body.produto_id,
        tipo: "AJUSTE",
        quantidade: body.quantidade,
        origem: "AJUSTE_MANUAL",
        motivo: body.motivo,
        movimentacao_de_id: null,
        responsavel_id: "00000000-0000-4000-8000-000000000001",
        criado_em: "2026-10-04T15:00:00Z",
      };
      state.items = [...state.items, created];
      return HttpResponse.json(created, { status: 201 });
    }),
  );
  return state;
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
