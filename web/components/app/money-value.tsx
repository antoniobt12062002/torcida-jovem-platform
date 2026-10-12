import { TriangleAlertIcon } from "lucide-react";

import { formatBRL } from "@/lib/money";
import { cn } from "@/lib/utils";

// Exibição de dinheiro (DSC-06 AC1 a AC3): centavos inteiros formatados por
// `formatBRL`, em algarismos tabulares. O sentido do valor vai sempre em texto
// (sinal, e ícone no saldo negativo), nunca só na cor.

export type MoneyVariant = "neutral" | "income" | "expense" | "balance";

export type MoneyValueProps = {
  /** Valor em centavos inteiros. */
  cents: number;
  /**
   * `neutral` (padrão): o sinal do próprio valor. `income`: "+" e tom de sucesso.
   * `expense`: "−". `balance`: negativo com "−", ícone e tom de erro.
   */
  variant?: MoneyVariant;
  /** `display`: valor de destaque (36/40) na fonte do token `--font-money-display`. */
  size?: "body" | "display";
  className?: string;
};

const MINUS = "−";

function signedText(cents: number, variant: MoneyVariant): string {
  if (variant === "neutral") return formatBRL(cents);
  const amount = formatBRL(Math.abs(cents));
  if (cents === 0) return amount;
  if (variant === "income") return `+${amount}`;
  if (variant === "expense") return `${MINUS}${amount}`;
  return cents < 0 ? `${MINUS}${amount}` : amount;
}

export function MoneyValue({ cents, variant = "neutral", size = "body", className }: MoneyValueProps) {
  const negativeBalance = variant === "balance" && cents < 0;
  return (
    <span
      data-slot="money-value"
      data-variant={variant}
      className={cn(
        "inline-flex items-center gap-1 whitespace-nowrap tabular-nums",
        size === "display" && "gap-2 font-money-display text-4xl leading-10 font-bold",
        variant === "income" && cents !== 0 && "text-success",
        negativeBalance && "text-destructive",
        className,
      )}
    >
      {negativeBalance ? (
        <TriangleAlertIcon aria-hidden="true" className={cn("shrink-0", size === "display" ? "size-6" : "size-4")} />
      ) : null}
      {signedText(cents, variant)}
      {negativeBalance ? <span className="sr-only">saldo negativo</span> : null}
    </span>
  );
}
