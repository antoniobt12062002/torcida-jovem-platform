import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { LancamentoDetail } from "@/features/financeiro/lancamentos/lancamento-detail";
import {
  mockComprovantes,
  mockContas,
  mockLancamentos,
  PERMISSOES_LEITURA,
  PERMISSOES_TESOURARIA,
  tesourariaSem,
} from "@/features/financeiro/lancamentos/test-helpers";
import { resetNavigation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, type Comprovante, comprovante, ids, lancamento, problem } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { ComprovantesPanel } from "./comprovantes-panel";
import { MAX_UPLOAD_BYTES } from "./limits";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());
afterEach(() => vi.restoreAllMocks());

const LANCAMENTO_ID = ids.lancamento;
const UPLOAD_PATH = `/api/v1/financeiro/lancamentos/${LANCAMENTO_ID}/comprovantes`;

const ANTIGO = comprovante({
  id: "00000000-0000-4000-8000-000000000311",
  original_filename: "recibo-agosto.pdf",
  content_type: "application/pdf",
  size_bytes: 2048,
  version: 1,
  uploaded_by: ids.user,
  uploaded_at: "2026-09-01T12:00:00Z",
});
const NOVO = comprovante({
  id: "00000000-0000-4000-8000-000000000312",
  original_filename: "foto-nota.png",
  content_type: "image/png",
  size_bytes: 512,
  version: 2,
  uploaded_by: ids.otherUser,
  uploaded_at: "2026-10-04T15:30:00Z",
});

async function renderPanel(items: Comprovante[] | (() => Comprovante[]) = [ANTIGO], permissions = PERMISSOES_TESOURARIA) {
  mockMe(authContext({ permissions }));
  const list = mockComprovantes(items);
  renderWithSession(<ComprovantesPanel lancamentoId={LANCAMENTO_ID} />);
  await screen.findByRole("heading", { name: "Comprovantes" });
  // A ação de anexar aparece quando a sessão termina de carregar.
  if (permissions.includes("financeiro:comprovante:create")) await screen.findByLabelText("Anexar comprovante");
  return list;
}

function rows() {
  return screen.queryAllByTestId("comprovante-row");
}

type Upload = { contentType: string | null; csrf: string | null; field: boolean; size: number; content: string };

/** Registra os envios de comprovante (multipart) e responde com `respond`. */
function mockUpload(respond: () => Response | Promise<Response>) {
  const uploads: Upload[] = [];
  server.use(
    http.post(apiUrl(UPLOAD_PATH), async ({ request }) => {
      const contentType = request.headers.get("Content-Type");
      const csrf = request.headers.get("X-CSRF-Token");
      const form = await request.formData();
      const file = form.get("file") as Blob | null;
      uploads.push({
        contentType,
        csrf,
        field: file !== null,
        size: file?.size ?? -1,
        content: file && file.size < 1024 ? await file.text() : "",
      });
      return respond();
    }),
  );
  return uploads;
}

function fileInput() {
  return screen.getByLabelText("Anexar comprovante") as HTMLInputElement;
}

function choose(file: File) {
  fireEvent.change(fileInput(), { target: { files: [file] } });
}

function fileOfSize(size: number, name = "comprovante.pdf") {
  return new File([new Uint8Array(size)], name, { type: "application/pdf" });
}

const LIMIT_MESSAGE = /excede o limite permitido/;

// FWB-05 AC1.
describe("lista", () => {
  it("lista nome, tipo, tamanho, versão, autor e data, do mais novo para o mais antigo", async () => {
    await renderPanel([ANTIGO, NOVO]);
    await waitFor(() => expect(rows()).toHaveLength(2));
    const texts = rows().map((r) => within(r).getAllByRole("cell").map((c) => c.textContent));
    expect(texts).toEqual([
      ["foto-nota.png", "PNG", "512 bytes", "2", ids.otherUser.slice(0, 8), "04/10/2026 12:30", "Baixar"],
      ["recibo-agosto.pdf", "PDF", "2,0 KB", "1", "você", "01/09/2026 09:00", "Baixar"],
    ]);
  });

  it("sem comprovantes, mostra o estado vazio", async () => {
    await renderPanel([]);
    expect(await screen.findByText("Nenhum comprovante anexado.")).toBeTruthy();
  });
});

