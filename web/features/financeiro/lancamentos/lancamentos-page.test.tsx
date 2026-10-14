import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { mockViewport } from "@/test/media";
import { authContext, type Lancamento, lancamento, problem } from "@/test/msw/fixtures";

import FinanceiroLancamentosPage from "@/app/(app)/financeiro/lancamentos/page";

import { STATUS_DEF } from "./labels";
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
  SESSOES_SEM_LANCAMENTO_READ,
  leituraMais,
  tesourariaSem,
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

// As linhas de dados da tabela "Lançamentos" (DataList), sem a do cabeçalho.
function rows() {
  const table = screen.queryByRole("table", { name: "Lançamentos" });
  return table ? within(table).getAllByRole("row").slice(1) : [];
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
    // DSP-05 AC10: o link "Detalhes" tem nome que identifica o lançamento.
    const link = within(rows()[0]).getByRole("link", { name: `Detalhes lançamento Receita ${RECEITA.id.slice(0, 8)}` });
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

  // FWB-03 AC2 e FND-04 AC3: "Novo lançamento" depende só de lancamento:create.
  // A rota só mostra o título com a sessão carregada, então a ausência do
  // botão não é efeito de sessão ainda pendente.
  async function renderRoute(permissions: string[]) {
    mockMe(authContext({ permissions }));
    const contas = mockContas();
    const list = mockLancamentos([RECEITA]);
    renderWithSession(<FinanceiroLancamentosPage />);
    return { contas, list };
  }

  it("leituras + só lancamento:create vê Novo lançamento", async () => {
    await renderRoute(leituraMais("financeiro:lancamento:create"));
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(await screen.findByRole("button", { name: "Novo lançamento" })).toBeTruthy();
  });

  it("todas as outras escritas, sem lancamento:create, não vê Novo lançamento", async () => {
    await renderRoute(tesourariaSem("financeiro:lancamento:create"));
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(screen.queryByRole("button", { name: "Novo lançamento" })).toBeNull();
  });

  // FND-04 AC4: a rota depende de lancamento:read, não das leituras vizinhas.
  it.each(SESSOES_SEM_LANCAMENTO_READ)("sem lancamento:read, %s: Sem acesso e nenhuma consulta", async (_, permissions) => {
    const { contas, list } = await renderRoute(permissions);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(screen.queryByRole("heading", { name: "Lançamentos" })).toBeNull();
    expect(list.count).toBe(0);
    expect(contas.count).toBe(0);
  });

  it("só com lancamento:read, a página mostra a lista", async () => {
    const { list } = await renderRoute(["financeiro:lancamento:read"]);
    expect(await screen.findByRole("heading", { name: "Lançamentos" })).toBeTruthy();
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(screen.queryByRole("heading", { name: "Sem acesso" })).toBeNull();
    expect(list.count).toBe(1);
  });
});

// DSM-03 AC4 e DSP-05 AC2 a AC4: tipo, status, conta e período na URL.
describe("filtros na URL", () => {
  const params = () => Object.fromEntries(new URLSearchParams(window.location.search));

  it("cada filtro vai para a query string, e Limpar filtros a esvazia", async () => {
    await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    await screen.findByRole("option", { name: "Mensalidades" });
    change(screen.getByLabelText("Tipo"), "RECEITA");
    change(screen.getByLabelText("Status"), "RECEBIDA");
    change(screen.getByLabelText("Conta"), CONTA_RECEITA.id);
    change(screen.getByLabelText("Criado de"), "2026-09-01");
    change(screen.getByLabelText("Criado até"), "2026-09-30");
    await waitFor(() =>
      expect(params()).toEqual({
        tipo: "RECEITA",
        status: "RECEBIDA",
        conta: CONTA_RECEITA.id,
        de: "2026-09-01",
        ate: "2026-09-30",
      }),
    );
    expect(rowTexts().map((r) => r[0])).toEqual(["10/09/2026"]);
    fireEvent.click(screen.getByRole("button", { name: "Limpar filtros" }));
    await waitFor(() => expect(window.location.search).toBe(""));
    expect(rows()).toHaveLength(3);
  });

  it("abre com os filtros da URL", async () => {
    setLocation("/financeiro/lancamentos?tipo=RECEITA&status=RECEBIDA");
    await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(rowTexts()[0][7]).toBe("Recebida");
    expect((screen.getByLabelText("Tipo") as HTMLSelectElement).value).toBe("RECEITA");
    expect((screen.getByLabelText("Status") as HTMLSelectElement).value).toBe("RECEBIDA");
    expect(screen.getByText("1 lançamento")).toBeTruthy();
  });

  it("abre com o período da URL", async () => {
    setLocation("/financeiro/lancamentos?de=2026-09-05&ate=2026-09-30");
    await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(rowTexts()[0][0]).toBe("10/09/2026");
    expect((screen.getByLabelText("Criado de") as HTMLInputElement).value).toBe("2026-09-05");
  });

  it("valor inválido na URL é ignorado, sem erro", async () => {
    setLocation("/financeiro/lancamentos?tipo=OUTRO&status=ESTORNADA&de=2026-02-30");
    await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    expect((screen.getByLabelText("Tipo") as HTMLSelectElement).value).toBe("");
    expect((screen.getByLabelText("Status") as HTMLSelectElement).value).toBe("");
    expect((screen.getByLabelText("Criado de") as HTMLInputElement).value).toBe("");
    expect(screen.queryByRole("button", { name: "Limpar filtros" })).toBeNull();
  });

  it("conta da URL que existe filtra a lista", async () => {
    setLocation(`/financeiro/lancamentos?conta=${CONTA_RECEITA.id}`);
    await renderPage();
    await screen.findByRole("option", { name: "Mensalidades" });
    await waitFor(() => expect(rows()).toHaveLength(2));
    expect((screen.getByLabelText("Conta") as HTMLSelectElement).value).toBe(CONTA_RECEITA.id);
  });

  it("conta da URL que não existe é ignorada", async () => {
    setLocation("/financeiro/lancamentos?conta=00000000-0000-4000-8000-000000000999");
    await renderPage();
    await screen.findByRole("option", { name: "Mensalidades" });
    await waitFor(() => expect(rows()).toHaveLength(3));
    expect((screen.getByLabelText("Conta") as HTMLSelectElement).value).toBe("");
    expect(screen.queryByRole("button", { name: "Limpar filtros" })).toBeNull();
    expect(screen.getByText("3 lançamentos")).toBeTruthy();
  });
});

