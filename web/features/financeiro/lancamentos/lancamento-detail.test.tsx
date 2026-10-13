import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { WEB_ROOT } from "@/test/design/guard";
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
  mockComprovantes,
  mockContas,
  mockContasPendentes,
  mockLancamentos,
  mockWrite,
  PERMISSOES_LEITURA,
  PERMISSOES_TESOURARIA,
  SESSOES_SEM_LANCAMENTO_READ,
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
  mockComprovantes();
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
    // Valor líquido com o sentido do lançamento (DSC-06, T6): receita com "+".
    expect(detailValue("Valor líquido")).toBe("+R$ 97,50");
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
    mockComprovantes();
    mockLancamentos([CRIADA]);
    renderWithSession(await FinanceiroLancamentoPage({ params: Promise.resolve({ id: ID }) }));
    expect(await screen.findByRole("heading", { name: `Receita ${ID.slice(0, 8)}` })).toBeTruthy();
  });

  async function renderRoute(permissions: string[]) {
    mockMe(authContext({ permissions }));
    const contas = mockContas();
    const comprovantes = mockComprovantes();
    const list = mockLancamentos([CRIADA]);
    renderWithSession(await FinanceiroLancamentoPage({ params: Promise.resolve({ id: ID }) }));
    return { contas, comprovantes, list };
  }

  // FND-04 AC4: a rota depende de lancamento:read, não das leituras vizinhas.
  it.each(SESSOES_SEM_LANCAMENTO_READ)("sem lancamento:read, %s: Sem acesso e nenhuma consulta", async (_, permissions) => {
    const { contas, comprovantes, list } = await renderRoute(permissions);
    expect(await screen.findByRole("heading", { name: "Sem acesso" })).toBeTruthy();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(screen.queryByRole("heading", { name: `Receita ${ID.slice(0, 8)}` })).toBeNull();
    expect(screen.queryByRole("heading", { name: "Comprovantes" })).toBeNull();
    expect(list.count).toBe(0);
    expect(comprovantes.count).toBe(0);
    expect(contas.count).toBe(0);
  });

  it("só com lancamento:read, a página mostra o detalhe, sem ações nem comprovantes", async () => {
    const { comprovantes, list } = await renderRoute(["financeiro:lancamento:read"]);
    expect(await screen.findByRole("heading", { name: `Receita ${ID.slice(0, 8)}` })).toBeTruthy();
    expect(detailValue("Valor bruto")).toBe("R$ 100,00");
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(screen.queryByRole("heading", { name: "Sem acesso" })).toBeNull();
    expect(screen.queryByRole("heading", { name: "Comprovantes" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Editar" })).toBeNull();
    expect(list.count).toBe(1);
    expect(comprovantes.count).toBe(0);
  });
});

// DSM-03 AC1 e AC2, DSP-01 e DSP-07 AC7: trilha, status no cabeçalho, valores
// com MoneyValue e identificador com o código em fonte monoespaçada.
describe("apresentação", () => {
  it("trilha volta à listagem e o status fica ao lado do título", async () => {
    await renderDetail();
    const heading = await screen.findByRole("heading", { level: 1, name: `Receita ${ID.slice(0, 8)}` });
    const nav = screen.getByRole("navigation", { name: "Trilha de navegação" });
    const crumbs = within(nav).getAllByRole("link", { name: "Lançamentos" });
    for (const crumb of crumbs) expect(crumb.getAttribute("href")).toBe("/financeiro/lancamentos");
    const badge = heading.parentElement?.querySelector<HTMLElement>('[data-slot="badge"]');
    expect(badge?.textContent).toBe("Criada");
    expect(badge?.querySelector("svg")).not.toBeNull();
    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
  });

  it("despesa mostra o valor líquido com −; os valores usam MoneyValue", async () => {
    await renderDetail([lancamento({ ...CRIADA, tipo: "DESPESA", conta_id: CONTA_DESPESA.id })]);
    await waitFor(() => expect(detailValue("Valor líquido")).toBe("−R$ 97,50"));
    const values = document.querySelectorAll('dd [data-slot="money-value"]');
    expect(Array.from(values).map((v) => v.textContent)).toEqual(["R$ 100,00", "R$ 2,50", "−R$ 97,50"]);
  });

  it("cancelado não leva sinal: não move o saldo", async () => {
    await renderDetail([lancamento({ ...CRIADA, status: "CANCELADA", motivo_cancelamento: "Duplicado" })]);
    await waitFor(() => expect(detailValue("Valor líquido")).toBe("R$ 97,50"));
  });

  it("a devolução mostra o código da receita em fonte monoespaçada", async () => {
    const receitaId = "00000000-0000-4000-8000-000000000299";
    await renderDetail([lancamento({ ...CRIADA, tipo: "DESPESA", conta_id: CONTA_DESPESA.id, devolucao_de_id: receitaId })]);
    const link = await screen.findByRole("link", { name: receitaId.slice(0, 8) });
    expect(link.className).toContain("font-mono");
  });

  it("id fora da lista oferece a volta à listagem", async () => {
    await renderDetail([CRIADA], PERMISSOES_TESOURARIA, "00000000-0000-4000-8000-000000000999");
    const back = await screen.findByRole("link", { name: "Voltar para lançamentos" });
    expect(back.getAttribute("href")).toBe("/financeiro/lancamentos");
  });
});

// DSM-05 AC2 e AC3: nada de select-field nem de <select> avulso em DS-FB.
describe("limpeza", () => {
  const ROOTS = ["features/financeiro/lancamentos", "features/financeiro/comprovantes", "app/(app)/financeiro/lancamentos"];

  function sources(dir: string): string[] {
    return readdirSync(join(WEB_ROOT, dir), { withFileTypes: true }).flatMap((entry) => {
      const path = `${dir}/${entry.name}`;
      if (entry.isDirectory()) return sources(path);
      return /\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name) ? [path] : [];
    });
  }

  it("select-field.tsx não existe mais", () => {
    expect(existsSync(join(WEB_ROOT, "features/financeiro/lancamentos/select-field.tsx"))).toBe(false);
  });

  it("nenhum <select> fora do NativeSelect do kit", () => {
    const files = ROOTS.flatMap(sources);
    expect(files.length).toBeGreaterThan(5);
    const offenders = files.filter((file) => /<select[\s>]|select-field/.test(readFileSync(join(WEB_ROOT, file), "utf8")));
    expect(offenders).toEqual([]);
  });
});

