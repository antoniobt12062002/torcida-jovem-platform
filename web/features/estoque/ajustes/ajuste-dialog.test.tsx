import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, movimentacao, problem, produto } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { ProdutoDetail } from "../produto-detail";
import { ProdutosPage } from "../produtos/produtos-page";
import { ALL, allExcept, mockEstoque, mockProdutos, PERM, READ_ONLY } from "../test-helpers";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const camisa = produto({ unidade_medida: "UN" });
const WARNING = "O saldo ficará negativo.";

function renderDetail(permissions: string[], saldo = 0) {
  mockMe(authContext({ permissions }));
  mockProdutos([camisa]);
  const backend = mockEstoque(saldo === 0 ? [] : [movimentacao({ tipo: "ENTRADA", quantidade: saldo })]);
  renderWithSession(<ProdutoDetail id={camisa.id} />);
  return backend;
}

async function openAjuste() {
  fireEvent.click(await screen.findByRole("button", { name: "Ajustar estoque" }));
  return screen.findByRole("dialog", { name: "Ajustar estoque" });
}

function fill(dialog: HTMLElement, quantidade: string, motivo: string) {
  const q = within(dialog).getByLabelText("Quantidade");
  fireEvent.change(q, { target: { value: quantidade } });
  fireEvent.blur(q);
  fireEvent.change(within(dialog).getByLabelText("Motivo"), { target: { value: motivo } });
}

function confirmButton(dialog: HTMLElement) {
  return within(dialog).getByRole("button", { name: "Confirmar ajuste" }) as HTMLButtonElement;
}

function previewText(dialog: HTMLElement, testId: string) {
  return within(dialog).getByTestId(testId).textContent;
}

// EWB-03 AC1, AC3 e o teste independente da spec.
describe("ajuste com sucesso", () => {
  it("−3 num produto com saldo 0: avisa o negativo, envia e atualiza saldo e histórico", async () => {
    const backend = renderDetail(ALL, 0);
    await waitFor(() => expect(screen.getByTestId("saldo-valor").textContent).toBe("0 UN"));
    const dialog = await openAjuste();
    fill(dialog, "-3", "  Contagem física do inventário  ");

    expect(previewText(dialog, "ajuste-saldo-atual")).toBe("0 UN");
    expect(previewText(dialog, "ajuste-saldo-resultante")).toBe("-3 UN");
    expect(within(dialog).getByText(WARNING, { exact: false })).toBeTruthy();
    expect(confirmButton(dialog).disabled).toBe(false);

    fireEvent.click(confirmButton(dialog));
    await waitFor(() => expect(screen.getByTestId("saldo-valor").textContent).toBe("-3 UN"));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Ajustar estoque" })).toBeNull());

    expect(backend.ajusteRequests).toEqual([
      {
        body: { produto_id: camisa.id, quantidade: -3, motivo: "Contagem física do inventário" },
        csrf: "csrf-token-de-teste",
      },
    ]);
    // Histórico rebuscado: o ajuste aparece com o motivo.
    const table = await screen.findByRole("table", { name: "Movimentações" });
    const row = within(table).getAllByRole("row")[1];
    const cells = within(row).getAllByRole("cell").map((c) => c.textContent);
    expect(cells.slice(0, 4)).toEqual(["Ajuste", "-3", "Ajuste manual", "Contagem física do inventário"]);
  });

  it("resultado positivo não mostra aviso", async () => {
    renderDetail(ALL, 5);
    await waitFor(() => expect(screen.getByTestId("saldo-valor").textContent).toBe("5 UN"));
    const dialog = await openAjuste();
    fill(dialog, "+2", "Achado no depósito");
    expect(previewText(dialog, "ajuste-saldo-atual")).toBe("5 UN");
    expect(previewText(dialog, "ajuste-saldo-resultante")).toBe("7 UN");
    expect(within(dialog).queryByText(WARNING, { exact: false })).toBeNull();
    expect(confirmButton(dialog).disabled).toBe(false);
  });
});