// FWB-05 AC2.
describe("anexar", () => {
  it("envia multipart/form-data no campo file com CSRF, mostra Enviando… e atualiza a lista", async () => {
    const state = { items: [ANTIGO] };
    const list = await renderPanel(() => state.items);
    await waitFor(() => expect(rows()).toHaveLength(1));
    let release: () => void = () => {};
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    const uploads = mockUpload(async () => {
      await gate;
      state.items = [ANTIGO, NOVO];
      return HttpResponse.json(NOVO, { status: 201 });
    });

    choose(new File(["%PDF-conteudo"], "nota.pdf", { type: "application/pdf" }));
    expect(await screen.findByText("Enviando…")).toBeTruthy();
    release();

    await waitFor(() => expect(rows()).toHaveLength(2));
    expect(screen.queryByText("Enviando…")).toBeNull();
    expect(uploads).toHaveLength(1);
    expect(uploads[0].contentType).toMatch(/^multipart\/form-data; boundary=/);
    expect(uploads[0].csrf).toBe("csrf-token-de-teste");
    expect(uploads[0].field).toBe(true);
    expect(uploads[0].content).toBe("%PDF-conteudo");
    expect(list.count).toBe(2);
    expect(await screen.findByText("Comprovante anexado.")).toBeTruthy();
  });

  it("o seletor sugere PDF, JPG, JPEG, PNG e WEBP", async () => {
    await renderPanel();
    expect(fileInput().getAttribute("accept")).toBe(".pdf,.jpg,.jpeg,.png,.webp");
  });
});

// FWB-05 AC3: limite efetivo do cliente.
describe("limite de tamanho", () => {
  it("o limite é 10 MiB − 64 KiB (10.420.224 bytes)", () => {
    expect(MAX_UPLOAD_BYTES).toBe(10 * 1024 * 1024 - 64 * 1024);
    expect(MAX_UPLOAD_BYTES).toBe(10_420_224);
  });

  it("abaixo do limite, envia", async () => {
    await renderPanel();
    const uploads = mockUpload(() => HttpResponse.json(NOVO, { status: 201 }));
    choose(fileOfSize(10_420_223));
    await waitFor(() => expect(uploads).toHaveLength(1));
    expect(uploads[0].size).toBe(10_420_223);
    expect(screen.queryByText(LIMIT_MESSAGE)).toBeNull();
  });

  it("exatamente no limite (10.420.224 bytes), envia", async () => {
    await renderPanel();
    const uploads = mockUpload(() => HttpResponse.json(NOVO, { status: 201 }));
    choose(fileOfSize(10_420_224));
    await waitFor(() => expect(uploads).toHaveLength(1));
    expect(uploads[0].size).toBe(10_420_224);
    expect(screen.queryByText(LIMIT_MESSAGE)).toBeNull();
  });

  it("um byte acima (10.420.225 bytes), informa o limite e não faz nenhuma requisição", async () => {
    const list = await renderPanel();
    await waitFor(() => expect(list.count).toBe(1));
    const requests: string[] = [];
    const onRequest = ({ request }: { request: Request }) => requests.push(`${request.method} ${request.url}`);
    server.events.on("request:start", onRequest);
    try {
      choose(fileOfSize(10_420_225));
      expect(
        await screen.findByText("O arquivo excede o limite permitido de 9,9 MB. Escolha um arquivo menor."),
      ).toBeTruthy();
      await new Promise((resolve) => setTimeout(resolve, 50));
      expect(requests).toEqual([]);
    } finally {
      server.events.removeListener("request:start", onRequest);
    }
  });
});

