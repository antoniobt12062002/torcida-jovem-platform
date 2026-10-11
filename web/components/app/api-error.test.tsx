import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ApiError, networkError } from "@/lib/api/problem";

import { ApiErrorAlert, apiErrorMessage } from "./api-error";

function apiError(status: number, code: string) {
  return new ApiError({ status, code, title: "Erro", detail: "detalhe interno da API" });
}

describe("apiErrorMessage", () => {
  // FND-04 AC5.
  it("403 forbidden vira a mensagem de permissão", () => {
    expect(apiErrorMessage(apiError(403, "forbidden"))).toBe("Você não tem permissão para esta ação.");
  });

  // FND-06 AC2.
  it("usa a mensagem do catálogo da tela para o code conhecido", () => {
    expect(apiErrorMessage(apiError(409, "saldo_insuficiente"), { saldo_insuficiente: "Saldo insuficiente." })).toBe(
      "Saldo insuficiente.",
    );
  });

  // FND-06 AC3.
  it("code fora do catálogo: mensagem genérica com o status, sem detalhes internos", () => {
    const message = apiErrorMessage(apiError(409, "codigo_novo"));
    expect(message).toContain("409");
    expect(message).not.toContain("codigo_novo");
    expect(message).not.toContain("detalhe interno");
  });

  // FND-06 AC4.
  it("413 informa conteúdo grande demais", () => {
    expect(apiErrorMessage(apiError(413, "payload_too_large"))).toBe("O conteúdo enviado é grande demais.");
  });
});

describe("ApiErrorAlert", () => {
  it("mostra a mensagem num alerta", () => {
    render(<ApiErrorAlert error={apiError(403, "forbidden")} />);
    expect(screen.getByRole("alert").textContent).toContain("Você não tem permissão para esta ação.");
  });

  // FND-06 AC5.
  it("com a API indisponível, informa e oferece tentar de novo", () => {
    const onRetry = vi.fn();
    render(<ApiErrorAlert error={networkError()} onRetry={onRetry} />);
    expect(screen.getByRole("alert").textContent).toContain("O serviço está indisponível");
    fireEvent.click(screen.getByRole("button", { name: "Tentar de novo" }));
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it("um erro que repetir não resolve não oferece tentar de novo", () => {
    render(<ApiErrorAlert error={apiError(403, "forbidden")} onRetry={() => {}} />);
    expect(screen.queryByRole("button", { name: "Tentar de novo" })).toBeNull();
  });
});
