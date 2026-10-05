import { fireEvent, screen, within } from "@testing-library/react";
import { http, HttpResponse } from "msw";

import { apiUrl } from "@/lib/session/test-utils";
import { type Conta, conta, type Lancamento } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

// Ajudantes dos testes de lançamentos (e comprovantes).

export const PERMISSOES_TESOURARIA = [
  "financeiro:conta:read",
  "financeiro:lancamento:read",
  "financeiro:lancamento:create",
  "financeiro:lancamento:update",
  "financeiro:lancamento:receive",
  "financeiro:lancamento:pay",
  "financeiro:lancamento:cancel",
  "financeiro:comprovante:read",
  "financeiro:comprovante:create",
];

export const PERMISSOES_LEITURA = [
  "financeiro:conta:read",
  "financeiro:lancamento:read",
  "financeiro:comprovante:read",
];

export const CONTA_RECEITA = conta({ id: "00000000-0000-4000-8000-000000000111", tipo: "RECEITA", nome: "Mensalidades" });
export const CONTA_RECEITA_INATIVA = conta({
  id: "00000000-0000-4000-8000-000000000112",
  tipo: "RECEITA",
  nome: "Rifas antigas",
  ativo: false,
});
export const CONTA_DESPESA = conta({ id: "00000000-0000-4000-8000-000000000121", tipo: "DESPESA", nome: "Devoluções" });
export const CONTA_DESPESA_2 = conta({ id: "00000000-0000-4000-8000-000000000122", tipo: "DESPESA", nome: "Transporte" });
export const CONTAS = [CONTA_RECEITA, CONTA_RECEITA_INATIVA, CONTA_DESPESA, CONTA_DESPESA_2];

const BASE = "/api/v1/financeiro";

/** GET /lancamentos devolve `current()` a cada chamada; conta as chamadas. */
export function mockLancamentos(items: Lancamento[] | (() => Lancamento[])) {
  const calls = { count: 0 };
  const current = typeof items === "function" ? items : () => items;
  server.use(
    http.get(apiUrl(`${BASE}/lancamentos`), () => {
      calls.count += 1;
      return HttpResponse.json({ items: current() });
    }),
  );
  return calls;
}

export function mockContas(items: Conta[] = CONTAS) {
  const calls = { count: 0 };
  server.use(
    http.get(apiUrl(`${BASE}/contas`), () => {
      calls.count += 1;
      return HttpResponse.json({ items });
    }),
  );
  return calls;
}

export type Recorded = { body: unknown; csrf: string | null };

/** Registra as escritas (método e caminho) com corpo JSON e cabeçalho CSRF. */
export function mockWrite(
  method: "post" | "put",
  path: string,
  respond: () => Response,
): Recorded[] {
  const requests: Recorded[] = [];
  server.use(
    http[method](apiUrl(`${BASE}${path}`), async ({ request }) => {
      const text = await request.text();
      requests.push({ body: text ? JSON.parse(text) : null, csrf: request.headers.get("X-CSRF-Token") });
      return respond();
    }),
  );
  return requests;
}

/** O diálogo de formulário aberto (os avisos também têm papel dialog). */
export function formDialog(): HTMLElement | null {
  return document.querySelector<HTMLElement>('[data-slot="dialog-content"]');
}

export function dialog() {
  const element = formDialog();
  if (!element) throw new Error("nenhum diálogo de formulário aberto");
  return within(element);
}

export function change(element: HTMLElement, value: string) {
  fireEvent.change(element, { target: { value } });
}

/** Textos ligados ao campo por aria-describedby. */
export function describedByText(el: HTMLElement): string[] {
  const ids = (el.getAttribute("aria-describedby") ?? "").split(" ").filter(Boolean);
  return ids.map((id) => document.getElementById(id)?.textContent ?? "");
}

/** Valor exibido para um rótulo da lista de definições do detalhe. */
export function detailValue(label: string): string | null {
  const term = screen.getByText(label, { selector: "dt" });
  return term.nextElementSibling?.textContent ?? null;
}
