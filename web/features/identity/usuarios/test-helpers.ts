import { screen, within } from "@testing-library/react";
import { http, HttpResponse } from "msw";

import type { components } from "@/lib/api/identity";
import { apiUrl } from "@/lib/session/test-utils";
import { server } from "@/test/msw/server";

// Ajudantes dos testes da administração de usuários.

export type UserPage = components["schemas"]["UserPage"];

/** Consultas recebidas por GET /users, como objeto de parâmetros. */
export type ListRequest = Record<string, string>;

/**
 * Responde GET /users com o que `respond` devolver para os parâmetros da
 * consulta (uma página ou uma Response) e registra cada consulta.
 */
export function mockListUsers(respond: (query: URLSearchParams) => UserPage | Response) {
  const requests: ListRequest[] = [];
  server.use(
    http.get(apiUrl("/api/v1/users"), ({ request }) => {
      const query = new URL(request.url).searchParams;
      requests.push(Object.fromEntries(query.entries()));
      const result = respond(query);
      return result instanceof Response ? result : HttpResponse.json(result);
    }),
  );
  return requests;
}

/** Requisição de escrita recebida: corpo e token CSRF. */
export type WriteRequest = { body: unknown; csrf: string | null };

/** Registra um POST/PUT de escrita e responde com `respond`. */
export function mockWrite(method: "post" | "put", path: string, respond: () => Response) {
  const requests: WriteRequest[] = [];
  server.use(
    http[method](apiUrl(path), async ({ request }) => {
      const text = await request.text();
      requests.push({ body: text ? JSON.parse(text) : null, csrf: request.headers.get("X-CSRF-Token") });
      return respond();
    }),
  );
  return requests;
}

export function noContent() {
  return new HttpResponse(null, { status: 204 });
}

/** Linha da tabela que contém o texto (por exemplo o e-mail). */
export function rowOf(text: string): HTMLElement {
  const cell = screen.getByText(text);
  const row = cell.closest("tr");
  if (!row) throw new Error(`sem linha para ${text}`);
  return row;
}

/** Textos das células de uma linha. */
export function cellsOf(text: string): string[] {
  return within(rowOf(text))
    .getAllByRole("cell")
    .map((cell) => cell.textContent ?? "");
}
