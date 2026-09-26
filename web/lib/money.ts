// Núcleo monetário compartilhado do front. Valores são inteiros em centavos
// (`number` seguro); nada aqui conhece lançamentos, receitas ou outros
// conceitos de domínio. Espelha api/internal/platform/money e usa os mesmos
// vetores de teste.

export class MoneyFormatError extends Error {
  constructor(message = "formato de valor inválido") {
    super(message);
    this.name = "MoneyFormatError";
  }
}

const AMOUNT_PATTERN = /^(?:\d{1,3}(?:\.\d{3})+|\d+)(?:,\d{1,2})?$/;
const MAX_SAFE_DIGITS = String(Number.MAX_SAFE_INTEGER);

function assertSafeInteger(cents: number): void {
  if (!Number.isSafeInteger(cents)) {
    throw new RangeError("o valor em centavos deve ser um inteiro seguro");
  }
}

/** Formata centavos em pt-BR: "R$ 1.234,56" ou "-R$ 1.234,56", com espaço comum. */
export function formatBRL(cents: number): string {
  assertSafeInteger(cents);
  const digits = String(Math.abs(cents)).padStart(3, "0");
  const reais = digits.slice(0, -2).replace(/\B(?=(\d{3})+(?!\d))/g, ".");
  const centavos = digits.slice(-2);
  return `${cents < 0 ? "-" : ""}R$ ${reais},${centavos}`;
}

/**
 * Lê um valor pt-BR não negativo, como "1.234,56", em centavos. Aceita só
 * dígitos, pontos de milhar em grupos de três e uma vírgula com até duas casas.
 * Lança MoneyFormatError para outro formato e RangeError acima do inteiro seguro.
 */
export function parseBRL(text: string): number {
  if (!AMOUNT_PATTERN.test(text)) {
    throw new MoneyFormatError();
  }
  const [whole, fraction = ""] = text.replaceAll(".", "").split(",");
  const digits = (whole + fraction.padEnd(2, "0")).replace(/^0+(?=\d)/, "");
  if (digits.length > MAX_SAFE_DIGITS.length || (digits.length === MAX_SAFE_DIGITS.length && digits > MAX_SAFE_DIGITS)) {
    throw new RangeError("o valor excede o maior inteiro seguro em centavos");
  }
  return Number(digits);
}
