import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { queryKeys } from "@/lib/api/query-keys";
import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, type Lancamento, lancamento, problem } from "@/test/msw/fixtures";
import { createTestQueryClient } from "@/test/render";

import { LancamentoDetail } from "./lancamento-detail";
import {
  change,
  CONTA_DESPESA,
  CONTA_RECEITA,
  detailValue,
  dialog,
  formDialog,
  mockComprovantes,
  mockContas,
  mockLancamentos,
  mockWrite,
  PERMISSOES_LEITURA,
  PERMISSOES_TESOURARIA,
  leituraMais,
  tesourariaSem,
} from "./test-helpers";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const ID = "00000000-0000-4000-8000-000000000231";

const RECEITA_CRIADA = lancamento({ id: ID, tipo: "RECEITA", conta_id: CONTA_RECEITA.id, status: "CRIADA" });
const DESPESA_CRIADA = lancamento({ id: ID, tipo: "DESPESA", conta_id: CONTA_DESPESA.id, status: "CRIADA" });
const RECEITA_RECEBIDA = lancamento({ ...RECEITA_CRIADA, status: "RECEBIDA" });

/**
 * Renderiza o detalhe com a lista servida por `state.items` e um saldo já em
 * cache, para observar a invalidação.
 */
async function renderDetail(initial: Lancamento[], permissions = PERMISSOES_TESOURARIA) {
  const state = { items: initial };
  mockMe(authContext({ permissions }));
  mockContas();
  mockComprovantes();
  const list = mockLancamentos(() => state.items);
  const queryClient = createTestQueryClient();
  queryClient.setQueryData(queryKeys.financeiro.saldo, { saldo_cents: 0 });
  renderWithSession(<LancamentoDetail id={ID} />, { queryClient });
  await screen.findByText("Valor bruto", { selector: "dt" });
  return { state, list, queryClient };
}

function saldoInvalidated(queryClient: ReturnType<typeof createTestQueryClient>) {
  return queryClient.getQueryState(queryKeys.financeiro.saldo)?.isInvalidated;
}

function badge() {
  return screen.getByRole("heading", { level: 1 }).parentElement?.querySelector('[data-slot="badge"]')?.textContent;
}

async function confirmIn(buttonName: string) {
  const alert = await screen.findByRole("alertdialog");
  fireEvent.click(within(alert).getByRole("button", { name: buttonName }));
}

// FWB-04 AC1.
describe("receber", () => {
  it("confirma, chama /receive com CSRF, mostra Recebida e invalida o saldo", async () => {
    const { state, list, queryClient } = await renderDetail([RECEITA_CRIADA]);
    const requests = mockWrite("post", `/lancamentos/${ID}/receive`, () => {
      state.items = [{ ...RECEITA_CRIADA, status: "RECEBIDA" }];
      return new HttpResponse(null, { status: 204 });
    });
    fireEvent.click(screen.getByRole("button", { name: "Receber" }));
    expect(requests).toEqual([]);
    await confirmIn("Confirmar recebimento");
    await waitFor(() => expect(requests).toEqual([{ body: null, csrf: "csrf-token-de-teste" }]));
    await waitFor(() => expect(badge()).toBe("Recebida"));
    expect(list.count).toBe(2);
    expect(saldoInvalidated(queryClient)).toBe(true);
  });

  it("lancamento_nao_pode_ser_recebido mostra a mensagem e atualiza a lista", async () => {
    const { state, list } = await renderDetail([RECEITA_CRIADA]);
    mockWrite("post", `/lancamentos/${ID}/receive`, () => {
      state.items = [{ ...RECEITA_CRIADA, status: "CANCELADA" }];
      return problem(409, "lancamento_nao_pode_ser_recebido");
    });
    fireEvent.click(screen.getByRole("button", { name: "Receber" }));
    await confirmIn("Confirmar recebimento");
    expect(await screen.findByText("Este lançamento não está mais em aberto.")).toBeTruthy();
    await waitFor(() => expect(list.count).toBe(2));
    await waitFor(() => expect(badge()).toBe("Cancelada"));
  });
});

// FWB-04 AC2.
describe("pagar", () => {
  it("confirma, chama /pay com CSRF, mostra Paga e invalida o saldo", async () => {
    const { state, list, queryClient } = await renderDetail([DESPESA_CRIADA]);
    const requests = mockWrite("post", `/lancamentos/${ID}/pay`, () => {
      state.items = [{ ...DESPESA_CRIADA, status: "PAGA" }];
      return new HttpResponse(null, { status: 204 });
    });
    fireEvent.click(screen.getByRole("button", { name: "Pagar" }));
    await confirmIn("Confirmar pagamento");
    await waitFor(() => expect(requests).toEqual([{ body: null, csrf: "csrf-token-de-teste" }]));
    await waitFor(() => expect(badge()).toBe("Paga"));
    expect(list.count).toBe(2);
    expect(saldoInvalidated(queryClient)).toBe(true);
  });

  it("lancamento_nao_pode_ser_pago mostra a mensagem e atualiza a lista", async () => {
    const { list } = await renderDetail([DESPESA_CRIADA]);
    mockWrite("post", `/lancamentos/${ID}/pay`, () => problem(409, "lancamento_nao_pode_ser_pago"));
    fireEvent.click(screen.getByRole("button", { name: "Pagar" }));
    await confirmIn("Confirmar pagamento");
    expect(await screen.findByText("Este lançamento não está mais em aberto.")).toBeTruthy();
    await waitFor(() => expect(list.count).toBe(2));
  });
});