// Smoke da rodada 6, D3: aberto direto pela URL, o detalhe não mostra o prefixo
// do id no lugar do nome da conta enquanto o plano de contas carrega.
describe("nome da conta no detalhe", () => {
  it("mostra Carregando… até o plano de contas chegar, e depois o nome", async () => {
    mockMe(authContext({ permissions: PERMISSOES_TESOURARIA }));
    const contas = mockContasPendentes();
    mockComprovantes();
    mockLancamentos([CRIADA]);
    renderWithSession(<LancamentoDetail id={ID} />);
    await screen.findByRole("heading", { name: `Receita ${ID.slice(0, 8)}` });
    expect(detailValue("Conta")).toBe("Carregando…");
    expect(detailValue("Conta")).not.toBe(CONTA_RECEITA.id.slice(0, 8));
    contas.release();
    await waitFor(() => expect(detailValue("Conta")).toBe("Mensalidades"));
  });

  it("conta que não está no plano de contas mantém o identificador curto", async () => {
    mockMe(authContext({ permissions: PERMISSOES_TESOURARIA }));
    mockContas([CONTA_DESPESA]);
    mockComprovantes();
    mockLancamentos([CRIADA]);
    renderWithSession(<LancamentoDetail id={ID} />);
    await screen.findByRole("heading", { name: `Receita ${ID.slice(0, 8)}` });
    await waitFor(() => expect(detailValue("Conta")).toBe(CONTA_RECEITA.id.slice(0, 8)));
  });

  it("sem financeiro:conta:read, não fica em Carregando… e mostra o identificador curto", async () => {
    await renderDetail([CRIADA], PERMISSOES_TESOURARIA.filter((p) => p !== "financeiro:conta:read"));
    await screen.findByRole("heading", { name: `Receita ${ID.slice(0, 8)}` });
    expect(detailValue("Conta")).toBe(CONTA_RECEITA.id.slice(0, 8));
  });
});

// Smoke da rodada 6, D2 (equivalente no detalhe): o link da receita devolvida
// tem alvo de toque pelo token `control-sm`.
describe("alvo de toque do link da devolução", () => {
  it("usa a altura do token de controle", async () => {
    const receitaId = "00000000-0000-4000-8000-000000000299";
    await renderDetail([
      lancamento({ ...CRIADA, tipo: "DESPESA", conta_id: CONTA_DESPESA.id, devolucao_de_id: receitaId }),
    ]);
    const link = await screen.findByRole("link", { name: receitaId.slice(0, 8) });
    for (const name of ["inline-flex", "min-h-control-sm", "items-center"]) {
      expect(link.className.split(" ")).toContain(name);
    }
  });
});
