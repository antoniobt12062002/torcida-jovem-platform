import { describe, expect, it } from "vitest";

import { findBlocks, parseCss, readGlobalsCss, resolveToken, rootBlocks, rootTokens, stripComments, toNumber } from "./css";
import { findRawValues } from "./guard";

const GLOBALS = "app/globals.css";
const source = readGlobalsCss();
const css = stripComments(source);
const sheet = parseCss(source);
const roots = rootBlocks(sheet);
const tokens = rootTokens(sheet);
const theme = findBlocks(sheet, "@theme inline")[0]?.declarations ?? {};

const HEX = /^#[0-9A-Fa-f]{6}$/;
const names = (record: Record<string, string>) => Object.keys(record);
const blockOf = (token: string) => roots.find((block) => token in block.declarations);
const px = (token: string, from: Record<string, string> = tokens) => toNumber(from[token]);

// Nomes que o kit e as telas já usam, mais os que o design acrescenta.
const SEMANTIC_COLORS = [
  "brand",
  "brand-foreground",
  "background",
  "card",
  "popover",
  "muted",
  "accent",
  "secondary",
  "disabled",
  "sidebar",
  "sidebar-accent",
  "border",
  "input",
  "border-strong",
  "sidebar-border",
  "overlay",
  "foreground",
  "card-foreground",
  "popover-foreground",
  "accent-foreground",
  "secondary-foreground",
  "muted-foreground",
  "disabled-foreground",
  "sidebar-foreground",
  "sidebar-accent-foreground",
  "sidebar-muted-foreground",
  "primary",
  "primary-foreground",
  "primary-hover",
  "sidebar-primary",
  "sidebar-primary-foreground",
  "destructive",
  "destructive-foreground",
  "destructive-subtle",
  "destructive-emphasis",
  "success",
  "success-subtle",
  "success-emphasis",
  "warning",
  "warning-subtle",
  "warning-emphasis",
  "info",
  "info-subtle",
  "ring",
  "ring-offset",
  "sidebar-ring",
];

describe("primitivos", () => {
  it("os quatro primitivos de marca têm exatamente os valores da paleta oficial", () => {
    expect(tokens["--tj-ink"]).toBe("#111111");
    expect(tokens["--tj-red"]).toBe("#D81E28");
    expect(tokens["--tj-white"]).toBe("#FFFFFF");
    expect(tokens["--tj-steel"]).toBe("#A7A9AC");
  });

  it("as escalas derivadas têm os valores do design", () => {
    const scale = (prefix: string) =>
      Object.fromEntries(names(tokens).filter((name) => name.startsWith(prefix)).map((name) => [name.slice(prefix.length), tokens[name]]));
    expect(scale("--tj-neutral-")).toEqual({
      "50": "#F6F6F7",
      "100": "#ECEDEE",
      "200": "#DCDDDF",
      "300": "#C4C6C8",
      "400": "#A7A9AC",
      "500": "#808387",
      "600": "#64676B",
      "700": "#4A4D50",
      "800": "#2A2C2E",
      "900": "#1C1C1E",
      "950": "#111111",
    });
    expect(scale("--tj-red-")).toEqual({ "50": "#FDECEC", "700": "#B3141D", "800": "#8F0F17" });
    expect(scale("--tj-green-")).toEqual({ "50": "#E8F5EC", "700": "#15803D", "800": "#166534" });
    expect(scale("--tj-amber-")).toEqual({ "50": "#FDF3E3", "700": "#B45309", "800": "#92400E" });
  });
});