// EWB-03 AC2.
describe("confirmação desabilitada", () => {
  it.each([
    ["sem quantidade nem motivo", "", ""],
    ["quantidade zero", "0", "Correção"],
    ["motivo vazio", "-3", ""],
    ["motivo só com espaços", "-3", "   \n\t "],
    ["quantidade decimal", "1,5", "Correção"],
  ])("%s: desabilitada e a API não é chamada", async (_name, quantidade, motivo) => {
    const backend = renderDetail(ALL, 0);
    const dialog = await openAjuste();
    fill(dialog, quantidade, motivo);
    expect(confirmButton(dialog).disabled).toBe(true);
    fireEvent.click(confirmButton(dialog));
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(backend.ajusteRequests).toEqual([]);
    expect(screen.getByRole("dialog", { name: "Ajustar estoque" })).toBeTruthy();
  });
});

// EWB-03 AC4.
describe("erros da API", () => {
  it.each([
    ["motivo_obrigatorio", "Informe o motivo do ajuste."],
    ["quantidade_invalida", "Quantidade inválida."],
  ])("422 %s mostra a mensagem e mantém o diálogo", async (code, message) => {
    renderDetail(ALL, 0);
    await screen.findByTestId("saldo-valor");
    server.use(http.post(apiUrl("/api/v1/estoque/ajustes"), () => problem(422, code)));
    const dialog = await openAjuste();
    fill(dialog, "-1", "Motivo");
    fireEvent.click(confirmButton(dialog));
    expect(await within(dialog).findByText(message)).toBeTruthy();
    expect(screen.getByTestId("saldo-valor").textContent).toBe("0 UN");
  });

  // N2 (smoke T13): os controles ficam desabilitados durante o envio; o erro
  // recebe o foco, dentro do diálogo.
  it("o erro da API recebe o foco, dentro do diálogo", async () => {
    renderDetail(ALL, 0);
    await screen.findByTestId("saldo-valor");
    server.use(http.post(apiUrl("/api/v1/estoque/ajustes"), () => problem(422, "quantidade_invalida")));
    const dialog = await openAjuste();
    fill(dialog, "-1", "Motivo");
    confirmButton(dialog).focus();
    fireEvent.click(confirmButton(dialog));
    const message = await within(dialog).findByText("Quantidade inválida.");
    const alert = message.closest('[role="alert"]');
    expect(alert).not.toBeNull();
    await waitFor(() => expect(document.activeElement).toBe(alert));
  });
});

// EWB-04 AC1-2.
describe("permissões", () => {
  it("somente leitura não vê o ajuste", async () => {
    renderDetail(READ_ONLY, 0);
    await screen.findByTestId("saldo-valor");
    expect(screen.queryByRole("button", { name: "Ajustar estoque" })).toBeNull();
  });

  it("com ajuste mas sem estoque:saldo:read não há prévia possível: nem ajuste nem chamada de saldo", async () => {
    const backend = renderDetail([PERM.produtoRead, PERM.movAdjust], 0);
    await screen.findByRole("heading", { name: "Camisa oficial" });
    expect(screen.queryByRole("button", { name: "Ajustar estoque" })).toBeNull();
    expect(backend.saldoCalls).toBe(0);
  });
});

