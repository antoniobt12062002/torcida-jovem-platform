import { readFileSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

import { findBlocks, parseCss, readGlobalsCss, rootTokens } from "./css";
import { findRawValues, WEB_ROOT } from "./guard";

const LAYOUT = "app/layout.tsx";
const layout = readFileSync(join(WEB_ROOT, LAYOUT), "utf8");
const sheet = parseCss(readGlobalsCss());
const theme = findBlocks(sheet, "@theme inline")[0]?.declarations ?? {};

/** Texto do objeto de opções passado a uma função de fonte, por exemplo `Geist({ ... })`. */
function fontOptions(loader: string) {
  const match = new RegExp(`=\\s*${loader}\\(\\{([\\s\\S]*?)\\}\\)`).exec(layout);
  return match?.[1] ?? "";
}

/** Nome da constante que recebe a fonte, por exemplo `geistSans`. */
function fontConstant(loader: string) {
  return new RegExp(`const\\s+(\\w+)\\s*=\\s*${loader}\\(`).exec(layout)?.[1];
}

describe("carregamento das fontes", () => {
  it("as três famílias vêm de `next/font/google`", () => {
    const imported = /import\s*\{([^}]*)\}\s*from\s*["']next\/font\/google["']/.exec(layout)?.[1] ?? "";
    expect(imported.split(",").map((name) => name.trim()).filter(Boolean).sort()).toEqual([
      "Barlow_Condensed",
      "Geist",
      "Geist_Mono",
    ]);
  });

  it("a Barlow Condensed carrega os pesos 600 e 700, subconjunto latin, com `display: swap`", () => {
    const options = fontOptions("Barlow_Condensed");
    expect(options).toMatch(/variable:\s*"--font-barlow-condensed"/);
    expect(options).toMatch(/weight:\s*\["600",\s*"700"\]/);
    expect(options).toMatch(/subsets:\s*\["latin"\]/);
    expect(options).toMatch(/display:\s*"swap"/);
  });

  it("Geist e Geist Mono mantêm as suas variáveis", () => {
    expect(fontOptions("Geist")).toMatch(/variable:\s*"--font-geist-sans"/);
    expect(fontOptions("Geist_Mono")).toMatch(/variable:\s*"--font-geist-mono"/);
  });

  it("o elemento `html` recebe as três variáveis de fonte", () => {
    const className = /<html[\s\S]*?className=\{`([^`]*)`\}/.exec(layout)?.[1] ?? "";
    for (const loader of ["Geist", "Geist_Mono", "Barlow_Condensed"]) {
      const constant = fontConstant(loader);
      expect(constant, loader).toBeDefined();
      expect(className, loader).toContain(`\${${constant}.variable}`);
    }
  });

  it("a folha global liga as variáveis a `--font-sans`, `--font-mono` e `--font-display`", () => {
    expect(theme["--font-sans"]).toMatch(/^var\(--font-geist-sans\)/);
    expect(theme["--font-mono"]).toMatch(/^var\(--font-geist-mono\)/);
    expect(theme["--font-display"]).toMatch(/^var\(--font-barlow-condensed\)/);
  });

  it("nenhuma fonte entra como dependência nem por requisição a terceiros", () => {
    const manifest = JSON.parse(readFileSync(join(WEB_ROOT, "package.json"), "utf8")) as {
      dependencies: Record<string, string>;
      devDependencies: Record<string, string>;
    };
    const packages = [...Object.keys(manifest.dependencies), ...Object.keys(manifest.devDependencies)];
    expect(packages.filter((name) => /font|geist|barlow/i.test(name))).toEqual([]);
    expect(layout).not.toMatch(/fonts\.googleapis|fonts\.gstatic|<link/);
  });
});

describe("metadados do documento", () => {
  it("declara `theme-color` com o preto de marca e `color-scheme` claro", () => {
    const viewport = /export const viewport: Viewport = \{([\s\S]*?)\};/.exec(layout)?.[1] ?? "";
    const themeColor = /themeColor:\s*"([^"]+)"/.exec(viewport)?.[1];
    expect(themeColor).toBe("#111111");
    expect(themeColor).toBe(rootTokens(sheet)["--tj-ink"]);
    expect(viewport).toMatch(/colorScheme:\s*"light"/);
    expect(layout).toMatch(/import type \{[^}]*\bViewport\b[^}]*\} from "next"/);
  });

  it("mantém idioma, título no formato \"Página | TJ Platform\" e descrição", () => {
    expect(layout).toMatch(/lang="pt-BR"/);
    expect(layout).toMatch(/default:\s*"TJ Platform"/);
    expect(layout).toMatch(/template:\s*"%s \| TJ Platform"/);
    expect(layout).toMatch(/description:\s*"Gestão institucional da Torcida Jovem do Campo Mourão Futsal\."/);
  });

  it("a única cor literal do layout é a do `theme-color`, que não aceita variável CSS", () => {
    const hits = findRawValues([LAYOUT]);
    expect(hits.map(({ rule, match }) => ({ rule, match }))).toEqual([{ rule: "literal-color", match: "#111111" }]);
    expect(layout.split(/\r?\n/)[hits[0].line - 1]).toMatch(/themeColor/);
    expect(
      findRawValues([LAYOUT], {
        exceptions: [
          {
            file: LAYOUT,
            rules: ["literal-color"],
            reason: "A meta `theme-color` não lê variável CSS; o valor é conferido contra `--tj-ink`.",
          },
        ],
      }),
    ).toEqual([]);
  });
});
