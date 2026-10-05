import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, movimentacao, problem, produto } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { ProdutoDetail } from "../produto-detail";
import { ALL, mockEstoque, mockProdutos, PERM, READ_ONLY } from "../test-helpers";

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
