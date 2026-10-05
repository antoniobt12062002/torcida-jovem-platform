import { screen } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";

import { mockProdutos } from "@/features/estoque/test-helpers";
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
