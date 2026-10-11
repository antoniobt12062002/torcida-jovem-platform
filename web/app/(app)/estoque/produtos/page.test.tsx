import { screen } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";

import { allExcept, mockMovimentacoes, mockProdutos, mockSaldo, PERM } from "@/features/estoque/test-helpers";
import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, produto } from "@/test/msw/fixtures";

import EstoqueProdutosPage from "./page";
import EstoqueProdutoPage from "./[id]/page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

it("/estoque/produtos monta a lista de produtos", async () => {
  mockMe(authContext({ permissions: ["estoque:produto:read"] }));
  mockProdutos([produto()]);
  renderWithSession(<EstoqueProdutosPage />);
  expect(await screen.findByRole("heading", { name: "Produtos" })).toBeTruthy();
  expect(await screen.findByText("Camisa oficial")).toBeTruthy();
});

it("/estoque/produtos/[id] monta o detalhe do produto do id", async () => {
  const p = produto();
  mockMe(authContext({ permissions: ["estoque:produto:read"] }));
  mockProdutos([p]);
  renderWithSession(await EstoqueProdutoPage({ params: Promise.resolve({ id: p.id }), searchParams: Promise.resolve({}) }));
  expect(await screen.findByRole("heading", { name: "Camisa oficial" })).toBeTruthy();
});

// FND-04 AC4: as duas rotas dependem de estoque:produto:read, e de nenhuma
// outra permissão do estoque.
function mockArea() {
  const p = produto();
  return { p, list: mockProdutos([p]), saldo: mockSaldo(3), movs: mockMovimentacoes([]) };
}

async function settle() {
  await new Promise((resolve) => setTimeout(resolve, 20));
}

it("/estoque/produtos sem estoque:produto:read, com todas as outras do estoque: Sem acesso, sem requisição", async () => {
  mockMe(authContext({ permissions: allExcept(PERM.produtoRead) }));
  const area = mockArea();
  renderWithSession(<EstoqueProdutosPage />);
  expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
  expect(screen.queryByRole("heading", { name: "Produtos" })).toBeNull();
  await settle();
  expect(area.list.count).toBe(0);
});

it("/estoque/produtos só com estoque:produto:read chama a lista uma vez", async () => {
  mockMe(authContext({ permissions: [PERM.produtoRead] }));
  const area = mockArea();
  renderWithSession(<EstoqueProdutosPage />);
  expect(await screen.findByText("Camisa oficial")).toBeTruthy();
  expect(screen.queryByRole("heading", { name: "Sem acesso" })).toBeNull();
  expect(area.list.count).toBe(1);
});

it("/estoque/produtos/[id] sem estoque:produto:read, com todas as outras do estoque: Sem acesso, sem requisição", async () => {
  mockMe(authContext({ permissions: allExcept(PERM.produtoRead) }));
  const area = mockArea();
  renderWithSession(
    await EstoqueProdutoPage({ params: Promise.resolve({ id: area.p.id }), searchParams: Promise.resolve({}) }),
  );
  expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
  expect(screen.queryByRole("heading", { name: "Camisa oficial" })).toBeNull();
  await settle();
  expect(area.list.count).toBe(0);
  expect(area.saldo.calls).toBe(0);
  expect(area.movs.calls).toBe(0);
});

it("/estoque/produtos/[id] só com estoque:produto:read abre o detalhe, sem saldo nem histórico", async () => {
  mockMe(authContext({ permissions: [PERM.produtoRead] }));
  const area = mockArea();
  renderWithSession(
    await EstoqueProdutoPage({ params: Promise.resolve({ id: area.p.id }), searchParams: Promise.resolve({}) }),
  );
  expect(await screen.findByRole("heading", { name: "Camisa oficial" })).toBeTruthy();
  expect(screen.queryByRole("heading", { name: "Sem acesso" })).toBeNull();
  expect(area.list.count).toBe(1);
  expect(area.saldo.calls).toBe(0);
  expect(area.movs.calls).toBe(0);
});