// FWB-05 AC4 e AC6 (envio).
describe("erros do envio", () => {
  it.each([
    [413, "document_too_large", "O arquivo excede o limite permitido. Escolha um arquivo menor."],
    [413, "payload_too_large", "O arquivo excede o limite permitido. Escolha um arquivo menor."],
    [422, "document_extension_not_allowed", "Extensão de arquivo não permitida. Use PDF, JPG, JPEG, PNG ou WEBP."],
    [422, "document_type_not_allowed", "Tipo de arquivo não permitido. Use PDF, JPG, JPEG, PNG ou WEBP."],
    [422, "document_type_mismatch", "O conteúdo do arquivo não corresponde à extensão do nome."],
    [404, "lancamento_nao_encontrado", "Lançamento não encontrado."],
  ])("%i %s", async (status, code, message) => {
    const list = await renderPanel();
    await waitFor(() => expect(rows()).toHaveLength(1));
    const uploads = mockUpload(() => problem(status, code));
    choose(new File(["conteudo"], "nota.pdf", { type: "application/pdf" }));
    expect(await screen.findByText(message)).toBeTruthy();
    expect(uploads).toHaveLength(1);
    expect(rows()).toHaveLength(1);
    expect(list.count).toBe(1);
  });
});

// FWB-05 AC5 e AC6 (download).
describe("baixar", () => {
  function mockUrl(respond: () => Response) {
    const calls = { count: 0 };
    server.use(
      http.get(apiUrl(`/api/v1/financeiro/comprovantes/${ANTIGO.id}/url`), () => {
        calls.count += 1;
        return respond();
      }),
    );
    return calls;
  }

  it("busca a URL assinada no clique, a cada clique, e a abre", async () => {
    const open = vi.spyOn(window, "open").mockImplementation(() => null);
    let n = 0;
    const calls = mockUrl(() => {
      n += 1;
      return HttpResponse.json({ url: `https://s3.exemplo.test/assinada-${n}` });
    });
    await renderPanel();
    const button = await screen.findByRole("button", { name: "Baixar recibo-agosto.pdf" });
    expect(calls.count).toBe(0);

    fireEvent.click(button);
    await waitFor(() => expect(open).toHaveBeenCalledTimes(1));
    expect(open.mock.calls[0][0]).toBe("https://s3.exemplo.test/assinada-1");

    fireEvent.click(button);
    await waitFor(() => expect(open).toHaveBeenCalledTimes(2));
    expect(open.mock.calls[1][0]).toBe("https://s3.exemplo.test/assinada-2");
    expect(calls.count).toBe(2);
  });

  it("não guarda a URL assinada no cache de consultas nem no de mutações", async () => {
    const open = vi.spyOn(window, "open").mockImplementation(() => null);
    mockUrl(() => HttpResponse.json({ url: "https://s3.exemplo.test/assinada-secreta" }));
    mockMe(authContext({ permissions: PERMISSOES_TESOURARIA }));
    mockComprovantes([ANTIGO]);
    const { queryClient } = renderWithSession(<ComprovantesPanel lancamentoId={LANCAMENTO_ID} />);

    fireEvent.click(await screen.findByRole("button", { name: "Baixar recibo-agosto.pdf" }));
    await waitFor(() => expect(open).toHaveBeenCalledTimes(1));

    const cached = JSON.stringify([
      queryClient.getQueryCache().getAll().map((q) => [q.queryKey, q.state.data]),
      queryClient.getMutationCache().getAll().map((m) => [m.state.data, m.state.variables]),
    ]);
    expect(cached).not.toContain("assinada-secreta");
  });

  it("404 document_not_found mostra a mensagem", async () => {
    const open = vi.spyOn(window, "open").mockImplementation(() => null);
    mockUrl(() => problem(404, "document_not_found"));
    await renderPanel();
    fireEvent.click(await screen.findByRole("button", { name: "Baixar recibo-agosto.pdf" }));
    expect(await screen.findByText("Comprovante não encontrado.")).toBeTruthy();
    expect(open).not.toHaveBeenCalled();
  });
});

