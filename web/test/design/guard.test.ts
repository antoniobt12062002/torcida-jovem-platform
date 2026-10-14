import { describe, expect, it } from "vitest";

import { findRawValues, findRawValuesInSource, GUARD_RULES, type GuardRule } from "./guard";

const FIXTURES = "test/design/fixtures";
const TSX = `${FIXTURES}/violacoes.tsx.txt`;
const CSS = `${FIXTURES}/violacoes.css.txt`;
const OK = `${FIXTURES}/conforme.tsx.txt`;

// Para cada regra: exemplos que passam (quase iguais ao proibido) e exemplos que falham.
const CASES: Record<GuardRule, { pass: string[]; fail: string[] }> = {
  "literal-color": {
    pass: ['href="#conteudo"', "bg-primary", "var(--primary)", "pedido #12"],
    fail: ["#D81E28", "#fff", "#11111180", "rgb(0 0 0)", "rgba(0,0,0,.5)", "hsl(0 0% 7%)", "oklch(0.5 0.2 27)"],
  },
  "raw-palette": {
    pass: ["text-destructive", "bg-muted", "border-border", "text-sm", "bg-success-subtle"],
    fail: ["text-red-500", "bg-zinc-50", "border-neutral-950", "ring-blue-500/40", "bg-black/50", "text-white"],
  },
  "arbitrary-color": {
    pass: ["bg-background/50", "text-foreground", "w-[calc(100%-2rem)]"],
    fail: ["bg-[#D81E28]", "text-[rgb(0,0,0)]", "border-[color:red]", "bg-[var(--tj-red)]", "bg-(--tj-red)"],
  },
  "arbitrary-text-size": {
    pass: ["text-sm", "text-xs", "text-base"],
    fail: ["text-[13px]", "text-[0.8rem]", "text-[length:1rem]"],
  },
  "arbitrary-spacing": {
    pass: ["p-4", "px-2.5", "gap-x-3", "-mt-2", "top-0"],
    fail: ["p-[13px]", "-mt-[2px]", "gap-[5px]", "space-y-[3px]", "top-[7px]"],
  },
  "arbitrary-radius": {
    pass: ["rounded-md", "rounded-lg", "rounded-full"],
    fail: ["rounded-[10px]", "rounded-t-[3px]", "rounded-[min(var(--radius-md),10px)]"],
  },
  "font-display": {
    pass: ["--font-display: var(--font-barlow-condensed)", "var(--font-display)", "font-money-display", "font-sans"],
    fail: ['className="font-display"', "md:font-display"],
  },
  uppercase: {
    pass: ["value.toUpperCase()", "normal-case", "nouppercase"],
    fail: ['className="uppercase"', "text-transform: uppercase;"],
  },
  "transition-all": {
    pass: ["transition-colors", "transition-opacity", "transition: opacity 120ms"],
    fail: ["transition-all", "transition: all 200ms", "transition-property: all"],
  },
  "dark-variant": {
    pass: ["@custom-variant dark (&:is(.dark *));", "darkness", "border-dark"],
    fail: ["dark:bg-input/30", "hover:dark:bg-input"],
  },
};

describe("guarda de design: regras", () => {
  it.each(GUARD_RULES)("%s: os exemplos conformes passam", (rule) => {
    for (const source of CASES[rule].pass) {
      expect(findRawValuesInSource(source, "exemplo").filter((hit) => hit.rule === rule), source).toEqual([]);
    }
  });

  it.each(GUARD_RULES)("%s: cada exemplo proibido falha com arquivo e linha", (rule) => {
    for (const source of CASES[rule].fail) {
      const hits = findRawValuesInSource(`linha um\n${source}`, "exemplo.tsx").filter((hit) => hit.rule === rule);
      expect(hits, source).toHaveLength(1);
      expect(hits[0], source).toMatchObject({ file: "exemplo.tsx", line: 2, rule });
    }
  });
});

describe("guarda de design: arquivos", () => {
  it("o arquivo conforme não tem nenhum apontamento", () => {
    expect(findRawValues([OK])).toEqual([]);
  });

  it("aponta arquivo, linha e regra de cada valor solto do exemplo de tela", () => {
    const found = findRawValues([TSX]).map(({ file, line, rule }) => ({ file, line, rule }));
    expect(found).toEqual([
      { file: TSX, line: 2, rule: "literal-color" },
      { file: TSX, line: 3, rule: "literal-color" },
      { file: TSX, line: 4, rule: "raw-palette" },
      { file: TSX, line: 5, rule: "raw-palette" },
      { file: TSX, line: 6, rule: "arbitrary-color" },
      { file: TSX, line: 7, rule: "arbitrary-text-size" },
      { file: TSX, line: 8, rule: "arbitrary-spacing" },
      { file: TSX, line: 9, rule: "arbitrary-radius" },
      { file: TSX, line: 10, rule: "font-display" },
      { file: TSX, line: 11, rule: "uppercase" },
      { file: TSX, line: 12, rule: "transition-all" },
      { file: TSX, line: 13, rule: "dark-variant" },
    ]);
  });

  it("aponta os valores soltos do exemplo de folha de estilos", () => {
    const found = findRawValues([CSS]).map(({ line, rule, match }) => ({ line, rule, match }));
    expect(found).toEqual([
      { line: 2, rule: "literal-color", match: "hsl(" },
      { line: 3, rule: "literal-color", match: "oklch(" },
      { line: 4, rule: "uppercase", match: "uppercase" },
      { line: 5, rule: "transition-all", match: "transition: all" },
    ]);
  });

  it("só lê os caminhos recebidos", () => {
    expect(findRawValues([OK, CSS]).every((hit) => hit.file === CSS)).toBe(true);
    expect(findRawValues([])).toEqual([]);
  });

  it("percorre um diretório e ignora extensões que não são código nem estilo", () => {
    expect(findRawValues([`${FIXTURES}/pasta`])).toEqual([
      { file: `${FIXTURES}/pasta/tela.ts`, line: 2, rule: "raw-palette", match: "text-red-500" },
    ]);
  });
});

describe("guarda de design: exceções", () => {
  it("uma exceção libera só as regras e o arquivo declarados", () => {
    const found = findRawValues([TSX, CSS], {
      exceptions: [
        { file: TSX, rules: ["font-display", "uppercase"], reason: "exemplo: arquivo de tipografia" },
      ],
    });
    expect(found.filter((hit) => hit.file === TSX).map((hit) => hit.rule)).toEqual([
      "literal-color",
      "literal-color",
      "raw-palette",
      "raw-palette",
      "arbitrary-color",
      "arbitrary-text-size",
      "arbitrary-spacing",
      "arbitrary-radius",
      "transition-all",
      "dark-variant",
    ]);
    expect(found.filter((hit) => hit.file === CSS).map((hit) => hit.rule)).toContain("uppercase");
  });

  it("recusa exceção sem motivo", () => {
    expect(() =>
      findRawValues([TSX], { exceptions: [{ file: TSX, rules: ["uppercase"], reason: " " }] }),
    ).toThrow(/sem motivo/);
  });
});
