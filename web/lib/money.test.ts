import { describe, expect, it } from "vitest";

import vectors from "../../api/internal/platform/money/testdata/vectors.json";
import { MoneyFormatError, formatBRL, parseBRL } from "./money";

describe("formatBRL (mesmos vetores do Go)", () => {
  it.each(vectors.format)("formata $cents como $text", ({ cents, text }) => {
    expect(formatBRL(cents)).toBe(text);
  });

  it("usa espaço comum, nunca espaço não separável", () => {
    for (const { cents } of vectors.format) {
      const out = formatBRL(cents);
      expect(out).not.toContain(" ");
      expect(out).not.toContain(" ");
    }
  });

  it("trata -0 como zero", () => {
    expect(formatBRL(-0)).toBe("R$ 0,00");
  });
});

describe("parseBRL (mesmos vetores do Go)", () => {
  it.each(vectors.parse_valid)("lê $text como $cents", ({ text, cents }) => {
    expect(parseBRL(text)).toBe(cents);
  });

  it.each(vectors.parse_invalid)("rejeita %j", (text) => {
    expect(() => parseBRL(text)).toThrow(MoneyFormatError);
  });

  it("devolve RangeError acima do inteiro seguro", () => {
    for (const text of ["90.071.992.547.409,92", "92.233.720.368.547.758,08", "999999999999999999999999"]) {
      expect(() => parseBRL(text)).toThrow(RangeError);
    }
  });

  it("aceita exatamente o maior inteiro seguro", () => {
    expect(parseBRL("90.071.992.547.409,91")).toBe(Number.MAX_SAFE_INTEGER);
  });

  it("volta ao valor original depois de formatar (sem o prefixo R$)", () => {
    let seed = 20260926;
    const next = () => {
      seed = (seed * 1103515245 + 12345) % 2147483648;
      return seed;
    };
    for (let i = 0; i < 2000; i++) {
      const cents = next() * 4194304 + next() * 2; // até ~9e15, dentro do inteiro seguro
      if (!Number.isSafeInteger(cents)) continue;
      const text = formatBRL(cents).replace("R$ ", "");
      expect(parseBRL(text)).toBe(cents);
    }
  });
});

describe("números que não são inteiros seguros", () => {
  const bad: unknown[] = [1.5, NaN, Infinity, -Infinity, 2 ** 53, -(2 ** 53), Number.MAX_SAFE_INTEGER + 2, "12", null, undefined];

  it.each(bad)("formatBRL lança RangeError para %s", (value) => {
    expect(() => formatBRL(value as number)).toThrow(RangeError);
  });
});
