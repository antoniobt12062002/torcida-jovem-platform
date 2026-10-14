import { readdirSync } from "node:fs";
import { join, relative } from "node:path";

import { describe, expect, it } from "vitest";

import { findRawValues, GUARD_RULES, WEB_ROOT, type GuardException } from "@/test/design/guard";

// Guarda de tokens da frente DS-ES (DSF-01 AC7, DSM-05 AC4): aplica todas as
// regras da guarda de design (DS-F) aos arquivos de `features/estoque` e de
// `app/(app)/estoque`. Os testes ficam fora: citam classes nas asserções.

const DIRS = ["features/estoque", "app/(app)/estoque"];

function sources(dir: string): string[] {
  return readdirSync(join(WEB_ROOT, dir), { recursive: true, encoding: "utf8" })
    .filter((name) => /\.tsx?$/.test(name) && !/\.test\.tsx?$/.test(name))
    .map((name) => relative(WEB_ROOT, join(WEB_ROOT, dir, name)).replaceAll("\\", "/"));
}

const files = DIRS.flatMap(sources);

// Nenhuma exceção. Uma exceção nova entra aqui com o arquivo, as regras e o motivo.
const EXCEPTIONS: GuardException[] = [];

describe("guarda de tokens do estoque", () => {
  it("encontra os arquivos de tela da frente", () => {
    expect(files).toEqual(
      expect.arrayContaining([
        "features/estoque/produtos/produtos-page.tsx",
        "features/estoque/produtos/create-produto-dialog.tsx",
        "features/estoque/produto-detail.tsx",
        "features/estoque/movimentacoes/movimentacoes-panel.tsx",
        "features/estoque/movimentacoes/movimentacao-dialog.tsx",
        "features/estoque/ajustes/ajuste-dialog.tsx",
        "app/(app)/estoque/produtos/page.tsx",
        "app/(app)/estoque/produtos/[id]/page.tsx",
      ]),
    );
  });

  it("nenhum arquivo tem valor visual solto", () => {
    const found = findRawValues(files, { exceptions: EXCEPTIONS });
    expect(found.map(({ file, line, rule, match }) => `${file}:${line} ${rule} ${match}`)).toEqual([]);
  });

  it("sem exceções e com todas as regras de DS-F", () => {
    expect(EXCEPTIONS).toEqual([]);
    expect(GUARD_RULES.length).toBeGreaterThanOrEqual(10);
  });

  it("a guarda lê de fato os arquivos: um valor solto seria apontado", () => {
    const found = findRawValues(["components/ui/typography.tsx"]);
    expect(found.length).toBeGreaterThan(0);
  });

  it("nenhuma tela monta contêiner, título ou select à mão", async () => {
    const { readFileSync } = await import("node:fs");
    for (const file of files) {
      const source = readFileSync(join(WEB_ROOT, file), "utf8");
      expect(source, file).not.toMatch(/<select|<h1|max-w-5xl|text-xl font-semibold/);
    }
  });
});