// FWB-04 AC3 e AC4.
describe("cancelar", () => {
  async function openCancel() {
    fireEvent.click(screen.getByRole("button", { name: "Cancelar lançamento" }));
    const alert = await screen.findByRole("alertdialog");
    return {
      alert: within(alert),
      confirm: () => within(alert).getByRole("button", { name: "Cancelar lançamento" }) as HTMLButtonElement,
    };
  }

  it.each([
    ["vazio", ""],
    ["só com espaços", "   \n\t "],
  ])("motivo %s mantém a confirmação desabilitada e não chama a API", async (_, text) => {
    await renderDetail([RECEITA_CRIADA]);
    const requests = mockWrite("post", `/lancamentos/${ID}/cancel`, () => new HttpResponse(null, { status: 204 }));
    const { alert, confirm } = await openCancel();
    change(alert.getByLabelText("Motivo do cancelamento"), text);
    expect(confirm().disabled).toBe(true);
    fireEvent.click(confirm());
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(requests).toEqual([]);
  });

  it("com motivo, envia reason, mostra Cancelada com o motivo e invalida o saldo", async () => {
    const { state, list, queryClient } = await renderDetail([RECEITA_RECEBIDA]);
    const requests = mockWrite("post", `/lancamentos/${ID}/cancel`, () => {
      state.items = [
        {
          ...RECEITA_RECEBIDA,
          status: "CANCELADA",
          motivo_cancelamento: "Lançado em duplicidade",
          cancelado_por: RECEITA_RECEBIDA.criado_por,
          cancelado_em: "2026-10-04T15:00:00Z",
        },
      ];
      return new HttpResponse(null, { status: 204 });
    });
    const { alert, confirm } = await openCancel();
    change(alert.getByLabelText("Motivo do cancelamento"), "  Lançado em duplicidade  ");
    expect(confirm().disabled).toBe(false);
    fireEvent.click(confirm());
    await waitFor(() =>
      expect(requests).toEqual([{ body: { reason: "Lançado em duplicidade" }, csrf: "csrf-token-de-teste" }]),
    );
    await waitFor(() => expect(badge()).toBe("Cancelada"));
    expect(detailValue("Motivo do cancelamento")).toBe("Lançado em duplicidade");
    expect(list.count).toBe(2);
    expect(saldoInvalidated(queryClient)).toBe(true);
  });

  it("lancamento_ja_cancelado mostra a mensagem e atualiza a lista", async () => {
    const { list } = await renderDetail([RECEITA_CRIADA]);
    mockWrite("post", `/lancamentos/${ID}/cancel`, () => problem(409, "lancamento_ja_cancelado"));
    const { alert, confirm } = await openCancel();
    change(alert.getByLabelText("Motivo do cancelamento"), "Erro de digitação");
    fireEvent.click(confirm());
    expect(await screen.findByText("Este lançamento já foi cancelado.")).toBeTruthy();
    await waitFor(() => expect(list.count).toBe(2));
  });

  it("motivo_obrigatorio da API mostra Informe o motivo.", async () => {
    await renderDetail([RECEITA_CRIADA]);
    mockWrite("post", `/lancamentos/${ID}/cancel`, () => problem(422, "motivo_obrigatorio"));
    const { alert, confirm } = await openCancel();
    change(alert.getByLabelText("Motivo do cancelamento"), "x");
    fireEvent.click(confirm());
    expect(await screen.findByText("Informe o motivo.")).toBeTruthy();
  });
});

