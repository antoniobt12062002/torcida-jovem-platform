import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, type Conta, conta, problem } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { ContasPage } from "./contas-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

// FWB-01: plano de contas.

const READ = "financeiro:conta:read";
const CREATE = "financeiro:conta:create";
const UPDATE = "financeiro:conta:update";
const DEACTIVATE = "financeiro:conta:deactivate";
const ALL = [READ, CREATE, UPDATE, DEACTIVATE];

const CONTAS = "/api/v1/financeiro/contas";
const CSRF = "csrf-token-de-teste";

const ID_RECEITAS = "00000000-0000-4000-8000-000000000a01";
const ID_MENSALIDADES = "00000000-0000-4000-8000-000000000a02";
const ID_RIFAS = "00000000-0000-4000-8000-000000000a03";
const ID_DESPESAS = "00000000-0000-4000-8000-000000000a04";

function planoInicial(): Conta[] {
  return [
    conta({ id: ID_RECEITAS, tipo: "RECEITA", nome: "Receitas", parent_id: null }),
    conta({ id: ID_MENSALIDADES, tipo: "RECEITA", nome: "Mensalidades", parent_id: ID_RECEITAS }),
    conta({ id: ID_RIFAS, tipo: "RECEITA", nome: "Rifas", parent_id: ID_RECEITAS, ativo: false }),
    conta({ id: ID_DESPESAS, tipo: "DESPESA", nome: "Despesas", parent_id: null }),
  ];
}

type Captured = { body: unknown; csrf: string | null };

/** Estado do servidor: GET /contas devolve sempre a lista atual. */
function mockServer(items: Conta[] = planoInicial()) {
  const state = { items, gets: 0 };
  server.use(
    http.get(apiUrl(CONTAS), () => {
      state.gets += 1;
      return HttpResponse.json({ items: state.items });
    }),
  );
  return state;
}

async function readBody(request: Request): Promise<unknown> {
  const text = await request.text();
  return text === "" ? null : JSON.parse(text);
}

async function renderPage(permissions: string[] = ALL) {
  mockMe(authContext({ permissions }));
  renderWithSession(<ContasPage />);
  await screen.findByRole("heading", { name: "Plano de contas" });
}

async function renderLoaded(permissions: string[] = ALL) {
  await renderPage(permissions);
  await screen.findByText("Receitas");
}

/** Item da árvore (li) da conta, com as subcontas dentro. */
function item(nome: string): HTMLElement {
  const li = screen.getByText(nome, { selector: "[data-slot=conta-nome]" }).closest("li");
  if (!li) throw new Error(`conta ${nome} fora de um item da árvore`);
  return li;
}

/** Linha da conta (sem as subcontas). */
function row(nome: string): HTMLElement {
  const el = screen.getByText(nome, { selector: "[data-slot=conta-nome]" }).closest("[data-slot=conta-row]");
  if (!el) throw new Error(`linha da conta ${nome} não encontrada`);
  return el as HTMLElement;
}

function button(name: string) {
  return screen.getByRole("button", { name });
}

function type(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}

