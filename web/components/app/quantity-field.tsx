"use client";

import { type NumericFieldProps, NumericTextField, type ParseResult } from "./numeric-text-field";

// Campo de quantidade (FND-05 AC5): só inteiros. `positive` aceita só maiores
// que zero; `nonZero` aceita sinal (+ ou -) e recusa só o zero. Vazio ou
// inválido entrega null.

export type QuantityMode = "positive" | "nonZero";

export const MESSAGE_QUANTITY_NOT_INTEGER = "Informe um número inteiro.";
export const MESSAGE_QUANTITY_NOT_POSITIVE = "Informe um número maior que zero.";
export const MESSAGE_QUANTITY_ZERO = "Informe um número diferente de zero.";
export const MESSAGE_QUANTITY_TOO_LARGE = "Número grande demais.";

const INTEGER = /^[+-]?\d+$/;

function parserFor(mode: QuantityMode) {
  return (text: string): ParseResult => {
    if (text === "") return { value: null, error: null };
    if (!INTEGER.test(text)) return { value: null, error: MESSAGE_QUANTITY_NOT_INTEGER };
    const value = Number(text);
    if (!Number.isSafeInteger(value)) return { value: null, error: MESSAGE_QUANTITY_TOO_LARGE };
    if (mode === "positive" && value <= 0) return { value: null, error: MESSAGE_QUANTITY_NOT_POSITIVE };
    if (mode === "nonZero" && value === 0) return { value: null, error: MESSAGE_QUANTITY_ZERO };
    return { value, error: null };
  };
}

const parsers: Record<QuantityMode, (text: string) => ParseResult> = {
  positive: parserFor("positive"),
  nonZero: parserFor("nonZero"),
};

export type QuantityFieldProps = NumericFieldProps & { mode: QuantityMode };

export function QuantityField({ mode, ...props }: QuantityFieldProps) {
  return (
    <NumericTextField
      {...props}
      parse={parsers[mode]}
      format={String}
      inputMode={mode === "positive" ? "numeric" : "text"}
    />
  );
}
