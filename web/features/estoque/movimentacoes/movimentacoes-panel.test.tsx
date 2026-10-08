import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, ids, movimentacao, problem, produto } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";
import { http } from "msw";

import { ProdutoDetail } from "../produto-detail";
import { ALL, allExcept, mockEstoque, mockProdutos, PERM, READ_ONLY, type } from "../test-helpers";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const camisa = produto({ unidade_medida: "UN" });

const M = {
  entrada: "00000000-0000-4000-8000-000000000511",
  saida: "00000000-0000-4000-8000-000000000512",
  ajuste: "00000000-0000-4000-8000-000000000513",
  devolucao: "00000000-0000-4000-8000-000000000514",
};

const historico = [
  movimentacao({ id: M.entrada, tipo: "ENTRADA", quantidade: 10, responsavel_id: ids.user }),
  movimentacao({ id: M.saida, tipo: "SAIDA", quantidade: -4, responsavel_id: ids.otherUser }),
  movimentacao({
    id: M.ajuste,
    tipo: "AJUSTE",
    quantidade: -1,
    origem: "AJUSTE_MANUAL",
    motivo: "Peça danificada",
  }),
  movimentacao({ id: M.devolucao, tipo: "DEVOLUCAO", quantidade: 4, movimentacao_de_id: M.saida }),
];

function renderDetail(permissions: string[], initial = historico) {
  mockMe(authContext({ permissions }));
  mockProdutos([camisa]);
  const backend = mockEstoque(initial);
  renderWithSession(<ProdutoDetail id={camisa.id} />);
  return backend;
}

/** Linhas do histórico; `hidden` lê também com um diálogo modal aberto por cima. */
function historyRows(hidden = false) {
  const table = screen.getByRole("table", { name: "Movimentações", hidden });
  return within(table).getAllByRole("row", { hidden }).slice(1);
}

function cells(row: HTMLElement) {
  return within(row)
    .getAllByRole("cell")
    .map((c) => c.textContent);
}

async function saldoText(expected: string) {
  await waitFor(() => expect(screen.getByTestId("saldo-valor").textContent).toBe(expected));
}

async function openDialog(trigger: HTMLElement, title: string) {
  fireEvent.click(trigger);
  return screen.findByRole("dialog", { name: title });
}

async function submit(dialog: HTMLElement) {
  const button = await within(dialog).findByRole("button", { name: "Registrar" });
  fireEvent.click(button);
}

// EWB-02 AC1.
describe("histórico", () => {
  it("chama o histórico e mostra todas as colunas, na ordem da API", async () => {
    const backend = renderDetail(READ_ONLY);
    await screen.findByRole("table", { name: "Movimentações" });
    const headers = within(screen.getByRole("table", { name: "Movimentações" }))
      .getAllByRole("columnheader")
      .map((h) => h.textContent);
    expect(headers).toEqual([
      "Tipo",
      "Quantidade",
      "Origem",
      "Motivo",
      "Devolução de",
      "Responsável",
      "Data",
    ]);
    expect(historyRows().map(cells)).toEqual([
      ["Entrada", "10", "Inventário", "—", "—", "você", "04/10/2026 09:00"],
      ["Saída", "-4", "Inventário", "—", "—", "00000000", "04/10/2026 09:00"],
      ["Ajuste", "-1", "Ajuste manual", "Peça danificada", "—", "você", "04/10/2026 09:00"],
      ["Devolução", "4", "Inventário", "—", M.saida.slice(0, 8), "você", "04/10/2026 09:00"],
    ]);
    expect(backend.movCalls).toBe(1);
  });

  // EWB-04 AC2.
  it("sem estoque:movimentacao:read a seção não aparece e o histórico não é chamado", async () => {
    const backend = renderDetail([PERM.produtoRead, PERM.saldoRead]);
    await screen.findByTestId("saldo-valor");
    expect(screen.queryByRole("heading", { name: "Movimentações" })).toBeNull();
    expect(backend.movCalls).toBe(0);
  });
});