// Casos de borda: permissões no detalhe do lançamento.
describe("permissões", () => {
  async function renderDetail(permissions: string[]) {
    mockMe(authContext({ permissions }));
    mockContas();
    mockLancamentos([lancamento()]);
    const list = mockComprovantes([ANTIGO]);
    renderWithSession(<LancamentoDetail id={LANCAMENTO_ID} />);
    await screen.findByText("Valor bruto", { selector: "dt" });
    return list;
  }

  it("tesouraria vê a lista e pode anexar", async () => {
    await renderDetail(PERMISSOES_TESOURARIA);
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(screen.getByLabelText("Anexar comprovante")).toBeTruthy();
  });

  it("só leitura vê e baixa, mas não anexa", async () => {
    await renderDetail(PERMISSOES_LEITURA);
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(screen.getByRole("button", { name: "Baixar recibo-agosto.pdf" })).toBeTruthy();
    expect(screen.queryByLabelText("Anexar comprovante")).toBeNull();
  });

  it("sem financeiro:comprovante:read não mostra o painel nem consulta comprovantes", async () => {
    const list = await renderDetail(["financeiro:lancamento:read", "financeiro:conta:read"]);
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(screen.queryByRole("heading", { name: "Comprovantes" })).toBeNull();
    expect(list.count).toBe(0);
  });

  // FWB-05 AC2 e FND-04 AC3: anexar depende só de comprovante:create.
  it("comprovante:read + só comprovante:create pode anexar", async () => {
    await renderDetail([
      "financeiro:lancamento:read",
      "financeiro:comprovante:read",
      "financeiro:comprovante:create",
    ]);
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(await screen.findByLabelText("Anexar comprovante")).toBeTruthy();
  });

  it("todas as escritas de lançamento, sem comprovante:create, não anexa", async () => {
    await renderDetail(tesourariaSem("financeiro:comprovante:create"));
    // A lista do painel e as ações do lançamento aparecem: a sessão já carregou.
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(await screen.findByRole("button", { name: "Cancelar lançamento" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Baixar recibo-agosto.pdf" })).toBeTruthy();
    expect(screen.queryByLabelText("Anexar comprovante")).toBeNull();
  });

  it("no painel isolado, só comprovante:create mostra o campo de anexar", async () => {
    await renderPanel([ANTIGO], ["financeiro:comprovante:read", "financeiro:comprovante:create"]);
    expect(fileInput().type).toBe("file");
  });

  it("no painel isolado, sem comprovante:create não há campo de anexar", async () => {
    await renderPanel([ANTIGO], tesourariaSem("financeiro:comprovante:create"));
    // "você" depende do usuário da sessão: a sessão já carregou.
    await waitFor(() => expect(within(rows()[0]).getByText("você")).toBeTruthy());
    expect(screen.queryByLabelText("Anexar comprovante")).toBeNull();
  });

  // FWB-05 AC1 e FND-04 AC3: o painel depende só de comprovante:read.
  it("lancamento:read + só comprovante:read vê o painel e a lista", async () => {
    const list = await renderDetail(["financeiro:lancamento:read", "financeiro:comprovante:read"]);
    expect(await screen.findByRole("heading", { name: "Comprovantes" })).toBeTruthy();
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(list.count).toBe(1);
  });

  it("todas as outras permissões, sem comprovante:read, não vê o painel nem consulta comprovantes", async () => {
    const list = await renderDetail(tesourariaSem("financeiro:comprovante:read"));
    // As ações do lançamento aparecem: a sessão já carregou.
    expect(await screen.findByRole("button", { name: "Cancelar lançamento" })).toBeTruthy();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(screen.queryByRole("heading", { name: "Comprovantes" })).toBeNull();
    expect(screen.queryByLabelText("Anexar comprovante")).toBeNull();
    expect(list.count).toBe(0);
  });
});
