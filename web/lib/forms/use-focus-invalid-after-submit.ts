import { useEffect, useRef } from "react";
import type { FieldValues, Path, UseFormReturn } from "react-hook-form";

// DSP-04 AC3 para erros de campo devolvidos pela API: durante o envio os
// controles ficam desabilitados e o `setError` não consegue focar o campo.
// Quando o envio termina, o foco vai ao primeiro campo inválido de `fields`
// (na ordem do formulário). O erro sem campo fica com o alerta, que se foca.

type FocusForm<T extends FieldValues> = Pick<UseFormReturn<T>, "formState" | "getFieldState" | "setFocus">;

export function useFocusInvalidAfterSubmit<T extends FieldValues>(
  { formState, getFieldState, setFocus }: FocusForm<T>,
  fields: readonly Path<T>[],
): void {
  const { isSubmitting, errors } = formState;
  const wasSubmitting = useRef(false);

  useEffect(() => {
    const finished = wasSubmitting.current && !isSubmitting;
    wasSubmitting.current = isSubmitting;
    if (!finished || errors.root) return;
    const first = fields.find((name) => getFieldState(name, formState).invalid);
    if (first) setFocus(first);
  });
}
