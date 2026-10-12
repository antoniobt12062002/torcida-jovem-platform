import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import AppError from "./(app)/error";
import GlobalError from "./global-error";
import NotFound from "./not-found";

// As fontes do Next são transformadas na compilação; no teste, só a variável importa.
vi.mock("next/font/google", () => ({
  Geist: () => ({ variable: "font-geist-sans" }),
  Barlow_Condensed: () => ({ variable: "font-barlow-condensed" }),
}));

// Erro de exemplo com tudo o que não pode aparecer na tela (DSS-04 AC4).
const MESSAGE = "falha interna: coluna tj_secret inexistente";
const DIGEST = "digest-4242";
function sampleError() {
  const error = new Error(MESSAGE) as Error & { digest?: string };
  error.digest = DIGEST;
  error.stack = `Error: ${MESSAGE}\n    at renderPage (app/(app)/page.tsx:10:5)`;
  return error;
}

function expectNoTechnicalDetail() {
  const text = document.body.textContent ?? "";
  expect(text).not.toContain(MESSAGE);
  expect(text).not.toContain(DIGEST);
  expect(text).not.toContain("renderPage");
}

let consoleError: ReturnType<typeof vi.spyOn>;
beforeEach(() => {
  document.title = "";
});
afterEach(() => consoleError?.mockRestore());

describe("erro numa rota autenticada (app/(app)/error.tsx)", () => {
  it("mostra o h1 em português com o foco, sem detalhe técnico", () => {
    render(<AppError error={sampleError()} retry={vi.fn()} />);
    const heading = screen.getByRole("heading", { level: 1, name: "Algo deu errado" });
    expect(document.activeElement).toBe(heading);
    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    expect(screen.getByText(/Não foi possível mostrar esta página/)).toBeTruthy();
    expect(document.title).toBe("Algo deu errado | TJ Platform");
    expectNoTechnicalDetail();
  });

  it("Tentar de novo chama a função de nova tentativa recebida; Voltar ao início leva a /inicio", () => {
    const retry = vi.fn();
    render(<AppError error={sampleError()} retry={retry} />);
    fireEvent.click(screen.getByRole("button", { name: "Tentar de novo" }));
    expect(retry).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("link", { name: "Voltar ao início" }).getAttribute("href")).toBe("/inicio");
  });

  it("fica numa única região principal, a do contêiner de página", () => {
    render(<AppError error={sampleError()} retry={vi.fn()} />);
    const mains = screen.getAllByRole("main");
    expect(mains).toHaveLength(1);
    expect(mains[0].id).toBe("conteudo");
  });
});

describe("endereço inexistente (app/not-found.tsx)", () => {
  it("mostra Página não encontrada com o foco, o escudo e o link para o início", () => {
    render(<NotFound />);
    const heading = screen.getByRole("heading", { level: 1, name: "Página não encontrada" });
    expect(document.activeElement).toBe(heading);
    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    expect(screen.getByRole("link", { name: "Ir para o início" }).getAttribute("href")).toBe("/inicio");
    expect(document.querySelector("[data-slot='access-layout'] [data-slot='brand-mark']")).not.toBeNull();
    expect(document.body.textContent).toContain("Torcida Jovem");
  });
});

describe("erro no layout raiz (app/global-error.tsx)", () => {
  it("traz documento próprio em pt-BR, escudo, h1 com foco e nenhum detalhe técnico", () => {
    // `html` e `body` dentro do contêiner do teste geram aviso de aninhamento do React.
    consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
    render(<GlobalError error={sampleError()} retry={vi.fn()} />);
    expect(document.documentElement.getAttribute("lang")).toBe("pt-BR");
    const heading = screen.getByRole("heading", { level: 1, name: "Algo deu errado" });
    expect(document.activeElement).toBe(heading);
    expect(document.querySelector("[data-slot='brand-mark']")).not.toBeNull();
    expect(document.body.textContent).toContain("Torcida Jovem");
    expectNoTechnicalDetail();
  });

  it("Tentar de novo chama a função de nova tentativa recebida", () => {
    consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
    const retry = vi.fn();
    render(<GlobalError error={sampleError()} retry={retry} />);
    fireEvent.click(screen.getByRole("button", { name: "Tentar de novo" }));
    expect(retry).toHaveBeenCalledTimes(1);
  });
});