// DSM-03 AC3 e DD-3: status com rótulo e ícone distintos; CANCELADA neutra.
describe("status", () => {
  it("CRIADA, RECEBIDA, PAGA e CANCELADA têm rótulo e ícone distintos entre si", () => {
    const defs = Object.values(STATUS_DEF);
    expect(defs).toHaveLength(4);
    expect(new Set(defs.map((d) => d.label)).size).toBe(4);
    expect(new Set(defs.map((d) => d.icon)).size).toBe(4);
    expect(STATUS_DEF.CANCELADA.tone).toBe("neutral");
    expect(STATUS_DEF.CRIADA.tone).toBe("neutral");
  });

  it("a lista mostra cada status com rótulo, ícone e tom", async () => {
    const CANCELADA = lancamento({ ...DESPESA_PAGA, id: "00000000-0000-4000-8000-000000000214", status: "CANCELADA" });
    await renderPage(PERMISSOES_TESOURARIA, [RECEITA, RECEITA_RECEBIDA, DESPESA_PAGA, CANCELADA]);
    await waitFor(() => expect(rows()).toHaveLength(4));
    const badges = rows().map((r) => r.querySelector<HTMLElement>('[data-slot="badge"]'));
    expect(badges.map((b) => b?.textContent)).toEqual(["Criada", "Recebida", "Paga", "Cancelada"]);
    expect(badges.map((b) => b?.dataset.tone)).toEqual(["neutral", "success", "success", "neutral"]);
    const icons = badges.map((b) => b?.querySelector("svg")?.getAttribute("class"));
    for (const icon of icons) expect(icon).toBeTruthy();
    expect(new Set(icons).size).toBe(4);
  });
});

