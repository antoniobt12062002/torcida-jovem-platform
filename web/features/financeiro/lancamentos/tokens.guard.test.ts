import { readdirSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

import { findRawValues, WEB_ROOT } from "@/test/design/guard";

// Guarda de tokens da frente DS-FB (DSF-01 AC7, DSM-05 AC4): aplica todas as
// regras da guarda de design aos lançamentos, aos comprovantes e às rotas de
// lançamentos. Os testes ficam fora: citam classes nas próprias asserções.

const ROOTS = [
  "features/financeiro/lancamentos",
  "features/financeiro/comprovantes",
  "app/(app)/financeiro/lancamentos",
];

function sources(dir: string): string[] {
  return readdirSync(join(WEB_ROOT, dir), { withFileTypes: true }).flatMap((entry) => {
    const path = `${dir}/${entry.name}`;
    if (entry.isDirectory()) return sources(path);
    return /\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name) ? [path] : [];
  });
}

const files = ROOTS.flatMap(sources);

describe("guarda de tokens de lançamentos e comprovantes", () => {
  it("encontra os arquivos da frente", () => {
    expect(files).toEqual(
      expect.arrayContaining([
        "features/financeiro/lancamentos/lancamentos-page.tsx",
        "features/financeiro/lancamentos/lancamento-detail.tsx",
        "features/financeiro/lancamentos/lancamento-form.tsx",
        "features/financeiro/comprovantes/comprovantes-panel.tsx",
        "app/(app)/financeiro/lancamentos/page.tsx",
        "app/(app)/financeiro/lancamentos/[id]/page.tsx",
      ]),
    );
  });

  it("nenhum arquivo tem valor visual solto", () => {
    const found = findRawValues(files);
    expect(found.map(({ file, line, rule, match }) => `${file}:${line} ${rule} ${match}`)).toEqual([]);
  });

  it("a guarda lê de fato os arquivos: um valor solto seria apontado", () => {
    const found = findRawValues(["components/ui/typography.tsx"]);
    expect(found.length).toBeGreaterThan(0);
  });
});