// FWB-01 AC1.
describe("árvore do plano de contas", () => {
  it("mostra as contas em árvore por parent_id, com tipo e situação", async () => {
    const state = mockServer();
    await renderLoaded([READ]);

    expect(state.gets).toBe(1);
    const receitas = item("Receitas");
    expect(within(receitas).getByText("Mensalidades")).toBeTruthy();
    expect(within(receitas).getByText("Rifas")).toBeTruthy();
    expect(within(item("Despesas")).queryByText("Mensalidades")).toBeNull();

    expect(within(row("Receitas")).getByText("Receita")).toBeTruthy();
    expect(within(row("Receitas")).getByText("Ativo")).toBeTruthy();
    expect(within(row("Despesas")).getByText("Despesa")).toBeTruthy();
    expect(within(row("Mensalidades")).getByText("Ativo")).toBeTruthy();
    // Inativa visível e marcada.
    expect(within(row("Rifas")).getByText("Inativo")).toBeTruthy();
    expect(within(row("Rifas")).queryByText("Ativo")).toBeNull();
  });

  it("conta cujo pai não veio na lista aparece na raiz", async () => {
    mockServer([
      conta({ id: ID_RECEITAS, nome: "Receitas" }),
      conta({ id: ID_RIFAS, nome: "Órfã", parent_id: "00000000-0000-4000-8000-00000000ffff" }),
    ]);
    await renderLoaded([READ]);
    expect(within(item("Receitas")).queryByText("Órfã")).toBeNull();
    expect(screen.getByText("Órfã")).toBeTruthy();
  });

  it("sem contas, informa que o plano está vazio", async () => {
    mockServer([]);
    await renderPage([READ]);
    expect(await screen.findByText("Nenhuma conta cadastrada.")).toBeTruthy();
  });

  it("falha ao carregar mostra o erro e tenta de novo", async () => {
    let fail = true;
    server.use(
      http.get(apiUrl(CONTAS), () =>
        fail ? problem(503, "service_unavailable") : HttpResponse.json({ items: planoInicial() }),
      ),
    );
    await renderPage([READ]);
    expect(
      await screen.findByText("O serviço está indisponível no momento. Tente de novo em instantes."),
    ).toBeTruthy();
    fail = false;
    fireEvent.click(button("Tentar de novo"));
    expect(await screen.findByText("Receitas")).toBeTruthy();
  });
});

// FWB-01 AC2 e AC3.
describe("criar conta", () => {
  function mockCreate(state: { items: Conta[] }, respond?: () => Response) {
    const requests: Captured[] = [];
    server.use(
      http.post(apiUrl(CONTAS), async ({ request }) => {
        const body = (await readBody(request)) as { tipo: Conta["tipo"]; nome: string; parent_id?: string };
        requests.push({ body, csrf: request.headers.get("X-CSRF-Token") });
        if (respond) return respond();
        const created = conta({
          id: "00000000-0000-4000-8000-000000000b01",
          tipo: body.tipo,
          nome: body.nome,
          parent_id: body.parent_id ?? null,
        });
        state.items = [...state.items, created];
        return HttpResponse.json(created, { status: 201 });
      }),
    );
    return requests;
  }

  it("cria conta raiz com tipo e nome e a mostra na árvore", async () => {
    const state = mockServer();
    const requests = mockCreate(state);
    await renderLoaded();

    fireEvent.click(button("Nova conta"));
    const dialog = await screen.findByRole("dialog", { name: "Nova conta" });
    type("Nome", "Patrocínios");
    fireEvent.click(within(dialog).getByRole("radio", { name: "Despesa" }));
    fireEvent.click(within(dialog).getByRole("button", { name: "Criar conta" }));

    await waitFor(() => expect(dialog.isConnected).toBe(false));
    expect(requests).toEqual([{ body: { tipo: "DESPESA", nome: "Patrocínios" }, csrf: CSRF }]);
    expect(within(row("Patrocínios")).getByText("Despesa")).toBeTruthy();
    expect(within(item("Despesas")).queryByText("Patrocínios")).toBeNull();
  });

  it("subconta fixa o tipo do pai e envia parent_id", async () => {
    const state = mockServer();
    const requests = mockCreate(state);
    await renderLoaded();

    fireEvent.click(button("Nova subconta em Despesas"));
    const dialog = await screen.findByRole("dialog", { name: "Nova subconta em Despesas" });
    const despesa = within(dialog).getByRole("radio", { name: "Despesa" }) as HTMLInputElement;
    const receita = within(dialog).getByRole("radio", { name: "Receita" }) as HTMLInputElement;
    expect(despesa.checked).toBe(true);
    expect(despesa.disabled).toBe(true);
    expect(receita.checked).toBe(false);
    expect(receita.disabled).toBe(true);

    type("Nome", "Transporte");
    fireEvent.click(within(dialog).getByRole("button", { name: "Criar conta" }));

    await waitFor(() => expect(dialog.isConnected).toBe(false));
    expect(requests).toEqual([
      { body: { tipo: "DESPESA", nome: "Transporte", parent_id: ID_DESPESAS }, csrf: CSRF },
    ]);
    expect(within(item("Despesas")).getByText("Transporte")).toBeTruthy();
  });

  it("sem nome ou sem tipo, não chama a API", async () => {
    const state = mockServer();
    const requests = mockCreate(state);
    await renderLoaded();

    fireEvent.click(button("Nova conta"));
    const dialog = await screen.findByRole("dialog", { name: "Nova conta" });
    type("Nome", "   ");
    fireEvent.click(within(dialog).getByRole("button", { name: "Criar conta" }));

    expect(await within(dialog).findByText("Informe o nome.")).toBeTruthy();
    expect(within(dialog).getByText("Escolha o tipo.")).toBeTruthy();
    expect(requests).toEqual([]);
  });

  // FWB-01 AC6 e FND-04 AC5.
  it.each([
    [422, "conta_tipo_incompativel", "A subconta precisa ter o mesmo tipo da conta pai."],
    [404, "conta_nao_encontrada", "Conta não encontrada. Atualize a página e tente de novo."],
    [403, "forbidden", "Você não tem permissão para esta ação."],
  ])("%i %s mostra a mensagem e não altera a árvore", async (status, code, message) => {
    const state = mockServer();
    mockCreate(state, () => problem(status, code));
    await renderLoaded();

    fireEvent.click(button("Nova subconta em Receitas"));
    const dialog = await screen.findByRole("dialog", { name: "Nova subconta em Receitas" });
    type("Nome", "Bilheteria");
    fireEvent.click(within(dialog).getByRole("button", { name: "Criar conta" }));

    expect(await within(dialog).findByText(message)).toBeTruthy();
    expect(screen.queryByText("Bilheteria", { selector: "[data-slot=conta-nome]" })).toBeNull();
  });
});

