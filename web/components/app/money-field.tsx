"use client";

import { formatBRL, MoneyFormatError, parseBRL } from "@/lib/money";

import { type NumericFieldProps, NumericTextField, type ParseResult } from "./numeric-text-field";

// Campo de dinheiro (FND-05 AC4): lê o valor pt-BR com parseBRL e entrega ao
// formulário centavos inteiros, nunca número de ponto flutuante. Vazio ou
// inválido entrega null.

export const MESSAGE_MONEY_INVALID = "Valor inválido. Use o formato 1.234,56.";
export const MESSAGE_MONEY_TOO_LARGE = "Valor grande demais.";

function parseMoney(text: string): ParseResult {
  if (text === "") return { value: null, error: null };
  try {
    return { value: parseBRL(text), error: null };
  } catch (error) {
    if (error instanceof MoneyFormatError) return { value: null, error: MESSAGE_MONEY_INVALID };
    if (error instanceof RangeError) return { value: null, error: MESSAGE_MONEY_TOO_LARGE };
    throw error;
  }
}

function formatMoney(cents: number): string {
  return formatBRL(cents).replace("R$ ", "");
}

export type MoneyFieldProps = NumericFieldProps;

/**
 * Uso com React Hook Form:
 * `<Controller name="valor_cents" control={control} render={({ field, fieldState }) =>
 *   <MoneyField label="Valor" {...field} error={fieldState.error?.message} />} />`
 * Para exigir um valor válido: `rules={{ validate: (v) => v !== null || "Informe o valor." }}`.
 */
export function MoneyField({ placeholder = "0,00", ...props }: MoneyFieldProps) {
  return (
    <NumericTextField
      {...props}
      placeholder={placeholder}
      parse={parseMoney}
      format={formatMoney}
      inputMode="decimal"
      prefix="R$"
    />
  );
}
