import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, ids, produto } from "@/test/msw/fixtures";

import { ProdutoDetail } from "./produto-detail";
import { mockMovimentacoes, mockProdutos, mockSaldo, PERM, READ_ONLY } from "./test-helpers";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const camisa = produto({ codigo: "CAM-001", nome: "Camisa oficial", unidade_medida: "UN" });

function renderDetail(permissions: string[], saldo = 0, id: string = ids.produto) {
  mockMe(authContext({ permissions }));
  const list = mockProdutos([camisa]);
  const saldoState = mockSaldo(saldo);
  const movs = mockMovimentacoes([]);
  renderWithSession(<ProdutoDetail id={id} />);
  return { list, saldoState, movs };
}

describe("detalhe do produto", () => {
  it("mostra código, nome e unidade lidos da lista", async () => {
    renderDetail([PERM.produtoRead]);
    expect(await screen.findByRole("heading", { name: "Camisa oficial" })).toBeTruthy();
    expect(screen.getByText("CAM-001")).toBeTruthy();
    expect(screen.getByText("UN")).toBeTruthy();
  });

  // Caso de borda: id que não está na lista.
  it("id fora da lista mostra Produto não encontrado sem consultar saldo nem histórico", async () => {
    const { saldoState, movs } = renderDetail(READ_ONLY, 5, "00000000-0000-4000-8000-000000000999");
    expect(await screen.findByText("Produto não encontrado.")).toBeTruthy();
    expect(saldoState.calls).toBe(0);
    expect(movs.calls).toBe(0);
  });
});

// EWB-01 AC4.
describe("saldo", () => {
  it("com estoque:saldo:read chama o saldo e mostra com a unidade", async () => {
    const { saldoState } = renderDetail([PERM.produtoRead, PERM.saldoRead], 12);
    const saldo = await screen.findByTestId("saldo-valor");
    expect(saldo.textContent).toBe("12 UN");
    expect(saldoState.calls).toBe(1);
    expect(screen.queryByText("Saldo negativo")).toBeNull();
  });

  it("saldo negativo é destacado", async () => {
    renderDetail([PERM.produtoRead, PERM.saldoRead], -3);
    const saldo = await screen.findByTestId("saldo-valor");
    expect(saldo.textContent).toBe("-3 UN");
    expect(screen.getByText("Saldo negativo")).toBeTruthy();
    expect(saldo.getAttribute("data-negative")).toBe("true");
  });

  it("saldo zero não é destacado", async () => {
    renderDetail([PERM.produtoRead, PERM.saldoRead], 0);
    const saldo = await screen.findByTestId("saldo-valor");
    expect(saldo.textContent).toBe("0 UN");
    expect(screen.queryByText("Saldo negativo")).toBeNull();
    expect(saldo.getAttribute("data-negative")).toBe("false");
  });
});

// EWB-04 AC2.
describe("seções por permissão", () => {
  it("sem estoque:saldo:read a seção de saldo não aparece e a API de saldo não é chamada", async () => {
    const { saldoState } = renderDetail([PERM.produtoRead, PERM.movRead]);
    await screen.findByRole("heading", { name: "Camisa oficial" });
    expect(screen.queryByRole("heading", { name: "Saldo" })).toBeNull();
    expect(screen.queryByTestId("saldo-valor")).toBeNull();
    expect(saldoState.calls).toBe(0);
  });

  it("sem estoque:produto:read mostra Sem acesso e não consulta nada", async () => {
    const { list, saldoState, movs } = renderDetail([PERM.saldoRead, PERM.movRead]);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    expect(list.count).toBe(0);
    expect(saldoState.calls).toBe(0);
    expect(movs.calls).toBe(0);
  });
});
