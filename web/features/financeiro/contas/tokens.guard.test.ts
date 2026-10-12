import { readdirSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

import { findRawValues, WEB_ROOT } from "@/test/design/guard";

// Guarda da frente DS-FA (DSF-01 AC4, DSC-07): aplica todas as regras da guarda
// de design aos arquivos de plano de contas, saldo e navegação do financeiro e
// às rotas dessas telas. Os arquivos de teste ficam fora: citam nomes de
// classes nas asserções.

const DIRS = [
  "features/financeiro/contas",
  "features/financeiro/saldo",
  "app/(app)/financeiro/contas",
  "app/(app)/financeiro/saldo",
];

const files = [
  ...DIRS.flatMap((dir) =>
    readdirSync(join(WEB_ROOT, dir))
      .filter((name) => /\.tsx?$/.test(name) && !/\.test\.tsx?$/.test(name))
      .map((name) => `${dir}/${name}`),
  ),
  "features/financeiro/nav.ts",
];

describe("guarda de tokens de DS-FA", () => {
  it("lê os arquivos da frente", () => {
    expect(files).toEqual(
      expect.arrayContaining([
        "features/financeiro/contas/contas-page.tsx",
        "features/financeiro/contas/conta-dialogs.tsx",
        "features/financeiro/saldo/saldo-page.tsx",
        "app/(app)/financeiro/contas/page.tsx",
        "app/(app)/financeiro/saldo/page.tsx",
      ]),
    );
  });

  it("nenhum arquivo tem valor visual solto", () => {
    const found = findRawValues(files);
    expect(found.map(({ file, line, rule, match }) => `${file}:${line} ${rule} ${match}`)).toEqual([]);
  });
});
