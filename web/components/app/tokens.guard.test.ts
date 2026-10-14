import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

import { findRawValues, GUARD_RULES, WEB_ROOT, type GuardException } from "@/test/design/guard";

// Guarda de tokens dos componentes de aplicação (DSF-01 AC4 e AC7, DSC-07):
// aplica todas as regras da guarda de design (DS-F) a cada arquivo de
// `components/app`. Os arquivos de teste ficam fora do escopo: citam os nomes
// das classes proibidas nas próprias asserções.

const APP = "components/app";
const files = readdirSync(join(WEB_ROOT, APP)).filter((name) => /\.tsx?$/.test(name));
const tests = files.filter((name) => /\.test\.tsx?$/.test(name));
const components = files.filter((name) => !/\.test\.tsx?$/.test(name));

// Nenhuma exceção. O design admite a fonte de identidade e a caixa alta no
// shell e na marca, mas os dois chegam a elas pelos papéis de
// `components/ui/typography.tsx` (`DisplayTitle`, `Eyebrow`), sem classe
// própria. Uma exceção nova entra aqui com o arquivo, as regras e o motivo.
const EXCEPTIONS: GuardException[] = [];

describe("guarda de tokens dos componentes de aplicação", () => {
  it("encontra os componentes do diretório", () => {
    expect(components).toEqual(
      expect.arrayContaining([
        "access-layout.tsx",
        "app-shell.tsx",
        "brand.tsx",
        "data-list.tsx",
        "filter-bar.tsx",
        "page-header.tsx",
        "shell-frame.tsx",
        "states.tsx",
      ]),
    );
    expect(components.length).toBeGreaterThanOrEqual(23);
  });

  it("nenhum componente tem valor visual solto", () => {
    const found = findRawValues(
      components.map((name) => `${APP}/${name}`),
      { exceptions: EXCEPTIONS },
    );
    expect(found.map(({ file, line, rule, match }) => `${file}:${line} ${rule} ${match}`)).toEqual([]);
  });

  it("não há exceções: shell e marca usam os papéis tipográficos do kit", () => {
    expect(EXCEPTIONS).toEqual([]);
    for (const name of ["shell-frame.tsx", "brand.tsx", "access-layout.tsx"]) {
      const source = readFileSync(join(WEB_ROOT, APP, name), "utf8");
      expect(source).not.toMatch(/(?<![\w-])(?:font-display|uppercase)(?![\w-])/);
    }
  });

  it("a guarda lê de fato os arquivos: um valor solto num deles seria apontado", () => {
    // Controle: o kit de tipografia, que tem exceção própria, é apontado quando lido sem ela.
    const found = findRawValues(["components/ui/typography.tsx"]);
    expect([...new Set(found.map((hit) => hit.rule))].sort()).toEqual(["font-display", "uppercase"]);
  });

  it("a guarda aplicada é a de DS-F, com todas as regras", () => {
    expect([...GUARD_RULES].sort()).toEqual(
      [
        "arbitrary-color",
        "arbitrary-radius",
        "arbitrary-spacing",
        "arbitrary-text-size",
        "dark-variant",
        "font-display",
        "literal-color",
        "raw-palette",
        "transition-all",
        "uppercase",
      ].sort(),
    );
  });
});

describe("cobertura dos componentes de aplicação", () => {
  const sources = tests.map((name) => readFileSync(join(WEB_ROOT, APP, name), "utf8"));

  it.each(components)("%s é referenciado por pelo menos um teste", (name) => {
    const specifier = name.replace(/\.tsx?$/, "");
    expect(
      sources.some(
        (source) => source.includes(`from "./${specifier}"`) || source.includes(`from "@/components/app/${specifier}"`),
      ),
    ).toBe(true);
  });
});