describe("três camadas", () => {
  it("primitivos, semânticos e tokens de componente ficam em blocos separados", () => {
    const primitive = blockOf("--tj-ink");
    const semantic = blockOf("--background");
    const component = blockOf("--control-height-md");
    expect(primitive).toBeDefined();
    expect(semantic).toBeDefined();
    expect(component).toBeDefined();
    expect(new Set([primitive, semantic, component]).size).toBe(3);
  });

  it("o bloco de primitivos só declara `--tj-*`, com valor literal", () => {
    const primitive = blockOf("--tj-ink")!.declarations;
    expect(names(primitive).filter((name) => !name.startsWith("--tj-"))).toEqual([]);
    expect(names(primitive).filter((name) => !HEX.test(primitive[name]))).toEqual([]);
    expect(names(tokens).filter((name) => name.startsWith("--tj-") && !(name in primitive))).toEqual([]);
  });

  it("nenhum primitivo é exposto no `@theme`", () => {
    expect(names(theme).length).toBeGreaterThan(0);
    expect(names(theme).filter((name) => name.includes("tj-"))).toEqual([]);
    expect(Object.values(theme).filter((value) => value.includes("--tj-"))).toEqual([]);
    expect(Object.values(theme).filter((value) => /#[0-9a-fA-F]{3,8}\b/.test(value))).toEqual([]);
  });

  it("o tema é um único bloco: todo semântico de cor é atribuído nele, sem literal", () => {
    const semantic = blockOf("--background")!.declarations;
    expect(SEMANTIC_COLORS.filter((name) => !(`--${name}` in semantic))).toEqual([]);
    expect(names(semantic).filter((name) => /#[0-9a-fA-F]{3,8}\b|\b(?:rgb|hsl|oklch)a?\(/.test(semantic[name]))).toEqual([]);
    // Nenhum outro bloco, nem media query, reatribui um semântico de cor.
    const elsewhere = [...findBlocks(sheet, ":root")].filter((block) => block.declarations !== semantic);
    expect(elsewhere.flatMap((block) => SEMANTIC_COLORS.filter((name) => `--${name}` in block.declarations))).toEqual([]);
  });
});

describe("tokens semânticos", () => {
  it("o `@theme` expõe todos os nomes semânticos, cada um ligado à sua variável", () => {
    for (const name of SEMANTIC_COLORS) {
      expect(theme[`--color-${name}`], name).toBe(`var(--${name})`);
    }
  });

  it("todo semântico exposto tem valor e resolve numa cor", () => {
    const colors = names(theme).filter((name) => name.startsWith("--color-"));
    expect(colors.length).toBeGreaterThanOrEqual(SEMANTIC_COLORS.length);
    for (const name of colors) {
      const value = resolveToken(tokens, name.slice("--color-".length));
      if (name === "--color-overlay") {
        expect(value).toBe("color-mix(in srgb, var(--tj-ink) 50%, transparent)");
      } else {
        expect(value, name).toMatch(HEX);
      }
    }
  });

  it("toda variável referenciada na folha tem valor", () => {
    // Definidas fora da folha, pelo `next/font`, em `app/layout.tsx`.
    const external = ["--font-geist-sans", "--font-geist-mono", "--font-barlow-condensed"];
    const defined = new Set([...names(tokens), ...names(theme), ...external]);
    const referenced = [...css.matchAll(/var\((--[\w-]+)/g)].map((match) => match[1]);
    expect(referenced.length).toBeGreaterThan(SEMANTIC_COLORS.length);
    expect([...new Set(referenced)].filter((name) => !defined.has(name))).toEqual([]);
  });

  it("cada token de cor tem o valor do design", () => {
    const resolved = Object.fromEntries(
      SEMANTIC_COLORS.filter((name) => name !== "overlay").map((name) => [name, resolveToken(tokens, name)]),
    );
    expect(resolved).toEqual({
      brand: "#D81E28",
      "brand-foreground": "#FFFFFF",
      background: "#FFFFFF",
      card: "#FFFFFF",
      popover: "#FFFFFF",
      muted: "#F6F6F7",
      accent: "#ECEDEE",
      secondary: "#ECEDEE",
      disabled: "#ECEDEE",
      sidebar: "#111111",
      "sidebar-accent": "#2A2C2E",
      border: "#DCDDDF",
      input: "#808387",
      "border-strong": "#111111",
      "sidebar-border": "#2A2C2E",
      foreground: "#111111",
      "card-foreground": "#111111",
      "popover-foreground": "#111111",
      "accent-foreground": "#111111",
      "secondary-foreground": "#111111",
      "muted-foreground": "#64676B",
      "disabled-foreground": "#64676B",
      "sidebar-foreground": "#FFFFFF",
      "sidebar-accent-foreground": "#FFFFFF",
      "sidebar-muted-foreground": "#A7A9AC",
      primary: "#111111",
      "primary-foreground": "#FFFFFF",
      "primary-hover": "#2A2C2E",
      "sidebar-primary": "#D81E28",
      "sidebar-primary-foreground": "#FFFFFF",
      destructive: "#B3141D",
      "destructive-foreground": "#FFFFFF",
      "destructive-subtle": "#FDECEC",
      "destructive-emphasis": "#8F0F17",
      success: "#15803D",
      "success-subtle": "#E8F5EC",
      "success-emphasis": "#166534",
      warning: "#B45309",
      "warning-subtle": "#FDF3E3",
      "warning-emphasis": "#92400E",
      info: "#4A4D50",
      "info-subtle": "#ECEDEE",
      ring: "#111111",
      "ring-offset": "#FFFFFF",
      "sidebar-ring": "#FFFFFF",
    });
  });

  it("a ação primária usa o preto de marca e o erro não é o vermelho de marca", () => {
    expect(resolveToken(tokens, "primary")).toBe(tokens["--tj-ink"]);
    expect(resolveToken(tokens, "destructive")).not.toBe(resolveToken(tokens, "brand"));
    expect(resolveToken(tokens, "destructive")).not.toBe(tokens["--tj-red"]);
  });

  it("o aço de marca não é texto nem borda de campo sobre superfície clara", () => {
    const steel = tokens["--tj-steel"];
    for (const name of ["foreground", "muted-foreground", "disabled-foreground", "input", "border-strong"]) {
      expect(resolveToken(tokens, name), name).not.toBe(steel);
    }
  });
});

describe("tema único", () => {
  it("não há bloco `.dark`, `prefers-color-scheme` nem tokens de gráfico", () => {
    expect(css).not.toMatch(/\.dark\s*\{/);
    expect(css).not.toMatch(/prefers-color-scheme/);
    expect(css).not.toMatch(/--(?:color-)?chart-/);
  });

  it("a variante `dark` continua presa à classe, para não seguir o sistema", () => {
    expect(css).toContain("@custom-variant dark (&:is(.dark *));");
  });

  it("o documento declara `color-scheme: light`", () => {
    const html = findBlocks(sheet, "html");
    expect(html.map((block) => block.declarations["color-scheme"])).toContain("light");
  });
});

describe("tipografia", () => {
  it("`--font-sans` aponta para a Geist e não para si mesma", () => {
    expect(theme["--font-sans"]).toMatch(/^var\(--font-geist-sans\)/);
    expect(theme["--font-sans"]).not.toContain("var(--font-sans)");
    expect(theme["--font-heading"]).toBe(theme["--font-sans"]);
  });

  it("`--font-mono` e `--font-display` apontam para as fontes carregadas, com reserva do sistema", () => {
    expect(theme["--font-mono"]).toMatch(/^var\(--font-geist-mono\), .*monospace$/);
    expect(theme["--font-display"]).toMatch(/^var\(--font-barlow-condensed\), .*sans-serif$/);
    expect(theme["--font-sans"]).toMatch(/sans-serif$/);
  });
});

describe("tokens de componente", () => {
  it("alturas, larguras e margem têm os valores do design", () => {
    expect({
      "control-height-sm": px("--control-height-sm"),
      "control-height-md": px("--control-height-md"),
      "control-height-lg": px("--control-height-lg"),
      "control-font-size": px("--control-font-size"),
      "touch-target": px("--touch-target"),
      "table-head-height": px("--table-head-height"),
      "table-row-height": px("--table-row-height"),
      "shell-header-height": px("--shell-header-height"),
      "shell-nav-width": px("--shell-nav-width"),
      "content-narrow": px("--content-narrow"),
      "content-default": px("--content-default"),
      "content-wide": px("--content-wide"),
      "page-gutter": px("--page-gutter"),
    }).toEqual({
      "control-height-sm": 32,
      "control-height-md": 36,
      "control-height-lg": 44,
      "control-font-size": 14,
      "touch-target": 44,
      "table-head-height": 40,
      "table-row-height": 44,
      "shell-header-height": 56,
      "shell-nav-width": 240,
      "content-narrow": 640,
      "content-default": 1120,
      "content-wide": 1440,
      "page-gutter": 16,
    });
  });

  it("sob `pointer: coarse` os controles sobem a 44 px e o texto de campo a 16 px", () => {
    const coarse = findBlocks(sheet, "@media (pointer: coarse)").flatMap((block) => findBlocks(block, ":root"));
    expect(coarse).toHaveLength(1);
    const touch = { ...tokens, ...coarse[0].declarations };
    expect(px("--control-height-sm", touch)).toBe(44);
    expect(px("--control-height-md", touch)).toBe(44);
    expect(px("--control-height-lg", touch)).toBe(44);
    expect(px("--touch-target", touch)).toBe(44);
    expect(px("--control-font-size", touch)).toBe(16);
    expect(px("--table-head-height", touch)).toBe(40);
    expect(px("--table-row-height", touch)).toBe(48);
  });

  it("a margem de página muda só nos dois pontos de quebra, 768 e 1024 px", () => {
    const widths = sheet.children
      .filter((block) => block.prelude.startsWith("@media") && block.prelude.includes("width"))
      .map((block) => [block.prelude, findBlocks(block, ":root")[0]?.declarations["--page-gutter"]]);
    expect(widths).toEqual([
      ["@media (min-width: 48rem)", "1.5rem"],
      ["@media (min-width: 64rem)", "2rem"],
    ]);
  });

  it("o `@theme` expõe os tokens de componente como utilitários", () => {
    expect(theme).toMatchObject({
      "--spacing-control-sm": "var(--control-height-sm)",
      "--spacing-control-md": "var(--control-height-md)",
      "--spacing-control-lg": "var(--control-height-lg)",
      "--spacing-touch": "var(--touch-target)",
      "--spacing-table-head": "var(--table-head-height)",
      "--spacing-table-row": "var(--table-row-height)",
      "--spacing-shell-header": "var(--shell-header-height)",
      "--spacing-shell-nav": "var(--shell-nav-width)",
      "--spacing-gutter": "var(--page-gutter)",
      "--container-content-narrow": "var(--content-narrow)",
      "--container-content-default": "var(--content-default)",
      "--container-content-wide": "var(--content-wide)",
      "--text-control": "var(--control-font-size)",
    });
  });
});

describe("forma, elevação e movimento", () => {
  it("há três raios, 4, 6 e 8 px, e os passos maiores do kit ficam em 8 px", () => {
    const radius = (step: string) => toNumber(theme[`--radius-${step}`]);
    expect([radius("sm"), radius("md"), radius("lg")]).toEqual([4, 6, 8]);
    expect(["xl", "2xl", "3xl", "4xl"].map(radius)).toEqual([8, 8, 8, 8]);
    expect(names(theme).filter((name) => name.startsWith("--radius-"))).toHaveLength(7);
  });

  it("há dois níveis de sombra, feitos do preto de marca a 12% e a 20%", () => {
    expect(tokens["--elevation-1"]).toBe("0 4px 12px color-mix(in srgb, var(--tj-ink) 12%, transparent)");
    expect(tokens["--elevation-2"]).toBe("0 16px 40px color-mix(in srgb, var(--tj-ink) 20%, transparent)");
    expect(names(theme).filter((name) => name.startsWith("--shadow-")).map((name) => theme[name])).toEqual([
      "var(--elevation-1)",
      "var(--elevation-2)",
    ]);
  });

  it("durações e curvas têm os valores do design", () => {
    expect([px("--duration-fast"), px("--duration-base"), px("--duration-slow")]).toEqual([120, 200, 300]);
    expect(theme).toMatchObject({
      "--transition-duration-fast": "var(--duration-fast)",
      "--transition-duration-base": "var(--duration-base)",
      "--transition-duration-slow": "var(--duration-slow)",
      "--ease-out": "cubic-bezier(0.2, 0, 0, 1)",
      "--ease-in": "cubic-bezier(0.4, 0, 1, 1)",
    });
  });

  it("com redução de movimento, animações e transições de todos os elementos ficam em 0,01 ms", () => {
    const base = findBlocks(sheet, "@layer base");
    const reduced = base.flatMap((block) => findBlocks(block, "@media (prefers-reduced-motion: reduce)"));
    expect(reduced).toHaveLength(1);
    expect(reduced[0].children).toHaveLength(1);
    const rule = reduced[0].children[0];
    expect(rule.prelude.split(",").map((part) => part.trim())).toEqual(["*", "::before", "::after"]);
    expect(rule.declarations["animation-duration"]).toBe("0.01ms !important");
    expect(rule.declarations["transition-duration"]).toBe("0.01ms !important");
  });

  it("a folha não tem gradiente nem `transition: all`", () => {
    expect(css).not.toMatch(/gradient\(/);
    expect(findRawValues([GLOBALS]).filter((hit) => hit.rule === "transition-all")).toEqual([]);
  });
});

describe("guarda de design aplicada à folha global", () => {
  const exception = {
    file: GLOBALS,
    rules: ["literal-color"] as const,
    reason: "Folha de tokens: único arquivo onde os valores literais da paleta são declarados.",
  };

  it("com a exceção declarada, a folha não tem nenhum valor solto", () => {
    expect(findRawValues([GLOBALS], { exceptions: [exception] })).toEqual([]);
  });

  it("sem a exceção, os únicos apontamentos são os literais dos primitivos `--tj-*`", () => {
    const hits = findRawValues([GLOBALS]);
    const lines = source.split(/\r?\n/);
    expect(hits.length).toBe(names(tokens).filter((name) => name.startsWith("--tj-")).length);
    expect(hits.filter((hit) => hit.rule !== "literal-color")).toEqual([]);
    expect(hits.filter((hit) => !/^\s*--tj-[\w-]+:\s*#[0-9A-Fa-f]{6};$/.test(lines[hit.line - 1]))).toEqual([]);
  });
});