// FWB-01 AC4.
describe("renomear conta", () => {
  function mockRename(state: { items: Conta[] }, respond?: () => Response) {
    const requests: (Captured & { id: string })[] = [];
    server.use(
      http.patch(apiUrl(`${CONTAS}/:id`), async ({ request, params }) => {
        const body = (await readBody(request)) as { nome: string };
        const id = String(params.id);
        requests.push({ id, body, csrf: request.headers.get("X-CSRF-Token") });
        if (respond) return respond();
        state.items = state.items.map((c) => (c.id === id ? { ...c, nome: body.nome } : c));
        return HttpResponse.json(state.items.find((c) => c.id === id));
      }),
    );
    return requests;
  }

  it("envia o novo nome e mostra a conta renomeada", async () => {
    const state = mockServer();
    const requests = mockRename(state);
    await renderLoaded();

    fireEvent.click(button("Renomear Mensalidades"));
    const dialog = await screen.findByRole("dialog", { name: "Renomear conta" });
    expect((screen.getByLabelText("Nome") as HTMLInputElement).value).toBe("Mensalidades");
    type("Nome", "Mensalidades 2026");
    fireEvent.click(within(dialog).getByRole("button", { name: "Salvar" }));

    await waitFor(() => expect(dialog.isConnected).toBe(false));
    expect(requests).toEqual([
      { id: ID_MENSALIDADES, body: { nome: "Mensalidades 2026" }, csrf: CSRF },
    ]);
    expect(within(item("Receitas")).getByText("Mensalidades 2026")).toBeTruthy();
  });

  // FWB-01 AC4 (409) e AC6.
  it.each([
    [409, "conta_ja_utilizada", "Esta conta já foi usada em um lançamento e não pode ser renomeada."],
    [404, "conta_nao_encontrada", "Conta não encontrada. Atualize a página e tente de novo."],
  ])("%i %s explica e mantém o nome", async (status, code, message) => {
    const state = mockServer();
    mockRename(state, () => problem(status, code));
    await renderLoaded();

    fireEvent.click(button("Renomear Mensalidades"));
    const dialog = await screen.findByRole("dialog", { name: "Renomear conta" });
    type("Nome", "Outro nome");
    fireEvent.click(within(dialog).getByRole("button", { name: "Salvar" }));

    expect(await within(dialog).findByText(message)).toBeTruthy();
    expect(screen.getByText("Mensalidades", { selector: "[data-slot=conta-nome]" })).toBeTruthy();
    expect(screen.queryByText("Outro nome", { selector: "[data-slot=conta-nome]" })).toBeNull();
  });
});

