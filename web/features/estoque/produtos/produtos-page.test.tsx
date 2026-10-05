import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, type Produto, problem, produto } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import {
  ALL,
  describedByText,
  isInvalid,
  mockProdutos,
  PERM,
  READ_ONLY,
  type Recorded,
  type,
} from "../test-helpers";
import { ProdutosPage } from "./produtos-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const camisa = produto({ id: "00000000-0000-4000-8000-000000000411", codigo: "CAM-001", nome: "Camisa oficial", unidade_medida: "UN" });
const faixa = produto({ id: "00000000-0000-4000-8000-000000000412", codigo: "FAI-002", nome: "Faixa da torcida", unidade_medida: "M" });

async function renderPage(permissions: string[], items: Produto[] = [camisa, faixa]) {
  mockMe(authContext({ permissions }));
  const calls = mockProdutos(items);
  renderWithSession(<ProdutosPage />);
  return calls;
}

function rows() {
  return screen.getAllByRole("row").slice(1);
}

function rowText(row: HTMLElement) {
  return within(row)
    .getAllByRole("cell")
    .map((c) => c.textContent);
}

/** POST /produtos com lista mutável: o produto criado entra na lista. */
function mockCreate(items: Produto[], respond?: () => Response) {
  const requests: Recorded[] = [];
  const state = { items: [...items] };
  server.use(
    http.get(apiUrl("/api/v1/estoque/produtos"), () => HttpResponse.json({ items: state.items })),
    http.post(apiUrl("/api/v1/estoque/produtos"), async ({ request }) => {
      const body = (await request.json()) as { codigo: string; nome: string; unidade_medida: string };
      requests.push({ body, csrf: request.headers.get("X-CSRF-Token") });
      if (respond) return respond();
      const created = produto({ id: "00000000-0000-4000-8000-000000000499", ...body });
      state.items = [...state.items, created];
      return HttpResponse.json(created, { status: 201 });
    }),
  );
  return requests;
}

async function openCreate() {
  fireEvent.click(await screen.findByRole("button", { name: "Novo produto" }));
  await screen.findByRole("dialog", { name: "Novo produto" });
}

function fillCreate(codigo: string, nome: string, unidade: string) {
  type("Código", codigo);
  type("Nome", nome);
  type("Unidade de medida", unidade);
}

function submitCreate() {
  fireEvent.click(screen.getByRole("button", { name: "Criar produto" }));
}

// EWB-01 AC1.
describe("listagem de produtos", () => {
  it("chama GET /produtos e lista código, nome e unidade, sem saldo", async () => {
    const calls = await renderPage([PERM.produtoRead]);
    await screen.findByText("Camisa oficial");
    expect(calls.count).toBe(1);
    const headers = screen.getAllByRole("columnheader").map((h) => h.textContent);
    expect(headers).toEqual(["Código", "Nome", "Unidade de medida"]);
    expect(rows().map(rowText)).toEqual([
      ["CAM-001", "Camisa oficial", "UN"],
      ["FAI-002", "Faixa da torcida", "M"],
    ]);
  });

  it("cada produto leva ao seu detalhe", async () => {
    await renderPage([PERM.produtoRead]);
    const link = await screen.findByRole("link", { name: "Faixa da torcida" });
    expect(link.getAttribute("href")).toBe(`/estoque/produtos/${faixa.id}`);
  });

  it("busca por código ou por nome, no cliente", async () => {
    const calls = await renderPage([PERM.produtoRead]);
    await screen.findByText("Camisa oficial");
    const search = screen.getByLabelText("Buscar por código ou nome");

    fireEvent.change(search, { target: { value: "fai-002" } });
    expect(rows().map(rowText)).toEqual([["FAI-002", "Faixa da torcida", "M"]]);

    fireEvent.change(search, { target: { value: "camisa" } });
    expect(rows().map(rowText)).toEqual([["CAM-001", "Camisa oficial", "UN"]]);

    fireEvent.change(search, { target: { value: "nada" } });
    expect(screen.getByText("Nenhum produto encontrado para a busca.")).toBeTruthy();
    // A busca não chama a API de novo.
    expect(calls.count).toBe(1);
  });
});

// EWB-01 AC2.
describe("criação de produto", () => {
  it("envia código, nome e unidade com CSRF e mostra o produto na lista", async () => {
    mockMe(authContext({ permissions: [PERM.produtoRead, PERM.produtoCreate] }));
    const requests = mockCreate([camisa]);
    renderWithSession(<ProdutosPage />);
    await screen.findByText("Camisa oficial");
    await openCreate();
    fillCreate("BON-003", "Boné", "UN");
    submitCreate();

    expect(await screen.findByRole("link", { name: "Boné" })).toBeTruthy();
    expect(requests).toEqual([
      { body: { codigo: "BON-003", nome: "Boné", unidade_medida: "UN" }, csrf: "csrf-token-de-teste" },
    ]);
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Novo produto" })).toBeNull());
  });

  // EWB-01 AC3.
  it("409 codigo_duplicado aparece no campo de código", async () => {
    mockMe(authContext({ permissions: ALL }));
    mockCreate([camisa], () => problem(409, "codigo_duplicado"));
    renderWithSession(<ProdutosPage />);
    await screen.findByText("Camisa oficial");
    await openCreate();
    fillCreate("CAM-001", "Outra camisa", "UN");
    submitCreate();

    const message = "Já existe um produto com este código.";
    expect(await screen.findByText(message)).toBeTruthy();
    const codigo = screen.getByLabelText("Código");
    expect(isInvalid(codigo)).toBe(true);
    expect(describedByText(codigo)).toContain(message);
    expect(isInvalid(screen.getByLabelText("Nome"))).toBe(false);
    expect(screen.getByRole("dialog", { name: "Novo produto" })).toBeTruthy();
  });

  it("422 validation_failed mostra cada erro no campo correspondente", async () => {
    mockMe(authContext({ permissions: ALL }));
    mockCreate([camisa], () =>
      problem(422, "validation_failed", {
        errors: [
          { field: "nome", code: "too_large" },
          { field: "unidade_medida", code: "required" },
        ],
      }),
    );
    renderWithSession(<ProdutosPage />);
    await screen.findByText("Camisa oficial");
    await openCreate();
    fillCreate("X-1", "Nome", "UN");
    submitCreate();

    expect(await screen.findByText("Valor acima do máximo permitido.")).toBeTruthy();
    const nome = screen.getByLabelText("Nome");
    const unidade = screen.getByLabelText("Unidade de medida");
    expect(isInvalid(nome)).toBe(true);
    expect(describedByText(nome)).toContain("Valor acima do máximo permitido.");
    expect(isInvalid(unidade)).toBe(true);
    expect(describedByText(unidade)).toContain("Campo obrigatório.");
    expect(isInvalid(screen.getByLabelText("Código"))).toBe(false);
  });
});

// EWB-04 AC1.
describe("permissões", () => {
  it("somente leitura não vê criar produto", async () => {
    await renderPage(READ_ONLY);
    await screen.findByText("Camisa oficial");
    expect(screen.queryByRole("button", { name: "Novo produto" })).toBeNull();
  });

  it("sem estoque:produto:read mostra Sem acesso e não chama a API", async () => {
    const calls = await renderPage([PERM.saldoRead]);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    expect(calls.count).toBe(0);
  });
});
