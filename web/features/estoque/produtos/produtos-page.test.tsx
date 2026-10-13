import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, type Produto, problem, produto } from "@/test/msw/fixtures";
import { mockViewport } from "@/test/media";
import { server } from "@/test/msw/server";

import {
  ALL,
  allExcept,
  describedByText,
  isInvalid,
  mockProdutos,
  PERM,
  READ_ONLY,
  type Recorded,
  type,
} from "../test-helpers";
import { PRODUTOS_PAGE_SIZE, ProdutosPage } from "./produtos-page";

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

// EWB-01 AC1-2, FND-04 AC3-4: a rota depende só de estoque:produto:read e
// "Novo produto", só de estoque:produto:create.
describe("isolamento das permissões da lista", () => {
  it("só estoque:produto:read lista os produtos, sem Novo produto", async () => {
    const calls = await renderPage([PERM.produtoRead]);
    await screen.findByText("Camisa oficial");
    expect(screen.queryByRole("heading", { name: "Sem acesso" })).toBeNull();
    expect(calls.count).toBe(1);
    expect(screen.queryByRole("button", { name: "Novo produto" })).toBeNull();
  });

  it("todas as permissões do estoque menos estoque:produto:read: Sem acesso e a API não é chamada", async () => {
    const calls = await renderPage(allExcept(PERM.produtoRead));
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    expect(screen.queryByRole("heading", { name: "Produtos" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Novo produto" })).toBeNull();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(calls.count).toBe(0);
  });

  it("leitura + só estoque:produto:create mostra Novo produto", async () => {
    await renderPage([PERM.produtoRead, PERM.produtoCreate]);
    await screen.findByText("Camisa oficial");
    expect(screen.getByRole("button", { name: "Novo produto" })).toBeTruthy();
  });

  it("todas as permissões menos estoque:produto:create: sem Novo produto", async () => {
    await renderPage(allExcept(PERM.produtoCreate));
    await screen.findByText("Camisa oficial");
    expect(screen.queryByRole("button", { name: "Novo produto" })).toBeNull();
  });

  it.each([
    ["estoque:movimentacao:create", PERM.movCreate],
    ["estoque:movimentacao:adjust", PERM.movAdjust],
  ])("leituras + só %s: sem Novo produto", async (_name, permission) => {
    await renderPage([...READ_ONLY, permission]);
    await screen.findByText("Camisa oficial");
    expect(screen.queryByRole("button", { name: "Novo produto" })).toBeNull();
  });
});

// DSP-05 AC2 a AC4, DSM-04 AC2: o termo de busca fica na URL.
describe("busca na URL", () => {
  it("abre com a busca da URL aplicada e no campo", async () => {
    setLocation("/estoque/produtos?busca=fai");
    await renderPage([PERM.produtoRead]);
    await screen.findByText("Faixa da torcida");
    expect((screen.getByLabelText("Buscar por código ou nome") as HTMLInputElement).value).toBe("fai");
    expect(rows().map(rowText)).toEqual([["FAI-002", "Faixa da torcida", "M"]]);
    expect(screen.getByText("1 produto")).toBeTruthy();
  });

  it("digitar escreve na URL uma vez, depois da espera, sem criar entrada de histórico", async () => {
    setLocation("/estoque/produtos");
    await renderPage([PERM.produtoRead]);
    await screen.findByText("Camisa oficial");
    const push = vi.spyOn(window.history, "pushState");
    const replace = vi.spyOn(window.history, "replaceState");
    const search = screen.getByLabelText("Buscar por código ou nome");

    for (const value of ["c", "ca", "cam"]) fireEvent.change(search, { target: { value } });
    // A lista filtra na hora; a URL só muda depois da espera.
    expect(rows().map(rowText)).toEqual([["CAM-001", "Camisa oficial", "UN"]]);
    expect(window.location.search).toBe("");

    await waitFor(() => expect(window.location.search).toBe("?busca=cam"));
    expect(replace).toHaveBeenCalledTimes(1);
    expect(push).not.toHaveBeenCalled();
    push.mockRestore();
    replace.mockRestore();
  });

  it("busca sem resultado oferece limpar, que volta a lista inteira e tira o parâmetro", async () => {
    setLocation("/estoque/produtos?busca=nada");
    await renderPage([PERM.produtoRead]);
    expect(await screen.findByRole("heading", { name: "Nenhum produto encontrado para a busca." })).toBeTruthy();
    expect(screen.getByText("0 produtos")).toBeTruthy();
    fireEvent.click(screen.getAllByRole("button", { name: "Limpar filtros" })[0]);
    await waitFor(() => expect(rows()).toHaveLength(2));
    await waitFor(() => expect(window.location.search).toBe(""));
  });
});

// DSP-05 AC5 a AC7, DSM-04 AC1.
describe("listagem responsiva", () => {
  it("abaixo de 768 px mostra lista empilhada com nome, código e unidade", async () => {
    mockViewport({ width: 390, coarse: true });
    await renderPage([PERM.produtoRead]);
    const list = await screen.findByRole("list", { name: "Produtos" });
    expect(screen.queryByRole("table")).toBeNull();
    const items = within(list).getAllByRole("listitem");
    expect(items).toHaveLength(2);
    expect(within(items[1]).getByRole("link", { name: "Faixa da torcida" }).getAttribute("href")).toBe(
      `/estoque/produtos/${faixa.id}`,
    );
    expect(items[1].textContent).toContain("FAI-002");
    expect(items[1].textContent).toContain("Unidade de medida: M");
  });

  // Smoke da rodada 6, D2: o nome é o único alvo do item; com ponteiro de
  // toque ele tem a altura do token `control-sm` (44 px em `pointer: coarse`).
  it("o link do nome tem alvo de toque, centrado na altura", async () => {
    mockViewport({ width: 390, coarse: true });
    await renderPage([PERM.produtoRead]);
    const list = await screen.findByRole("list", { name: "Produtos" });
    const link = within(list).getByRole("link", { name: "Faixa da torcida" });
    for (const name of ["min-h-control-sm", "items-center", "content-center", "touch-manipulation"]) {
      expect(link.className.split(" ")).toContain(name);
    }
  });

  it("a partir de 768 px mostra a tabela", async () => {
    mockViewport({ width: 1024 });
    await renderPage([PERM.produtoRead]);
    expect(await screen.findByRole("table", { name: "Produtos" })).toBeTruthy();
  });

  it(`mostra ${PRODUTOS_PAGE_SIZE} produtos por vez e "Mostrar mais" só quando há mais`, async () => {
    const muitos = Array.from({ length: PRODUTOS_PAGE_SIZE + 3 }, (_, i) =>
      produto({
        id: `00000000-0000-4000-8000-${String(100000 + i).padStart(12, "0")}`,
        codigo: `P-${String(i).padStart(3, "0")}`,
        nome: `Produto ${i}`,
      }),
    );
    await renderPage([PERM.produtoRead], muitos);
    await screen.findByText("Produto 0");
    expect(rows()).toHaveLength(PRODUTOS_PAGE_SIZE);
    expect(screen.getByText(`Mostrando ${PRODUTOS_PAGE_SIZE} de ${PRODUTOS_PAGE_SIZE + 3} produtos`)).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Mostrar mais" }));
    expect(rows()).toHaveLength(PRODUTOS_PAGE_SIZE + 3);
    expect(screen.queryByRole("button", { name: "Mostrar mais" })).toBeNull();
  });

  it(`com exatamente ${PRODUTOS_PAGE_SIZE} produtos não há "Mostrar mais"`, async () => {
    const exatos = Array.from({ length: PRODUTOS_PAGE_SIZE }, (_, i) =>
      produto({ id: `00000000-0000-4000-8000-${String(200000 + i).padStart(12, "0")}`, codigo: `Q-${i}`, nome: `Item ${i}` }),
    );
    await renderPage([PERM.produtoRead], exatos);
    await screen.findByText("Item 0");
    expect(rows()).toHaveLength(PRODUTOS_PAGE_SIZE);
    expect(screen.queryByRole("button", { name: "Mostrar mais" })).toBeNull();
  });
});

// DSP-02, DSM-04 AC1: cabeçalho e estado vazio com a ação de criar.
describe("cabeçalho e estado vazio", () => {
  it("um único h1 Produtos, no cabeçalho de página", async () => {
    await renderPage([PERM.produtoRead]);
    await screen.findByText("Camisa oficial");
    expect(screen.getAllByRole("heading", { level: 1 }).map((h) => h.textContent)).toEqual(["Produtos"]);
  });

  it("sem produtos, quem cria vê Novo produto só no estado vazio, e ele abre a criação", async () => {
    await renderPage([PERM.produtoRead, PERM.produtoCreate], []);
    expect(await screen.findByRole("heading", { name: "Nenhum produto cadastrado." })).toBeTruthy();
    expect(screen.getAllByRole("button", { name: "Novo produto" })).toHaveLength(1);
    expect(screen.queryByLabelText("Buscar por código ou nome")).toBeNull();
    await openCreate();
  });

  it("sem produtos e sem estoque:produto:create, o estado vazio não tem ação", async () => {
    await renderPage(READ_ONLY, []);
    expect(await screen.findByRole("heading", { name: "Nenhum produto cadastrado." })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Novo produto" })).toBeNull();
  });
});
