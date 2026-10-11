import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/problem";

import { applyProblemToForm } from "./apply-problem";

type Form = { nome: string; valor: string };

function validationError(errors: { field: string; code: string }[]) {
  return new ApiError({
    status: 422,
    code: "validation_failed",
    title: "Unprocessable Entity",
    errors,
  });
}

// FND-06 AC1 e design 01-fundacao (lib/forms): cada errors[] vai para o campo
// do formulário; o que não tem campo vira erro do formulário (root).
describe("applyProblemToForm", () => {
  it("põe cada erro de errors[] no campo correspondente", () => {
    const setError = vi.fn();

    applyProblemToForm<Form>(
      validationError([
        { field: "nome", code: "required" },
        { field: "valor_bruto_cents", code: "too_small" },
      ]),
      setError,
      { nome: "nome", valor_bruto_cents: "valor" },
    );

    expect(setError).toHaveBeenCalledWith(
      "nome",
      expect.objectContaining({ type: "required", message: expect.any(String) }),
    );
    expect(setError).toHaveBeenCalledWith(
      "valor",
      expect.objectContaining({ type: "too_small", message: expect.any(String) }),
    );
    expect(setError.mock.calls.map(([name]) => name)).not.toContain("root");
  });

  it("dá a cada campo uma mensagem em português não vazia", () => {
    const setError = vi.fn();
    applyProblemToForm<Form>(validationError([{ field: "nome", code: "required" }]), setError, {
      nome: "nome",
    });
    const [, options] = setError.mock.calls[0];
    expect(options.message.trim().length).toBeGreaterThan(0);
    expect(options.message).not.toBe("required");
  });

  it("põe em root o erro de campo que não existe no formulário", () => {
    const setError = vi.fn();

    applyProblemToForm<Form>(
      validationError([
        { field: "nome", code: "required" },
        { field: "parent_id", code: "invalid_format" },
      ]),
      setError,
      { nome: "nome" },
    );

    expect(setError).toHaveBeenCalledWith("nome", expect.objectContaining({ type: "required" }));
    expect(setError).toHaveBeenCalledWith(
      "root",
      expect.objectContaining({ message: expect.any(String) }),
    );
    expect(setError.mock.calls.map(([name]) => name)).not.toContain("parent_id");
  });

  it("põe em root, com a mensagem do catálogo, um erro sem errors[]", () => {
    const setError = vi.fn();
    const conflict = new ApiError({
      status: 409,
      code: "conta_inativa",
      title: "Conflict",
      errors: [],
    });

    applyProblemToForm<Form>(conflict, setError, { nome: "nome" }, {
      catalog: { conta_inativa: "A conta está inativa." },
    });

    expect(setError).toHaveBeenCalledTimes(1);
    expect(setError).toHaveBeenCalledWith(
      "root",
      expect.objectContaining({ type: "conta_inativa", message: "A conta está inativa." }),
    );
  });
});