// FWB-04 AC5 e AC6.
describe("devolução", () => {
  async function openDevolucao() {
    fireEvent.click(screen.getByRole("button", { name: "Registrar devolução" }));
    await waitFor(() => expect(formDialog()).not.toBeNull());
    const form = dialog();
    await waitFor(() => expect(form.getByRole("option", { name: "Devoluções" })).toBeTruthy());
    return form;
  }

  it("envia devolucao_de_id da receita, conta de despesa, valores e forma, com CSRF", async () => {
    const { list } = await renderDetail([RECEITA_RECEBIDA]);
    const requests = mockWrite("post", "/lancamentos/devolucoes", () =>
      HttpResponse.json(lancamento({ id: "devolucao", tipo: "DESPESA", devolucao_de_id: ID }), { status: 201 }),
    );
    const form = await openDevolucao();
    const contaOptions = Array.from((form.getByLabelText("Conta") as HTMLSelectElement).options).map(
      (o) => o.textContent,
    );
    expect(contaOptions).toEqual(["Selecione a conta", "Devoluções", "Transporte"]);
    change(form.getByLabelText("Conta"), CONTA_DESPESA.id);
    change(form.getByLabelText("Valor bruto"), "50,00");
    change(form.getByLabelText("Taxa"), "1,00");
    change(form.getByLabelText("Forma de pagamento"), "PIX");
    fireEvent.click(form.getByRole("button", { name: "Registrar devolução" }));
    await waitFor(() => expect(requests).toHaveLength(1));
    expect(requests[0]).toEqual({
      body: {
        devolucao_de_id: ID,
        conta_id: CONTA_DESPESA.id,
        valor_bruto_cents: 5000,
        taxa_cents: 100,
        forma_pagamento: "PIX",
      },
      csrf: "csrf-token-de-teste",
    });
    await waitFor(() => expect(formDialog()).toBeNull());
    await waitFor(() => expect(list.count).toBe(2));
  });

  it("devolucao_invalida mostra a mensagem", async () => {
    await renderDetail([RECEITA_RECEBIDA]);
    mockWrite("post", "/lancamentos/devolucoes", () => problem(422, "devolucao_invalida"));
    const form = await openDevolucao();
    change(form.getByLabelText("Conta"), CONTA_DESPESA.id);
    change(form.getByLabelText("Valor bruto"), "50,00");
    change(form.getByLabelText("Forma de pagamento"), "PIX");
    fireEvent.click(form.getByRole("button", { name: "Registrar devolução" }));
    expect(await form.findByText("Só é possível devolver uma receita recebida.")).toBeTruthy();
  });
});

// FWB-04 AC6-7 e casos de borda: cada ação só no status e com a permissão certos.
describe("ações visíveis", () => {
  const ALL = ["Receber", "Pagar", "Editar", "Registrar devolução", "Cancelar lançamento"];

  function visibleActions() {
    return ALL.filter((name) => screen.queryByRole("button", { name }) !== null);
  }

  it.each([
    ["receita CRIADA", RECEITA_CRIADA, ["Receber", "Editar", "Cancelar lançamento"]],
    ["despesa CRIADA", DESPESA_CRIADA, ["Pagar", "Editar", "Cancelar lançamento"]],
    ["receita RECEBIDA", RECEITA_RECEBIDA, ["Registrar devolução", "Cancelar lançamento"]],
    ["despesa PAGA", lancamento({ ...DESPESA_CRIADA, status: "PAGA" }), ["Cancelar lançamento"]],
    ["receita CANCELADA", lancamento({ ...RECEITA_CRIADA, status: "CANCELADA" }), []],
  ])("tesouraria em %s", async (_, item, expected) => {
    await renderDetail([item]);
    expect(visibleActions()).toEqual(expected);
  });

  it.each([RECEITA_CRIADA, DESPESA_CRIADA, RECEITA_RECEBIDA])(
    "só leitura não vê nenhuma ação ($tipo $status)",
    async (item) => {
      await renderDetail([item], PERMISSOES_LEITURA);
      expect(visibleActions()).toEqual([]);
    },
  );

  // FND-04 AC3: cada ação depende só da permissão que a spec nomeia para ela.
  // `outras` são as ações que o status do lançamento também permite.
  const ISOLAMENTO: [string, string, Lancamento, string[]][] = [
    ["Receber", "financeiro:lancamento:receive", RECEITA_CRIADA, ["Editar", "Cancelar lançamento"]],
    ["Pagar", "financeiro:lancamento:pay", DESPESA_CRIADA, ["Editar", "Cancelar lançamento"]],
    ["Editar", "financeiro:lancamento:update", RECEITA_CRIADA, ["Receber", "Cancelar lançamento"]],
    ["Editar", "financeiro:lancamento:update", DESPESA_CRIADA, ["Pagar", "Cancelar lançamento"]],
    ["Cancelar lançamento", "financeiro:lancamento:cancel", RECEITA_CRIADA, ["Receber", "Editar"]],
    ["Cancelar lançamento", "financeiro:lancamento:cancel", DESPESA_CRIADA, ["Pagar", "Editar"]],
    ["Cancelar lançamento", "financeiro:lancamento:cancel", RECEITA_RECEBIDA, ["Registrar devolução"]],
    ["Registrar devolução", "financeiro:lancamento:create", RECEITA_RECEBIDA, ["Cancelar lançamento"]],
  ];

  it.each(ISOLAMENTO)("leituras + só %s (%s) vê só essa ação", async (action, permission, item) => {
    await renderDetail([item], leituraMais(permission));
    await waitFor(() => expect(visibleActions()).toEqual([action]));
  });

  it.each(ISOLAMENTO)("todas as outras escritas, sem a de %s (%s), não vê essa ação", async (action, permission, item, outras) => {
    await renderDetail([item], tesourariaSem(permission));
    const expected = ALL.filter((name) => outras.includes(name));
    // As outras ações aparecem: a sessão já carregou e a ausência é da permissão.
    await waitFor(() => expect(visibleActions()).toEqual(expected));
    expect(screen.queryByRole("button", { name: action })).toBeNull();
  });
});
