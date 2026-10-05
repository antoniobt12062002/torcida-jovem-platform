import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, lancamento, problem } from "@/test/msw/fixtures";

import FinanceiroLancamentosPage from "@/app/(app)/financeiro/lancamentos/page";

import { LancamentosPage } from "./lancamentos-page";
import {
  change,
  CONTA_DESPESA,
  CONTA_DESPESA_2,
  CONTA_RECEITA,
  describedByText,
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

const RECEITA = lancamento({
  id: "00000000-0000-4000-8000-000000000211",
  tipo: "RECEITA",
  conta_id: CONTA_RECEITA.id,
  valor_bruto_cents: 10000,
  taxa_cents: 250,
  valor_liquido_cents: 9750,
  forma_pagamento: "PIX",
  status: "CRIADA",
  criado_em: "2026-10-04T12:00:00Z",
});
const RECEITA_RECEBIDA = lancamento({
  id: "00000000-0000-4000-8000-000000000212",
  tipo: "RECEITA",
  conta_id: CONTA_RECEITA.id,
  valor_bruto_cents: 5000,
  valor_liquido_cents: 5000,
  forma_pagamento: "DINHEIRO",
  status: "RECEBIDA",
  criado_em: "2026-09-10T12:00:00Z",
});
const DESPESA_PAGA = lancamento({
  id: "00000000-0000-4000-8000-000000000213",
  tipo: "DESPESA",
  conta_id: CONTA_DESPESA_2.id,
  valor_bruto_cents: 3000,
  valor_liquido_cents: 3000,
  forma_pagamento: "TRANSFERENCIA",
  status: "PAGA",
  criado_em: "2026-09-01T12:00:00Z",
});

async function renderPage(permissions = PERMISSOES_TESOURARIA, items = [RECEITA, RECEITA_RECEBIDA, DESPESA_PAGA]) {
  mockMe(authContext({ permissions }));
  mockContas();
  const list = mockLancamentos(items);
  renderWithSession(<LancamentosPage />);
  await screen.findByRole("heading", { name: "Lançamentos" });
  return list;
}

function rows() {
  return screen.queryAllByTestId("lancamento-row");
}

function rowTexts() {
  return rows().map((r) => within(r).getAllByRole("cell").map((c) => c.textContent));
}

// FWB-03 AC1.
describe("lista", () => {
  it("chama GET /lancamentos e mostra os campos, do mais novo para o mais antigo", async () => {
    const list = await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    expect(list.count).toBe(1);
    // O nome da conta chega pela consulta de contas.
    await waitFor(() =>
      expect(rowTexts()[0]).toEqual([
        "04/10/2026",
        "Receita",
        "Mensalidades",
        "R$ 100,00",
        "R$ 2,50",
        "R$ 97,50",
        "Pix",
        "Criada",
        "Detalhes",
      ]),
    );
    expect(rowTexts()[2]).toEqual([
      "01/09/2026",
      "Despesa",
      "Transporte",
      "R$ 30,00",
      "R$ 0,00",
      "R$ 30,00",
      "Transferência",
      "Paga",
      "Detalhes",
    ]);
    const link = within(rows()[0]).getByRole("link", { name: "Detalhes" });
    expect(link.getAttribute("href")).toBe(`/financeiro/lancamentos/${RECEITA.id}`);
  });

  it("filtra por status no cliente, sem nova chamada", async () => {
    const list = await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    change(screen.getByLabelText("Status"), "RECEBIDA");
    expect(rowTexts().map((r) => r[7])).toEqual(["Recebida"]);
    change(screen.getByLabelText("Status"), "");
    expect(rows()).toHaveLength(3);
    expect(list.count).toBe(1);
  });

  it("filtra por tipo, conta e período", async () => {
    await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    change(screen.getByLabelText("Tipo"), "DESPESA");
    expect(rowTexts().map((r) => r[1])).toEqual(["Despesa"]);
    fireEvent.click(screen.getByRole("button", { name: "Limpar filtros" }));
    expect(rows()).toHaveLength(3);

    change(screen.getByLabelText("Conta"), CONTA_RECEITA.id);
    expect(rowTexts().map((r) => r[0])).toEqual(["04/10/2026", "10/09/2026"]);
    change(screen.getByLabelText("Conta"), "");

    change(screen.getByLabelText("Criado de"), "2026-09-05");
    change(screen.getByLabelText("Criado até"), "2026-09-30");
    expect(rowTexts().map((r) => r[0])).toEqual(["10/09/2026"]);
  });

  it("filtro sem resultado explica o vazio", async () => {
    await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    change(screen.getByLabelText("Status"), "CANCELADA");
    expect(screen.getByText("Nenhum lançamento corresponde aos filtros.")).toBeTruthy();
  });

  it("sem lançamentos, mostra o estado vazio", async () => {
    await renderPage(PERMISSOES_TESOURARIA, []);
    expect(await screen.findByText("Nenhum lançamento registrado.")).toBeTruthy();
  });
});

async function openCreate() {
  fireEvent.click(await screen.findByRole("button", { name: "Novo lançamento" }));
  await waitFor(() => expect(formDialog()).not.toBeNull());
}

// FWB-03 AC2 e premissas de contas oferecidas e prévia do líquido.
describe("criação", () => {
  it("envia os valores em centavos inteiros, com CSRF, e atualiza a lista", async () => {
    const list = await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    const requests = mockWrite("post", "/lancamentos", () =>
      HttpResponse.json(lancamento({ id: "novo" }), { status: 201 }),
    );
    await openCreate();
    const form = dialog();
    change(form.getByLabelText("Conta"), CONTA_RECEITA.id);
    change(form.getByLabelText("Valor bruto"), "150,00");
    change(form.getByLabelText("Taxa"), "2,50");
    expect(screen.getByTestId("previa-liquido").textContent).toBe("R$ 147,50");
    change(form.getByLabelText("Forma de pagamento"), "PIX");
    fireEvent.click(form.getByRole("button", { name: "Criar lançamento" }));

    await waitFor(() => expect(requests).toHaveLength(1));
    expect(requests[0]).toEqual({
      body: {
        tipo: "RECEITA",
        conta_id: CONTA_RECEITA.id,
        valor_bruto_cents: 15000,
        taxa_cents: 250,
        forma_pagamento: "PIX",
      },
      csrf: "csrf-token-de-teste",
    });
    await waitFor(() => expect(formDialog()).toBeNull());
    await waitFor(() => expect(list.count).toBe(2));
    expect(await screen.findByText("Lançamento criado.")).toBeTruthy();
  });

  it("despesa oferece só contas ativas de despesa", async () => {
    await renderPage();
    await openCreate();
    const form = dialog();
    const contaOptions = () =>
      Array.from((form.getByLabelText("Conta") as HTMLSelectElement).options).map((o) => o.textContent);
    await waitFor(() => expect(contaOptions()).toEqual(["Selecione a conta", "Mensalidades"]));
    change(form.getByLabelText("Tipo"), "DESPESA");
    expect(contaOptions()).toEqual(["Selecione a conta", "Devoluções", "Transporte"]);
  });

  it("cria despesa com o tipo escolhido", async () => {
    await renderPage();
    const requests = mockWrite("post", "/lancamentos", () =>
      HttpResponse.json(lancamento({ id: "novo", tipo: "DESPESA" }), { status: 201 }),
    );
    await openCreate();
    const form = dialog();
    change(form.getByLabelText("Tipo"), "DESPESA");
    await waitFor(() => expect(form.getByRole("option", { name: "Devoluções" })).toBeTruthy());
    change(form.getByLabelText("Conta"), CONTA_DESPESA.id);
    change(form.getByLabelText("Valor bruto"), "1.234,56");
    change(form.getByLabelText("Taxa"), "0");
    change(form.getByLabelText("Forma de pagamento"), "DINHEIRO");
    fireEvent.click(form.getByRole("button", { name: "Criar lançamento" }));
    await waitFor(() => expect(requests).toHaveLength(1));
    expect(requests[0].body).toEqual({
      tipo: "DESPESA",
      conta_id: CONTA_DESPESA.id,
      valor_bruto_cents: 123456,
      taxa_cents: 0,
      forma_pagamento: "DINHEIRO",
    });
  });

  it("sem conta, valor ou forma de pagamento não chama a API", async () => {
    await renderPage();
    const requests = mockWrite("post", "/lancamentos", () => HttpResponse.json({}, { status: 201 }));
    await openCreate();
    fireEvent.click(dialog().getByRole("button", { name: "Criar lançamento" }));
    expect(await dialog().findByText("Escolha a conta.")).toBeTruthy();
    expect(dialog().getByText("Informe o valor bruto.")).toBeTruthy();
    expect(dialog().getByText("Escolha a forma de pagamento.")).toBeTruthy();
    expect(requests).toEqual([]);
  });
});

// FWB-03 AC5.
describe("erros da API junto do campo", () => {
  it.each([
    ["conta_invalida", "Conta", "Escolha uma conta ativa do mesmo tipo do lançamento."],
    ["lancamento_tipo_incompativel", "Conta", "O tipo do lançamento não combina com o tipo da conta."],
    ["amount_out_of_range", "Valor bruto", "Valor fora da faixa permitida."],
  ])("%s", async (code, label, message) => {
    const list = await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    mockWrite("post", "/lancamentos", () => problem(422, code));
    await openCreate();
    const form = dialog();
    change(form.getByLabelText("Conta"), CONTA_RECEITA.id);
    change(form.getByLabelText("Valor bruto"), "10,00");
    change(form.getByLabelText("Forma de pagamento"), "PIX");
    fireEvent.click(form.getByRole("button", { name: "Criar lançamento" }));
    expect(await form.findByText(message)).toBeTruthy();
    const field = form.getByLabelText(label);
    expect(field.getAttribute("aria-invalid")).toBe("true");
    expect(describedByText(field)).toContain(message);
    expect(formDialog()).not.toBeNull();
    expect(list.count).toBe(1);
  });
});

// Casos de borda: sessão só de leitura não vê escrita.
describe("permissões", () => {
  it("só leitura não vê Novo lançamento", async () => {
    await renderPage(PERMISSOES_LEITURA);
    await waitFor(() => expect(rows()).toHaveLength(3));
    expect(screen.queryByRole("button", { name: "Novo lançamento" })).toBeNull();
  });

  it("a página exige financeiro:lancamento:read", async () => {
    mockMe(authContext({ permissions: ["financeiro:saldo:read"] }));
    renderWithSession(<FinanceiroLancamentosPage />);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
  });

  it("com a permissão, a página mostra a lista", async () => {
    mockMe(authContext({ permissions: PERMISSOES_LEITURA }));
    mockContas();
    mockLancamentos([RECEITA]);
    renderWithSession(<FinanceiroLancamentosPage />);
    expect(await screen.findByRole("heading", { name: "Lançamentos" })).toBeTruthy();
    await waitFor(() => expect(rows()).toHaveLength(1));
  });
});
