import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, ids, produto } from "@/test/msw/fixtures";

import { ProdutoDetail } from "./produto-detail";
import { allExcept, mockMovimentacoes, mockProdutos, mockSaldo, PERM, READ_ONLY, WRITES } from "./test-helpers";

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

// FND-04 AC3-4, EWB-01 AC4, EWB-02 AC1, EWB-04 AC2: cada seção e a rota são
// decididas pela sua própria permissão de leitura, e por nenhuma outra.
describe("isolamento das permissões de leitura", () => {
  it("só estoque:produto:read abre o detalhe, sem seção de saldo nem de histórico", async () => {
    const { list, saldoState, movs } = renderDetail([PERM.produtoRead]);
    expect(await screen.findByRole("heading", { name: "Camisa oficial" })).toBeTruthy();
    expect(screen.queryByRole("heading", { name: "Sem acesso" })).toBeNull();
    expect(list.count).toBe(1);
    expect(screen.queryByRole("heading", { name: "Saldo" })).toBeNull();
    expect(screen.queryByRole("heading", { name: "Movimentações" })).toBeNull();
    expect(saldoState.calls).toBe(0);
    expect(movs.calls).toBe(0);
  });

  it("todas as permissões do estoque menos estoque:produto:read: Sem acesso e nenhuma requisição", async () => {
    const { list, saldoState, movs } = renderDetail(allExcept(PERM.produtoRead), 5);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    expect(screen.queryByRole("heading", { name: "Camisa oficial" })).toBeNull();
    expect(screen.queryByRole("heading", { name: "Saldo" })).toBeNull();
    expect(screen.queryByRole("heading", { name: "Movimentações" })).toBeNull();
    // Dá tempo a uma consulta indevida de acontecer antes de contar.
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(list.count).toBe(0);
    expect(saldoState.calls).toBe(0);
    expect(movs.calls).toBe(0);
  });

  it("estoque:saldo:read sozinha (com a rota) mostra o saldo e não chama o histórico", async () => {
    const { saldoState, movs } = renderDetail([PERM.produtoRead, PERM.saldoRead], 7);
    expect((await screen.findByTestId("saldo-valor")).textContent).toBe("7 UN");
    expect(screen.getByRole("heading", { name: "Saldo" })).toBeTruthy();
    expect(saldoState.calls).toBe(1);
    expect(screen.queryByRole("heading", { name: "Movimentações" })).toBeNull();
    expect(movs.calls).toBe(0);
  });

  it("todas as permissões menos estoque:saldo:read: sem seção de saldo e sem chamar o saldo", async () => {
    const { saldoState, movs } = renderDetail(allExcept(PERM.saldoRead), 7);
    expect(await screen.findByRole("heading", { name: "Movimentações" })).toBeTruthy();
    await screen.findByText("Nenhuma movimentação registrada.");
    expect(movs.calls).toBe(1);
    expect(screen.queryByRole("heading", { name: "Saldo" })).toBeNull();
    expect(screen.queryByTestId("saldo-valor")).toBeNull();
    expect(saldoState.calls).toBe(0);
  });

  it("estoque:movimentacao:read sozinha (com a rota) mostra o histórico e não chama o saldo", async () => {
    const { saldoState, movs } = renderDetail([PERM.produtoRead, PERM.movRead], 7);
    expect(await screen.findByRole("heading", { name: "Movimentações" })).toBeTruthy();
    await screen.findByText("Nenhuma movimentação registrada.");
    expect(movs.calls).toBe(1);
    expect(screen.queryByRole("heading", { name: "Saldo" })).toBeNull();
    expect(saldoState.calls).toBe(0);
  });

  it("todas as permissões menos estoque:movimentacao:read: sem histórico e sem chamá-lo", async () => {
    const { saldoState, movs } = renderDetail(allExcept(PERM.movRead), 7);
    expect((await screen.findByTestId("saldo-valor")).textContent).toBe("7 UN");
    expect(saldoState.calls).toBe(1);
    expect(screen.queryByRole("heading", { name: "Movimentações" })).toBeNull();
    expect(screen.queryByRole("table")).toBeNull();
    expect(movs.calls).toBe(0);
  });

  // As escritas não abrem nenhuma leitura.
  it("só as permissões de escrita, com a rota: nenhuma seção de leitura", async () => {
    const { saldoState, movs } = renderDetail([PERM.produtoRead, ...WRITES], 7);
    await screen.findByRole("heading", { name: "Camisa oficial" });
    expect(screen.queryByRole("heading", { name: "Saldo" })).toBeNull();
    expect(screen.queryByRole("heading", { name: "Movimentações" })).toBeNull();
    expect(saldoState.calls).toBe(0);
    expect(movs.calls).toBe(0);
  });
});