// EWB-03 AC1, FND-04 AC3 (lacuna G5): o ajuste depende só de
// estoque:movimentacao:adjust.
describe("isolamento de estoque:movimentacao:adjust", () => {
  it("leitura do saldo + só estoque:movimentacao:adjust: ajusta e envia o POST /ajustes", async () => {
    const backend = renderDetail([PERM.produtoRead, PERM.saldoRead, PERM.movAdjust], 4);
    const dialog = await openAjuste();
    fill(dialog, "-1", "Peça danificada");
    fireEvent.click(confirmButton(dialog));
    await waitFor(() => expect(screen.getByTestId("saldo-valor").textContent).toBe("3 UN"));
    expect(backend.ajusteRequests).toEqual([
      {
        body: { produto_id: camisa.id, quantidade: -1, motivo: "Peça danificada" },
        csrf: "csrf-token-de-teste",
      },
    ]);
    expect(backend.movRequests).toEqual([]);
  });

  it("todas as permissões menos estoque:movimentacao:adjust: sem Ajustar estoque", async () => {
    renderDetail(allExcept(PERM.movAdjust), 4);
    await screen.findByTestId("saldo-valor");
    expect(await screen.findByRole("button", { name: "Registrar entrada" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Ajustar estoque" })).toBeNull();
  });

  it.each([
    ["estoque:movimentacao:create", PERM.movCreate],
    ["estoque:produto:create", PERM.produtoCreate],
  ])("leituras + só %s: sem Ajustar estoque", async (_name, permission) => {
    renderDetail([...READ_ONLY, permission], 4);
    await screen.findByTestId("saldo-valor");
    expect(screen.queryByRole("button", { name: "Ajustar estoque" })).toBeNull();
  });
});

// DSM-04 AC6, DSP-03 AC6: o aviso de saldo negativo é um alerta de aviso, com
// ícone e texto, e não bloqueia a confirmação.
describe("aviso de saldo negativo", () => {
  it("é um alerta no tom de aviso, com ícone e texto, sem desabilitar a confirmação", async () => {
    renderDetail(ALL, 1);
    await waitFor(() => expect(screen.getByTestId("saldo-valor").textContent).toBe("1 UN"));
    const dialog = await openAjuste();
    fill(dialog, "-3", "Contagem");
    const aviso = within(dialog).getByRole("alert");
    expect(aviso.textContent).toContain(WARNING);
    expect(aviso.getAttribute("data-testid")).toBe("ajuste-aviso-negativo");
    expect(aviso.className).toContain("bg-warning-subtle");
    expect(aviso.querySelector("svg")).toBeTruthy();
    expect(within(dialog).getByTestId("ajuste-saldo-resultante").className).toContain("tabular-nums");
    expect(within(dialog).getByTestId("ajuste-saldo-atual").className).toContain("tabular-nums");
    expect(confirmButton(dialog).disabled).toBe(false);
  });

  it("as ações ficam no rodapé do diálogo, que ocupa a tela abaixo de 640 px", async () => {
    renderDetail(ALL, 0);
    const dialog = await openAjuste();
    const footer = dialog.querySelector('[data-slot="dialog-footer"]');
    expect(footer?.contains(confirmButton(dialog))).toBe(true);
    expect(footer?.textContent).toContain("Cancelar");
  });
});

// EWB-04 AC1, T10: somente leitura (CONSELHO_FISCAL) sem ação de escrita em
// nenhuma tela do estoque.
describe("somente leitura em todas as telas do estoque", () => {
  const ESCRITAS = [/^Novo produto$/, /^Registrar entrada$/, /^Registrar saída$/, /^Devolver/, /^Ajustar estoque$/];

  it("na lista de produtos, nenhuma ação de escrita", async () => {
    mockMe(authContext({ permissions: READ_ONLY }));
    mockProdutos([camisa]);
    renderWithSession(<ProdutosPage />);
    await screen.findByText(camisa.nome);
    for (const name of ESCRITAS) expect(screen.queryByRole("button", { name })).toBeNull();
  });

  it("no detalhe, com saldo e histórico, nenhuma ação de escrita", async () => {
    renderDetail(READ_ONLY, 5);
    await screen.findByTestId("saldo-valor");
    await screen.findByRole("table", { name: "Movimentações" });
    for (const name of ESCRITAS) expect(screen.queryByRole("button", { name })).toBeNull();
    expect(screen.queryAllByRole("button").map((b) => b.textContent)).toEqual([]);
  });
});
