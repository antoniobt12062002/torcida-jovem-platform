import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ApiError, networkError } from "@/lib/api/problem";

import { Button } from "@/components/ui/button";

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

// DSP-02 AC1: esqueleto no formato aproximado do conteúdo, anunciado.
describe("LoadingState com shape", () => {
  function skeletons(container: HTMLElement) {
    return [...container.querySelectorAll("[data-slot=skeleton]")];
  }

  it.each(["table", "list", "detail", "value"] as const)("%s anuncia o carregamento e esconde o esqueleto", (shape) => {
    const { container } = render(<LoadingState shape={shape} label="Carregando contas…" />);
    const status = screen.getByRole("status");
    expect(status.textContent).toBe("Carregando contas…");
    expect(skeletons(container).length).toBeGreaterThan(0);
    for (const skeleton of skeletons(container)) expect(skeleton.getAttribute("aria-hidden")).toBe("true");
  });

  it("table desenha linhas com a altura de linha de tabela", () => {
    const { container } = render(<LoadingState shape="table" />);
    const rows = skeletons(container).filter((skeleton) => skeleton.classList.contains("h-table-row"));
    expect(rows.length).toBeGreaterThanOrEqual(3);
  });

  it("value desenha um único bloco de valor", () => {
    const { container } = render(<LoadingState shape="value" />);
    expect(skeletons(container)).toHaveLength(1);
    expect(skeletons(container)[0].classList.contains("h-10")).toBe(true);
  });

  it("os quatro formatos são diferentes entre si", () => {
    const shapes = (["table", "list", "detail", "value"] as const).map((shape) => {
      const { container, unmount } = render(<LoadingState shape={shape} />);
      const drawn = skeletons(container)
        .map((skeleton) => skeleton.className)
        .join("|");
      unmount();
      return drawn;
    });
    expect(new Set(shapes).size).toBe(4);
  });
});

// DSP-02 AC2 e AC3: vazio e sem resultado são estados diferentes, com
// orientação e ação.
describe("EmptyState por tipo", () => {
  it("vazio: título como h2, frase de orientação e a ação de criar", () => {
    const { container } = render(
      <EmptyState
        title="Nenhum lançamento ainda."
        description="Crie o primeiro para começar o controle."
        action={<Button>Novo lançamento</Button>}
      />,
    );
    expect(screen.getByRole("heading", { level: 2, name: "Nenhum lançamento ainda." })).toBeTruthy();
    expect(screen.getByText("Crie o primeiro para começar o controle.")).toBeTruthy();
    expect(screen.getByRole("button", { name: "Novo lançamento" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Limpar filtros" })).toBeNull();
    expect(container.querySelector("[data-kind]")?.getAttribute("data-kind")).toBe("empty");
  });

  it("sem resultado: título próprio e a ação Limpar filtros", () => {
    const onClear = vi.fn();
    const { container } = render(<EmptyState kind="no-results" onClear={onClear} />);
    expect(screen.getByRole("heading", { level: 2, name: "Nenhum resultado para os filtros" })).toBeTruthy();
    expect(screen.queryByText("Nenhum item encontrado.")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Limpar filtros" }));
    expect(onClear).toHaveBeenCalledTimes(1);
    expect(container.querySelector("[data-kind]")?.getAttribute("data-kind")).toBe("no-results");
  });

  it("os dois tipos têm ícones diferentes, escondidos de leitores de tela", () => {
    const empty = render(<EmptyState />);
    const emptyIcon = empty.container.querySelector("svg");
    const emptyDrawing = emptyIcon?.innerHTML;
    expect(emptyIcon?.getAttribute("aria-hidden")).toBe("true");
    empty.unmount();
    const none = render(<EmptyState kind="no-results" />);
    const noneIcon = none.container.querySelector("svg");
    expect(noneIcon?.getAttribute("aria-hidden")).toBe("true");
    expect(emptyDrawing).toBeTruthy();
    expect(noneIcon?.innerHTML).not.toBe(emptyDrawing);
  });

  it("dentro de uma seção, o título pode descer para h3 (DSP-01 AC5)", () => {
    render(<EmptyState title="Nenhum comprovante anexado." headingLevel={3} />);
    expect(screen.getByRole("heading", { level: 3, name: "Nenhum comprovante anexado." })).toBeTruthy();
  });
});

// DSP-02 AC4: mensagem de erro com ícone e "Tentar de novo".
describe("ErrorState com ícone", () => {
  it("o alerta de erro leva ícone além do texto", () => {
    render(<ErrorState error={networkError()} onRetry={() => {}} />);
    const alert = screen.getByRole("alert");
    expect(alert.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true");
    expect(within(alert.parentElement as HTMLElement).getByRole("button", { name: "Tentar de novo" })).toBeTruthy();
  });
});