// Caso de borda e EWB-02 AC3.
describe("Devolver", () => {
  it("aparece só nas movimentações ENTRADA e SAIDA", async () => {
    renderDetail(ALL);
    await screen.findByRole("table", { name: "Movimentações" });
    const withDevolver = historyRows().map(
      (row) => within(row).queryByRole("button", { name: "Devolver" }) !== null,
    );
    expect(withDevolver).toEqual([true, true, false, false]);
  });
});

// EWB-02 AC2, AC4, AC6 (e o teste independente da spec).
describe("registro de movimentações", () => {
  it("entrada de 10, saída de 4, devolução da saída; corpo sempre com INVENTARIO", async () => {
    const backend = renderDetail(ALL, []);
    await saldoText("0 UN");

    // Entrada: a origem é fixa, sem opção de escolha.
    let dialog = await openDialog(screen.getByRole("button", { name: "Registrar entrada" }), "Registrar entrada");
    expect(within(dialog).getByText("Inventário")).toBeTruthy();
    expect(within(dialog).queryByRole("combobox")).toBeNull();
    expect(within(dialog).queryByRole("radio")).toBeNull();
    expect(within(dialog).queryByRole("listbox")).toBeNull();
    for (const reservada of ["Venda", "Compra", "Evento", "Ajuste manual", "VENDA", "COMPRA", "EVENTO", "AJUSTE_MANUAL"]) {
      expect(within(dialog).queryByText(reservada)).toBeNull();
    }
    type("Quantidade", "10");
    await submit(dialog);
    await saldoText("10 UN");
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Registrar entrada" })).toBeNull());
    expect(historyRows().map((r) => cells(r).slice(0, 3))).toEqual([["Entrada", "10", "Inventário"]]);

    // Saída.
    dialog = await openDialog(screen.getByRole("button", { name: "Registrar saída" }), "Registrar saída");
    expect(within(dialog).getByText("Inventário")).toBeTruthy();
    type("Quantidade", "4");
    await submit(dialog);
    await saldoText("6 UN");
    await waitFor(() => expect(historyRows()).toHaveLength(2));

    // Devolução da saída, a partir do histórico.
    const saidaRow = historyRows()[1];
    dialog = await openDialog(within(saidaRow).getByRole("button", { name: "Devolver" }), "Registrar devolução");
    expect(within(dialog).getByText("Inventário")).toBeTruthy();
    type("Quantidade", "4");
    await submit(dialog);
    await saldoText("10 UN");
    await waitFor(() => expect(historyRows()).toHaveLength(3));

    const saidaId = backend.items[1].id;
    expect(backend.movRequests).toEqual([
      {
        body: { tipo: "ENTRADA", produto_id: camisa.id, quantidade: 10, origem: "INVENTARIO" },
        csrf: "csrf-token-de-teste",
      },
      {
        body: { tipo: "SAIDA", produto_id: camisa.id, quantidade: 4, origem: "INVENTARIO" },
        csrf: "csrf-token-de-teste",
      },
      {
        body: {
          tipo: "DEVOLUCAO",
          produto_id: camisa.id,
          quantidade: 4,
          origem: "INVENTARIO",
          movimentacao_de_id: saidaId,
        },
        csrf: "csrf-token-de-teste",
      },
    ]);
    // Histórico e saldo rebuscados depois de cada registro (EWB-02 AC6).
    expect(backend.movCalls).toBe(4);
    expect(backend.saldoCalls).toBe(4);
  });

  // EWB-02 AC5.
  it("saída maior que o saldo mostra saldo insuficiente e não altera nada", async () => {
    const backend = renderDetail(ALL, [movimentacao({ tipo: "ENTRADA", quantidade: 2 })]);
    await saldoText("2 UN");
    const dialog = await openDialog(screen.getByRole("button", { name: "Registrar saída" }), "Registrar saída");
    type("Quantidade", "5");
    await submit(dialog);
    expect(await within(dialog).findByText("Saldo insuficiente para esta saída.")).toBeTruthy();
    expect(screen.getByRole("dialog", { name: "Registrar saída" })).toBeTruthy();
    expect(backend.movRequests).toHaveLength(1);
    expect(screen.getByTestId("saldo-valor").textContent).toBe("2 UN");
    expect(historyRows(true)).toHaveLength(1);
  });

  it.each([
    [404, "produto_nao_encontrado", "Produto não encontrado."],
    [422, "quantidade_invalida", "Quantidade inválida."],
    [422, "devolucao_invalida", "Só é possível devolver uma entrada ou saída deste produto."],
  ])("%i %s mostra a mensagem", async (status, code, message) => {
    renderDetail(ALL);
    await screen.findByRole("table", { name: "Movimentações" });
    server.use(http.post(apiUrl("/api/v1/estoque/movimentacoes"), () => problem(status, code)));
    const dialog = await openDialog(within(historyRows()[0]).getByRole("button", { name: "Devolver" }), "Registrar devolução");
    type("Quantidade", "1");
    await submit(dialog);
    expect(await within(dialog).findByText(message)).toBeTruthy();
  });

  it("quantidade zero, negativa ou vazia não chama a API", async () => {
    const backend = renderDetail(ALL, []);
    await saldoText("0 UN");
    const dialog = await openDialog(screen.getByRole("button", { name: "Registrar entrada" }), "Registrar entrada");
    await submit(dialog);
    expect(await within(dialog).findByText("Informe a quantidade.")).toBeTruthy();
    type("Quantidade", "0");
    await submit(dialog);
    type("Quantidade", "-2");
    await submit(dialog);
    await waitFor(() => expect(within(dialog).getByText("Informe um número maior que zero.")).toBeTruthy());
    expect(backend.movRequests).toEqual([]);
  });
});

