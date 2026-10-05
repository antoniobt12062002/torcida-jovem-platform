import { fireEvent, screen, waitFor } from "@testing-library/react";
import { HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, ids, lancamento, problem } from "@/test/msw/fixtures";

import FinanceiroLancamentoPage from "@/app/(app)/financeiro/lancamentos/[id]/page";

import { LancamentoDetail } from "./lancamento-detail";
import {
  change,
  CONTA_DESPESA,
  CONTA_RECEITA,
  describedByText,
  detailValue,
  dialog,
  formDialog,
  mockContas,
  mockLancamentos,
  mockWrite,
  PERMISSOES_LEITURA,
  PERMISSOES_TESOURARIA,
} from "./test-helpers";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const ID = "00000000-0000-4000-8000-000000000221";

const CRIADA = lancamento({
  id: ID,
  tipo: "RECEITA",
  conta_id: CONTA_RECEITA.id,
  valor_bruto_cents: 10000,
  taxa_cents: 250,
  valor_liquido_cents: 9750,
  forma_pagamento: "CARTAO",
  status: "CRIADA",
  criado_por: ids.user,
  criado_em: "2026-10-04T12:00:00Z",
  atualizado_em: "2026-10-04T13:30:00Z",
});

async function renderDetail(items = [CRIADA], permissions = PERMISSOES_TESOURARIA, id = ID) {
  mockMe(authContext({ permissions }));
  mockContas();
  const list = mockLancamentos(items);
  renderWithSession(<LancamentoDetail id={id} />);
  return list;
}

// FWB-03 AC6.
describe("detalhe", () => {
  it("mostra todos os campos do lançamento", async () => {
    await renderDetail();
    expect(await screen.findByRole("heading", { name: `Receita ${ID.slice(0, 8)}` })).toBeTruthy();
    await waitFor(() => expect(detailValue("Conta")).toBe("Mensalidades"));
    expect(detailValue("Tipo")).toBe("Receita");
    expect(detailValue("Valor bruto")).toBe("R$ 100,00");
    expect(detailValue("Taxa")).toBe("R$ 2,50");
    expect(detailValue("Valor líquido")).toBe("R$ 97,50");
    expect(detailValue("Forma de pagamento")).toBe("Cartão");
    expect(detailValue("Criado por")).toBe("você");
    expect(detailValue("Criado em")).toBe("04/10/2026 09:00");
    expect(detailValue("Atualizado em")).toBe("04/10/2026 10:30");
    expect(screen.queryByText("Motivo do cancelamento")).toBeNull();
    expect(screen.queryByText("Devolução da receita")).toBeNull();
  });

  it("cancelado: mostra motivo, autor e data do cancelamento", async () => {
    await renderDetail([
      lancamento({
        ...CRIADA,
        status: "CANCELADA",
        motivo_cancelamento: "Lançado em duplicidade",
        cancelado_por: ids.otherUser,
        cancelado_em: "2026-10-05T15:00:00Z",
      }),
    ]);
    expect(await screen.findByText("Motivo do cancelamento")).toBeTruthy();
    expect(detailValue("Motivo do cancelamento")).toBe("Lançado em duplicidade");
    expect(detailValue("Cancelado por")).toBe(ids.otherUser.slice(0, 8));
    expect(detailValue("Cancelado em")).toBe("05/10/2026 12:00");
  });

  it("devolução: mostra a referência à receita devolvida", async () => {
    const receitaId = "00000000-0000-4000-8000-000000000299";
    await renderDetail([
      lancamento({ ...CRIADA, tipo: "DESPESA", conta_id: CONTA_DESPESA.id, devolucao_de_id: receitaId }),
    ]);
    const link = await screen.findByRole("link", { name: receitaId.slice(0, 8) });
    expect(link.getAttribute("href")).toBe(`/financeiro/lancamentos/${receitaId}`);
    expect(detailValue("Devolução da receita")).toBe(receitaId.slice(0, 8));
  });

  it("id fora da lista mostra Lançamento não encontrado", async () => {
    await renderDetail([CRIADA], PERMISSOES_TESOURARIA, "00000000-0000-4000-8000-000000000999");
    expect(await screen.findByText(/^Lançamento não encontrado/)).toBeTruthy();
  });
});

async function openEdit() {
  fireEvent.click(await screen.findByRole("button", { name: "Editar" }));
  await waitFor(() => expect(formDialog()).not.toBeNull());
  return dialog();
}