// FWB-01 AC5.
describe("desativar conta", () => {
  function mockDeactivate(state: { items: Conta[] }, respond?: () => Response) {
    const requests: (Captured & { id: string })[] = [];
    server.use(
      http.post(apiUrl(`${CONTAS}/:id/deactivate`), async ({ request, params }) => {
        const id = String(params.id);
        requests.push({ id, body: await readBody(request), csrf: request.headers.get("X-CSRF-Token") });
        if (respond) return respond();
        state.items = state.items.map((c) => (c.id === id ? { ...c, ativo: false } : c));
        return new HttpResponse(null, { status: 204 });
      }),
    );
    return requests;
  }

  it("só chama a API depois de confirmar e mostra a conta como inativa", async () => {
    const state = mockServer();
    const requests = mockDeactivate(state);
    await renderLoaded();

    fireEvent.click(button("Desativar Mensalidades"));
    await screen.findByRole("alertdialog", { name: "Desativar a conta Mensalidades?" });
    fireEvent.click(button("Cancelar"));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(requests).toEqual([]);

    fireEvent.click(button("Desativar Mensalidades"));
    const dialog = await screen.findByRole("alertdialog", { name: "Desativar a conta Mensalidades?" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Desativar" }));

    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(requests).toEqual([{ id: ID_MENSALIDADES, body: null, csrf: CSRF }]);
    await waitFor(() => expect(within(row("Mensalidades")).getByText("Inativo")).toBeTruthy());
    // Continua na árvore, sob o pai.
    expect(within(item("Receitas")).getByText("Mensalidades")).toBeTruthy();
  });

  // FWB-01 AC6 e FND-04 AC5.
  it.each([
    [404, "conta_nao_encontrada", "Conta não encontrada. Atualize a página e tente de novo."],
    [403, "forbidden", "Você não tem permissão para esta ação."],
  ])("%i %s mostra a mensagem e mantém a conta ativa", async (status, code, message) => {
    const state = mockServer();
    mockDeactivate(state, () => problem(status, code));
    await renderLoaded();

    fireEvent.click(button("Desativar Mensalidades"));
    const dialog = await screen.findByRole("alertdialog", { name: "Desativar a conta Mensalidades?" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Desativar" }));

    expect(await screen.findByText(message)).toBeTruthy();
    expect(within(row("Mensalidades")).getByText("Ativo")).toBeTruthy();
  });
});

// FWB-01 AC2, AC4, AC5 e caso de borda de só leitura: cada ação só com a sua permissão.
describe("ações por permissão", () => {
  const names = {
    criar: "Nova conta",
    subconta: "Nova subconta em Receitas",
    renomear: "Renomear Mensalidades",
    desativar: "Desativar Mensalidades",
  };

  function visible() {
    return Object.fromEntries(
      Object.entries(names).map(([key, name]) => [key, screen.queryByRole("button", { name }) !== null]),
    );
  }

  it("só leitura: nenhuma ação de escrita", async () => {
    mockServer();
    await renderLoaded([READ]);
    expect(visible()).toEqual({ criar: false, subconta: false, renomear: false, desativar: false });
    expect(screen.queryAllByRole("button", { name: /^(Nova|Renomear|Desativar)/ })).toEqual([]);
  });

  it.each([
    [CREATE, { criar: true, subconta: true, renomear: false, desativar: false }],
    [UPDATE, { criar: false, subconta: false, renomear: true, desativar: false }],
    [DEACTIVATE, { criar: false, subconta: false, renomear: false, desativar: true }],
  ])("%s mostra só a ação correspondente", async (permission, expected) => {
    mockServer();
    await renderLoaded([READ, permission]);
    expect(visible()).toEqual(expected);
  });

  it("conta já inativa não oferece desativar", async () => {
    mockServer();
    await renderLoaded();
    expect(screen.queryByRole("button", { name: "Desativar Rifas" })).toBeNull();
    expect(button("Renomear Rifas")).toBeTruthy();
  });
});