// EWB-04 AC1.
describe("somente leitura", () => {
  it("vê histórico e saldo, sem entrada, saída nem devolução", async () => {
    renderDetail(READ_ONLY);
    await screen.findByRole("table", { name: "Movimentações" });
    await screen.findByTestId("saldo-valor");
    expect(screen.queryByRole("button", { name: "Registrar entrada" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Registrar saída" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Devolver" })).toBeNull();
  });
});

// EWB-02 AC1 (lacuna G1): a tela mantém a ordem devolvida pela API. As datas
// são distintas e não estão em ordem crescente nem decrescente, e os ids,
// quantidades e tipos também não, então qualquer reordenação aparece.
describe("ordem do histórico", () => {
  const O = {
    a: "00000000-0000-4000-8000-000000000523",
    b: "00000000-0000-4000-8000-000000000521",
    c: "00000000-0000-4000-8000-000000000524",
    d: "00000000-0000-4000-8000-000000000522",
  };
  const foraDeOrdem = [
    movimentacao({ id: O.a, tipo: "SAIDA", quantidade: -2, criado_em: "2026-10-03T12:00:00Z" }),
    movimentacao({ id: O.b, tipo: "ENTRADA", quantidade: 30, criado_em: "2026-10-05T13:30:00Z" }),
    movimentacao({
      id: O.c,
      tipo: "AJUSTE",
      quantidade: 5,
      origem: "AJUSTE_MANUAL",
      motivo: "Recontagem",
      criado_em: "2026-10-01T11:15:00Z",
    }),
    movimentacao({ id: O.d, tipo: "ENTRADA", quantidade: 8, criado_em: "2026-10-04T14:45:00Z" }),
  ];

  it("a fixture não está em ordem cronológica, em nenhum sentido", () => {
    const datas = foraDeOrdem.map((m) => m.criado_em);
    expect(new Set(datas).size).toBe(datas.length);
    expect(datas).not.toEqual([...datas].sort());
    expect(datas).not.toEqual([...datas].sort().reverse());
  });

  it("mostra as linhas exatamente na ordem da resposta, sem reordenar por data", async () => {
    const backend = renderDetail(READ_ONLY, foraDeOrdem);
    await screen.findByRole("table", { name: "Movimentações" });
    expect(historyRows().map(cells)).toEqual([
      ["Saída", "-2", "Inventário", "—", "—", "você", "03/10/2026 09:00"],
      ["Entrada", "30", "Inventário", "—", "—", "você", "05/10/2026 10:30"],
      ["Ajuste", "5", "Ajuste manual", "Recontagem", "—", "você", "01/10/2026 08:15"],
      ["Entrada", "8", "Inventário", "—", "—", "você", "04/10/2026 11:45"],
    ]);
    expect(backend.movCalls).toBe(1);
  });

  it("a ordem da resposta também vale quando a API devolve o inverso", async () => {
    renderDetail(READ_ONLY, [...foraDeOrdem].reverse());
    await screen.findByRole("table", { name: "Movimentações" });
    expect(historyRows().map((row) => cells(row)[6])).toEqual([
      "04/10/2026 11:45",
      "01/10/2026 08:15",
      "05/10/2026 10:30",
      "03/10/2026 09:00",
    ]);
  });
});

