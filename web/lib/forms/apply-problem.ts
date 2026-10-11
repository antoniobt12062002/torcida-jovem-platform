import type { FieldValues, Path, UseFormSetError } from "react-hook-form";

import { type ApiError, type ErrorCatalog, messageFor } from "@/lib/api/problem";

// FND-06 AC1: os errors[] de uma resposta problem+json vão para os campos do
// formulário (React Hook Form); o que não tem campo vira erro do formulário
// (root). Sem biblioteca de schema (WEB-D-003).

/** Mensagens por `code` de erro de campo, como a API os emite. */
export const FIELD_MESSAGES: ErrorCatalog = {
  required: "Campo obrigatório.",
  invalid_type: "Valor em formato inválido.",
  invalid_enum: "Opção inválida.",
  invalid_format: "Formato inválido.",
  too_small: "Valor abaixo do mínimo permitido.",
  too_large: "Valor acima do máximo permitido.",
  invalid_value: "Valor inválido.",
};

/** Campo da API (`errors[].field`) → nome do campo no formulário. */
export type ProblemFieldMap<T extends FieldValues> = Readonly<Record<string, Path<T>>>;

export type ApplyProblemOptions = {
  /** Mensagens da tela por `code` do erro, para o que vai para `root`. */
  catalog?: ErrorCatalog;
  /** Mensagens de campo por `code`, somadas às padrão. */
  fieldMessages?: ErrorCatalog;
};

export function applyProblemToForm<T extends FieldValues>(
  error: ApiError,
  setError: UseFormSetError<T>,
  fieldMap: ProblemFieldMap<T>,
  { catalog, fieldMessages }: ApplyProblemOptions = {},
): void {
  const messages = { ...FIELD_MESSAGES, ...fieldMessages };
  let unplaced = error.errors.length === 0;

  for (const { field, code } of error.errors) {
    const name = Object.hasOwn(fieldMap, field) ? fieldMap[field] : undefined;
    if (!name) {
      unplaced = true;
      continue;
    }
    const message = Object.hasOwn(messages, code) ? messages[code] : FIELD_MESSAGES.invalid_value;
    setError(name, { type: code, message });
  }

  if (unplaced) {
    setError("root", { type: error.code, message: messageFor(error, catalog) });
  }
}
