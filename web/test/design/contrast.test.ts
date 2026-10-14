import { describe, expect, it } from "vitest";

import { contrastRatio } from "@/lib/design/contrast";
import { CONTRAST_EXEMPT, CONTRAST_MINIMUM, CONTRAST_PAIRS } from "@/lib/design/contrast-pairs";

import { findBlocks, parseCss, readGlobalsCss, resolveToken, rootTokens } from "./css";

const sheet = parseCss(readGlobalsCss());
const tokens = rootTokens(sheet);
const theme = findBlocks(sheet, "@theme inline")[0]?.declarations ?? {};
const colorTokens = Object.keys(theme)
  .filter((name) => name.startsWith("--color-"))
  .map((name) => name.slice("--color-".length));

describe("cálculo de contraste (WCAG)", () => {
  it("preto sobre branco é 21:1 e a ordem dos argumentos não importa", () => {
    expect(contrastRatio("#000000", "#FFFFFF")).toBeCloseTo(21, 5);
    expect(contrastRatio("#FFF", "#000")).toBeCloseTo(21, 5);
  });

  it("reproduz as razões da paleta oficial", () => {
    expect(contrastRatio("#111111", "#FFFFFF")).toBeCloseTo(18.88, 2);
    expect(contrastRatio("#D81E28", "#FFFFFF")).toBeCloseTo(5.08, 2);
    expect(contrastRatio("#D81E28", "#111111")).toBeCloseTo(3.72, 2);
    expect(contrastRatio("#A7A9AC", "#111111")).toBeCloseTo(8.02, 2);
  });
});

describe("contraste dos pares semânticos", () => {
  it("os mínimos são 4,5:1 em texto e 3:1 em gráfico e borda de campo", () => {
    expect(CONTRAST_MINIMUM).toEqual({ text: 4.5, graphic: 3 });
  });

  it.each(CONTRAST_PAIRS.map((pair) => [`${pair.foreground} sobre ${pair.background} (${pair.kind})`, pair] as const))(
    "%s atinge o mínimo",
    (_name, pair) => {
      const ratio = contrastRatio(resolveToken(tokens, pair.foreground), resolveToken(tokens, pair.background));
      expect(ratio).toBeGreaterThanOrEqual(CONTRAST_MINIMUM[pair.kind]);
    },
  );

  it("todo token de texto do tema está em algum par de texto, salvo isenção declarada", () => {
    const textTokens = colorTokens.filter((name) => name === "foreground" || name.endsWith("-foreground"));
    expect(textTokens.length).toBeGreaterThan(10);
    const covered = new Set(CONTRAST_PAIRS.filter((pair) => pair.kind === "text").map((pair) => pair.foreground));
    expect(textTokens.filter((name) => !covered.has(name) && !(name in CONTRAST_EXEMPT))).toEqual([]);
    expect(Object.keys(CONTRAST_EXEMPT)).toEqual([]);
  });

  it("o texto secundário e o de controle desabilitado têm par sobre as superfícies de hover", () => {
    const text = CONTRAST_PAIRS.filter((pair) => pair.kind === "text").map((pair) => `${pair.foreground}/${pair.background}`);
    expect(text).toEqual(expect.arrayContaining(["muted-foreground/accent", "disabled-foreground/disabled"]));
  });

  it("borda de campo, foco e indicadores de estado têm par de gráfico", () => {
    const graphic = CONTRAST_PAIRS.filter((pair) => pair.kind === "graphic").map(
      (pair) => `${pair.foreground}/${pair.background}`,
    );
    expect(graphic).toEqual(
      expect.arrayContaining([
        "input/background",
        "input/muted",
        "input/accent",
        "ring/ring-offset",
        "sidebar-ring/sidebar",
        "sidebar-primary/sidebar",
        "brand/background",
      ]),
    );
  });
});