// EWB-02 AC2-3, EWB-03 AC1, FND-04 AC3 (lacuna G5): entrada, saída e Devolver
// dependem só de estoque:movimentacao:create; o ajuste, só de
// estoque:movimentacao:adjust.
describe("isolamento das permissões de escrita no detalhe", () => {
  const HEADERS = ["Tipo", "Quantidade", "Origem", "Motivo", "Devolução de", "Responsável", "Data"];

  async function ready() {
    await screen.findByRole("table", { name: "Movimentações" });
    await screen.findByTestId("saldo-valor");
  }

  function headers() {
    return within(screen.getByRole("table", { name: "Movimentações" }))
      .getAllByRole("columnheader")
      .map((h) => h.textContent);
  }

  function devolverPorLinha() {
    return historyRows().map((row) => within(row).queryByRole("button", { name: "Devolver" }) !== null);
  }

  it("leituras + só estoque:movimentacao:create: entrada, saída e Devolver, sem Ajustar", async () => {
    renderDetail([...READ_ONLY, PERM.movCreate]);
    await ready();
    expect(screen.getByRole("button", { name: "Registrar entrada" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Registrar saída" })).toBeTruthy();
    expect(devolverPorLinha()).toEqual([true, true, false, false]);
    expect(headers()).toEqual([...HEADERS, "Ações"]);
    expect(screen.queryByRole("button", { name: "Ajustar estoque" })).toBeNull();
  });

  it("leituras + só estoque:movimentacao:adjust: Ajustar, sem entrada, saída nem Devolver", async () => {
    renderDetail([...READ_ONLY, PERM.movAdjust]);
    await ready();
    expect(screen.getByRole("button", { name: "Ajustar estoque" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Registrar entrada" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Registrar saída" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Devolver" })).toBeNull();
    expect(devolverPorLinha()).toEqual([false, false, false, false]);
    // Sem a permissão, nem a coluna de ações existe.
    expect(headers()).toEqual(HEADERS);
  });

  it("todas menos estoque:movimentacao:create: sem entrada, saída nem Devolver; Ajustar continua", async () => {
    renderDetail(allExcept(PERM.movCreate));
    await ready();
    expect(screen.queryByRole("button", { name: "Registrar entrada" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Registrar saída" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Devolver" })).toBeNull();
    expect(headers()).toEqual(HEADERS);
    expect(screen.getByRole("button", { name: "Ajustar estoque" })).toBeTruthy();
  });

  it("todas menos estoque:movimentacao:adjust: sem Ajustar; entrada, saída e Devolver continuam", async () => {
    renderDetail(allExcept(PERM.movAdjust));
    await ready();
    expect(screen.queryByRole("button", { name: "Ajustar estoque" })).toBeNull();
    expect(screen.getByRole("button", { name: "Registrar entrada" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Registrar saída" })).toBeTruthy();
    expect(devolverPorLinha()).toEqual([true, true, false, false]);
  });

  it("estoque:produto:create não libera nenhuma ação do detalhe", async () => {
    renderDetail([...READ_ONLY, PERM.produtoCreate]);
    await ready();
    expect(screen.queryByRole("button", { name: "Registrar entrada" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Registrar saída" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Devolver" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Ajustar estoque" })).toBeNull();
    expect(headers()).toEqual(HEADERS);
  });

  // Entrada e saída ficam fora das seções de leitura: só dependem da rota e
  // de estoque:movimentacao:create.
  it("só a rota + estoque:movimentacao:create: entrada e saída, sem consultar saldo nem histórico", async () => {
    const backend = renderDetail([PERM.produtoRead, PERM.movCreate]);
    expect(await screen.findByRole("button", { name: "Registrar entrada" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Registrar saída" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Ajustar estoque" })).toBeNull();
    expect(backend.saldoCalls).toBe(0);
    expect(backend.movCalls).toBe(0);
  });
});