// DSP-05 AC12 e AC13: 50 por vez; o filtro considera todos os registros.
describe("mostrar mais", () => {
  /** `n` lançamentos de receita, do mais antigo (índice 0) ao mais novo. */
  function many(n: number, overrides: (index: number) => Partial<Lancamento> = () => ({})): Lancamento[] {
    return Array.from({ length: n }, (_, index) =>
      lancamento({
        id: `${String(index).padStart(8, "0")}-0000-4000-8000-000000000000`,
        conta_id: CONTA_RECEITA.id,
        criado_em: new Date(Date.UTC(2026, 0, 1) + index * 3_600_000).toISOString(),
        ...overrides(index),
      }),
    );
  }

  it("com 51, mostra 50 e Mostrar mais acrescenta o restante", async () => {
    await renderPage(PERMISSOES_TESOURARIA, many(51));
    await waitFor(() => expect(rows()).toHaveLength(50));
    expect(screen.getByText("Mostrando 50 de 51 lançamentos")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Mostrar mais" }));
    expect(rows()).toHaveLength(51);
    expect(screen.queryByRole("button", { name: "Mostrar mais" })).toBeNull();
  });

  it("com 120, acrescenta 50 por vez", async () => {
    await renderPage(PERMISSOES_TESOURARIA, many(120));
    await waitFor(() => expect(rows()).toHaveLength(50));
    fireEvent.click(screen.getByRole("button", { name: "Mostrar mais" }));
    expect(rows()).toHaveLength(100);
    fireEvent.click(screen.getByRole("button", { name: "Mostrar mais" }));
    expect(rows()).toHaveLength(120);
  });

  it("com exatamente 50, o controle não aparece", async () => {
    await renderPage(PERMISSOES_TESOURARIA, many(50));
    await waitFor(() => expect(rows()).toHaveLength(50));
    expect(screen.queryByRole("button", { name: "Mostrar mais" })).toBeNull();
    // O bloco inteiro some, não só o botão: nem a contagem "Mostrando…" aparece.
    expect(document.querySelector('[data-slot="show-more"]')).toBeNull();
    expect(screen.queryByText(/^Mostrando /)).toBeNull();
  });

  it("o filtro considera todos os registros, não só os 50 visíveis", async () => {
    // O mais antigo (fora dos 50 primeiros) é o único recebido.
    await renderPage(PERMISSOES_TESOURARIA, many(60, (index) => (index === 0 ? { status: "RECEBIDA" } : {})));
    await waitFor(() => expect(rows()).toHaveLength(50));
    change(screen.getByLabelText("Status"), "RECEBIDA");
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(rowTexts()[0][7]).toBe("Recebida");
    expect(screen.getByText("1 lançamento")).toBeTruthy();
  });
});

// DSM-03 AC5 e DSP-05 AC7: abaixo de 768 px, itens empilhados.
describe("tela estreita", () => {
  it("cada item mostra data, tipo, conta, valor líquido e status", async () => {
    mockViewport({ width: 390 });
    await renderPage(PERMISSOES_TESOURARIA, [RECEITA]);
    const list = await screen.findByRole("list", { name: "Lançamentos" });
    expect(screen.queryByRole("table")).toBeNull();
    const [item] = within(list).getAllByRole("listitem");
    await waitFor(() => expect(item.querySelector('[data-slot="data-list-title"]')?.textContent).toBe("Mensalidades"));
    const fields = Array.from(item.querySelectorAll('[data-slot="data-list-field"]')).map((f) => f.textContent);
    expect(fields).toEqual(["Criado em: 04/10/2026", "Tipo: Receita"]);
    expect(item.querySelector('[data-slot="data-list-value"]')?.textContent).toBe("Valor líquido: R$ 97,50");
    expect(item.querySelector('[data-slot="data-list-meta"]')?.textContent).toBe("Criada");
    const link = within(item).getByRole("link", { name: `Detalhes lançamento Receita ${RECEITA.id.slice(0, 8)}` });
    expect(link.getAttribute("href")).toBe(`/financeiro/lancamentos/${RECEITA.id}`);
  });
});

// DSP-05 AC10: nenhum link de detalhe repete o nome entre linhas.
describe("nome do link de detalhe", () => {
  it("identifica cada lançamento", async () => {
    const items = [
      lancamento({ ...RECEITA, id: "aaaaaaaa-0000-4000-8000-000000000001" }),
      lancamento({ ...DESPESA_PAGA, id: "bbbbbbbb-0000-4000-8000-000000000002" }),
    ];
    await renderPage(PERMISSOES_TESOURARIA, items);
    await waitFor(() => expect(rows()).toHaveLength(2));
    expect(rows().map((r) => within(r).getByRole("link").getAttribute("aria-label"))).toEqual([
      "Detalhes lançamento Receita aaaaaaaa",
      "Detalhes lançamento Despesa bbbbbbbb",
    ]);
  });
});

// Smoke da rodada 6, D1: de 768 a 1440 px a tabela cabe no contêiner. Forma de
// pagamento sai junto com as colunas de prioridade 3, e a conta quebra (até
// dentro da palavra) em vez de alargar a tabela.
describe("colunas por largura", () => {
  function header(name: string) {
    const table = screen.getByRole("table", { name: "Lançamentos" });
    return within(table).getByRole("columnheader", { name });
  }

  it("forma de pagamento, valor bruto e taxa são de prioridade 3; as demais ficam sempre", async () => {
    await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    for (const name of ["Forma de pagamento", "Valor bruto", "Taxa"]) {
      expect(header(name).className).toContain("hidden");
      expect(header(name).className).toContain("lg:table-cell");
    }
    for (const name of ["Criado em", "Tipo", "Conta", "Valor líquido", "Status"]) {
      expect(header(name).className).not.toContain("hidden");
    }
  });

  it("a conta quebra em qualquer ponto em vez de alargar a tabela, sem cortar o nome", async () => {
    await renderPage();
    await waitFor(() => expect(rows()).toHaveLength(3));
    const conta = await waitFor(() => within(rows()[0]).getByText("Mensalidades"));
    expect(conta.className.split(" ")).toContain("wrap-anywhere");
    const cell = conta.closest("td");
    expect(cell?.className).not.toContain("truncate");
    expect(cell?.className).not.toContain("whitespace-nowrap");
  });
});
