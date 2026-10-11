import { describe, expect, it } from "vitest";

import { problemBody } from "@/test/msw/fixtures";

import {
  ApiError,
  isUnavailable,
  messageFor,
  MESSAGE_PAYLOAD_TOO_LARGE,
  MESSAGE_UNAVAILABLE,
  networkError,
  toApiError,
} from "./problem";

function problemResponse(status: number) {
  return new Response(null, {
    status,
    headers: { "Content-Type": "application/problem+json" },
  });
}

// FND-01 AC4 e FND-06.
describe("toApiError", () => {
  it("converte problem+json em ApiError com status, code, title, detail e errors[]", () => {
    const body = problemBody(422, "validation_failed", {
      title: "Unprocessable Entity",
      detail: "campos inválidos",
      errors: [
        { field: "nome", code: "required" },
        { field: "valor_bruto_cents", code: "too_small" },
      ],
    });

    const error = toApiError(problemResponse(422), body);

    expect(error).toBeInstanceOf(ApiError);
    expect(error.status).toBe(422);
    expect(error.code).toBe("validation_failed");
    expect(error.title).toBe("Unprocessable Entity");
    expect(error.detail).toBe("campos inválidos");
    expect(error.errors).toEqual([
      { field: "nome", code: "required" },
      { field: "valor_bruto_cents", code: "too_small" },
    ]);
  });

  it("entrega errors[] vazio quando o problem+json não traz erros por campo", () => {
    const error = toApiError(problemResponse(409), problemBody(409, "email_taken"));
    expect(error.code).toBe("email_taken");
    expect(error.errors).toEqual([]);
  });

  it("aceita resposta de erro sem problem+json (por exemplo 413 de um proxy)", () => {
    const error = toApiError(new Response("too large", { status: 413 }), "too large");
    expect(error).toBeInstanceOf(ApiError);
    expect(error.status).toBe(413);
    expect(typeof error.code).toBe("string");
    expect(error.errors).toEqual([]);
  });
});

describe("messageFor", () => {
  it("usa a mensagem do catálogo da tela para um code conhecido (FND-06 AC2)", () => {
    const error = toApiError(problemResponse(409), problemBody(409, "email_taken"));
    expect(messageFor(error, { email_taken: "Este e-mail já está cadastrado." })).toBe(
      "Este e-mail já está cadastrado.",
    );
  });

  it("mostra mensagem genérica com o status HTTP para code fora do catálogo, sem detalhes internos (FND-06 AC3)", () => {
    const error = toApiError(
      problemResponse(409),
      problemBody(409, "conflito_interno", {
        title: "Conflict",
        detail: "violates unique constraint users_email_key",
      }),
    );

    const message = messageFor(error, { email_taken: "Este e-mail já está cadastrado." });

    expect(message).toContain("409");
    expect(message).not.toContain("conflito_interno");
    expect(message).not.toContain("Conflict");
    expect(message).not.toContain("unique constraint");
  });

  it("informa que o conteúdo é grande demais em 413 (FND-06 AC4)", () => {
    const error = toApiError(problemResponse(413), problemBody(413, "payload_too_large"));
    expect(messageFor(error)).toBe(MESSAGE_PAYLOAD_TOO_LARGE);
    expect(MESSAGE_PAYLOAD_TOO_LARGE).toMatch(/grande demais/);
  });

  it("informa serviço indisponível em 503 (FND-06 AC5)", () => {
    const error = toApiError(problemResponse(503), problemBody(503, "service_unavailable"));
    expect(messageFor(error)).toBe(MESSAGE_UNAVAILABLE);
    expect(MESSAGE_UNAVAILABLE).toMatch(/indisponível/);
    expect(isUnavailable(error)).toBe(true);
  });

  it("informa serviço indisponível em falha de rede (FND-06 AC5)", () => {
    const error = networkError(new TypeError("fetch failed"));
    expect(error).toBeInstanceOf(ApiError);
    expect(messageFor(error)).toBe(MESSAGE_UNAVAILABLE);
    expect(isUnavailable(error)).toBe(true);
  });

  it("não trata erro de regra de negócio como indisponibilidade", () => {
    const error = toApiError(problemResponse(409), problemBody(409, "email_taken"));
    expect(isUnavailable(error)).toBe(false);
  });
});