// FWB-03 AC3 e AC4.
describe("edição", () => {
  it("edita um lançamento CRIADA com PUT, valores em centavos e CSRF", async () => {
    const list = await renderDetail();
    const requests = mockWrite("put", `/lancamentos/${ID}`, () =>
      HttpResponse.json(lancamento({ ...CRIADA, valor_bruto_cents: 20000 })),
    );
    const form = await openEdit();
    await waitFor(() => expect((form.getByLabelText("Conta") as HTMLSelectElement).value).toBe(CONTA_RECEITA.id));
    expect((form.getByLabelText("Valor bruto") as HTMLInputElement).value).toBe("100,00");
    change(form.getByLabelText("Valor bruto"), "200,00");
    change(form.getByLabelText("Taxa"), "5,00");
    expect(screen.getByTestId("previa-liquido").textContent).toBe("R$ 195,00");
    change(form.getByLabelText("Forma de pagamento"), "PIX");
    fireEvent.click(form.getByRole("button", { name: "Salvar alterações" }));

    await waitFor(() => expect(requests).toHaveLength(1));
    expect(requests[0]).toEqual({
      body: {
        conta_id: CONTA_RECEITA.id,
        valor_bruto_cents: 20000,
        taxa_cents: 500,
        forma_pagamento: "PIX",
      },
      csrf: "csrf-token-de-teste",
    });
    await waitFor(() => expect(formDialog()).toBeNull());
    await waitFor(() => expect(list.count).toBe(2));
  });

  it.each(["RECEBIDA", "PAGA", "CANCELADA"] as const)("não oferece editar em %s", async (status) => {
    await renderDetail([lancamento({ ...CRIADA, tipo: status === "PAGA" ? "DESPESA" : "RECEITA", status })]);
    await screen.findByText("Valor bruto", { selector: "dt" });
    expect(screen.queryByRole("button", { name: "Editar" })).toBeNull();
  });

  it("lancamento_imutavel explica e atualiza a lista", async () => {
    const list = await renderDetail();
    mockWrite("put", `/lancamentos/${ID}`, () => problem(409, "lancamento_imutavel"));
    const form = await openEdit();
    await waitFor(() => expect((form.getByLabelText("Conta") as HTMLSelectElement).value).toBe(CONTA_RECEITA.id));
    fireEvent.click(form.getByRole("button", { name: "Salvar alterações" }));
    expect(await form.findByText("Só lançamentos em aberto podem ser editados.", { selector: "[data-slot=alert-description]" })).toBeTruthy();
    await waitFor(() => expect(list.count).toBe(2));
  });

  it("conta_invalida aparece junto do campo Conta", async () => {
    await renderDetail();
    mockWrite("put", `/lancamentos/${ID}`, () => problem(422, "conta_invalida"));
    const form = await openEdit();
    await waitFor(() => expect((form.getByLabelText("Conta") as HTMLSelectElement).value).toBe(CONTA_RECEITA.id));
    fireEvent.click(form.getByRole("button", { name: "Salvar alterações" }));
    const message = "Escolha uma conta ativa do mesmo tipo do lançamento.";
    expect(await form.findByText(message)).toBeTruthy();
    expect(describedByText(form.getByLabelText("Conta"))).toContain(message);
  });
});

// Casos de borda: só leitura e acesso à página.
describe("permissões", () => {
  it("só leitura não vê editar em lançamento CRIADA", async () => {
    await renderDetail([CRIADA], PERMISSOES_LEITURA);
    await screen.findByText("Valor bruto", { selector: "dt" });
    expect(screen.queryByRole("button", { name: "Editar" })).toBeNull();
  });

  it("a página exige financeiro:lancamento:read", async () => {
    mockMe(authContext({ permissions: ["financeiro:saldo:read"] }));
    renderWithSession(await FinanceiroLancamentoPage({ params: Promise.resolve({ id: ID }) }));
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
  });

  it("a página mostra o detalhe do id da rota", async () => {
    mockMe(authContext({ permissions: PERMISSOES_LEITURA }));
    mockContas();
    mockLancamentos([CRIADA]);
    renderWithSession(await FinanceiroLancamentoPage({ params: Promise.resolve({ id: ID }) }));
    expect(await screen.findByRole("heading", { name: `Receita ${ID.slice(0, 8)}` })).toBeTruthy();
  });
});
