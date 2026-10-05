import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ApiError, networkError } from "@/lib/api/problem";

import { EmptyState, ErrorState, LoadingState } from "./states";

// FND-05 AC3: estados padronizados de carregamento, lista vazia e erro com
// opção de tentar de novo.

describe("LoadingState", () => {
  it("anuncia o carregamento", () => {
    render(<LoadingState />);
    expect(screen.getByRole("status").textContent).toContain("Carregando");
  });
});

describe("EmptyState", () => {
  it("mostra o título e a descrição da lista vazia", () => {
    render(<EmptyState title="Nenhuma conta cadastrada." description="Crie a primeira conta." />);
    expect(screen.getByText("Nenhuma conta cadastrada.")).toBeTruthy();
    expect(screen.getByText("Crie a primeira conta.")).toBeTruthy();
  });

  it("sem título, usa a mensagem padrão", () => {
    render(<EmptyState />);
    expect(screen.getByText("Nenhum item encontrado.")).toBeTruthy();
  });
});

describe("ErrorState", () => {
  it("mostra a mensagem do erro da API e tenta de novo ao clicar", () => {
    const onRetry = vi.fn();
    render(<ErrorState error={networkError()} onRetry={onRetry} />);
    expect(screen.getByRole("alert").textContent).toContain("O serviço está indisponível");
    fireEvent.click(screen.getByRole("button", { name: "Tentar de novo" }));
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it("usa o catálogo da tela para o code conhecido (FND-06 AC2)", () => {
    const error = new ApiError({ status: 409, code: "conta_inativa", title: "Conflict" });
    render(<ErrorState error={error} catalog={{ conta_inativa: "A conta está inativa." }} onRetry={() => {}} />);
    expect(screen.getByRole("alert").textContent).toContain("A conta está inativa.");
  });

  it("para um erro que não é da API, mostra mensagem genérica sem detalhes internos", () => {
    render(<ErrorState error={new Error("TypeError: x is undefined")} onRetry={() => {}} />);
    const text = screen.getByRole("alert").textContent ?? "";
    expect(text).toContain("Não foi possível carregar os dados.");
    expect(text).not.toContain("TypeError");
  });
});
