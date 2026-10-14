import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { formatBRL } from "@/lib/money";

import { MoneyValue } from "./money-value";

// DSC-06 AC1 a AC3: centavos inteiros por formatBRL, algarismos tabulares,
// sinal em texto. Nenhuma variante depende só de cor.

const MINUS = "−";

function value(container: HTMLElement) {
  return container.querySelector("[data-slot=money-value]") as HTMLElement;
}

describe("variante neutra (padrão)", () => {
  it.each([
    [-1050, "-R$ 10,50"],
    [0, "R$ 0,00"],
    [19850, "R$ 198,50"],
  ])("%i centavos vira %s, como formatBRL", (cents, text) => {
    const { container } = render(<MoneyValue cents={cents} />);
    expect(value(container).textContent).toBe(text);
    expect(value(container).textContent).toBe(formatBRL(cents));
    expect(value(container).getAttribute("data-variant")).toBe("neutral");
  });

  it("usa algarismos tabulares e não quebra entre R$ e o número", () => {
    const { container } = render(<MoneyValue cents={19850} />);
    const classes = value(container).className.split(/\s+/);
    expect(classes).toContain("tabular-nums");
    expect(classes).toContain("whitespace-nowrap");
  });
});

describe("receita", () => {
  it('leva o sinal "+" em texto, além do tom de sucesso', () => {
    const { container } = render(<MoneyValue cents={19850} variant="income" />);
    expect(value(container).textContent).toBe("+R$ 198,50");
    expect(value(container).className).toContain("text-success");
  });

  it("zero não leva sinal", () => {
    const { container } = render(<MoneyValue cents={0} variant="income" />);
    expect(value(container).textContent).toBe("R$ 0,00");
  });
});

describe("despesa", () => {
  it('leva o sinal "−" (U+2212) em texto, em tom neutro', () => {
    const { container } = render(<MoneyValue cents={1050} variant="expense" />);
    expect(value(container).textContent).toBe(`${MINUS}R$ 10,50`);
    expect(value(container).className).not.toContain("text-success");
    expect(value(container).className).not.toContain("text-destructive");
  });

  it("zero não leva sinal", () => {
    const { container } = render(<MoneyValue cents={0} variant="expense" />);
    expect(value(container).textContent).toBe("R$ 0,00");
  });
});

describe("saldo", () => {
  it("negativo: sinal de menos, ícone e texto para leitor de tela, além do tom de erro", () => {
    const { container } = render(<MoneyValue cents={-1050} variant="balance" />);
    const element = value(container);
    expect(element.textContent).toContain(`${MINUS}R$ 10,50`);
    expect(element.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true");
    const hidden = screen.getByText("saldo negativo");
    expect(hidden.classList.contains("sr-only")).toBe(true);
    expect(element.contains(hidden)).toBe(true);
    expect(element.className).toContain("text-destructive");
  });

  it.each([
    [0, "R$ 0,00"],
    [19850, "R$ 198,50"],
  ])("%i centavos: sem sinal, sem ícone e sem aviso", (cents, text) => {
    const { container } = render(<MoneyValue cents={cents} variant="balance" />);
    expect(value(container).textContent).toBe(text);
    expect(value(container).querySelector("svg")).toBeNull();
    expect(screen.queryByText("saldo negativo")).toBeNull();
    expect(value(container).className).not.toContain("text-destructive");
  });
});

describe("tamanho", () => {
  it("display usa a fonte do token de valor de destaque, em 36/40, ainda tabular", () => {
    const { container } = render(<MoneyValue cents={19850} variant="balance" size="display" />);
    const classes = value(container).className.split(/\s+/);
    expect(classes).toEqual(expect.arrayContaining(["font-money-display", "text-4xl", "leading-10", "tabular-nums"]));
  });

  it("body (padrão) não usa a fonte de destaque", () => {
    const { container } = render(<MoneyValue cents={19850} />);
    expect(value(container).className).not.toContain("font-money-display");
  });
});

it("recusa valor que não é inteiro em centavos", () => {
  expect(() => render(<MoneyValue cents={10.5} />)).toThrow(RangeError);
});
