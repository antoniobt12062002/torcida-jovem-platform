import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

import { findRawValues, WEB_ROOT, type GuardException } from "@/test/design/guard";

// Guarda de design aplicada a todo o grupo `app/(dev)` (DSF-06 AC2): o
// espécime só usa tokens e componentes reais, sem valor visual solto. Os
// próprios testes entram na leitura: nenhum deles cita classe proibida.

const DEV = "app/(dev)";

// Nenhuma exceção: a fonte de identidade e a caixa alta chegam pelos papéis de
// `components/ui/typography.tsx`.
const EXCEPTIONS: GuardException[] = [];

function walk(dir: string): string[] {
  return readdirSync(join(WEB_ROOT, dir), { withFileTypes: true }).flatMap((entry) =>
    entry.isDirectory() ? walk(`${dir}/${entry.name}`) : [`${dir}/${entry.name}`],
  );
}

describe("guarda de design do espécime", () => {
  it("lê os arquivos do espécime", () => {
    const files = walk(DEV);
    expect(files).toEqual(expect.arrayContaining([`${DEV}/design-system/page.tsx`, `${DEV}/design-system/colors-section.tsx`]));
  });

  it("nenhum arquivo de app/(dev) tem valor visual solto", () => {
    const found = findRawValues([DEV], { exceptions: EXCEPTIONS });
    expect(found.map(({ file, line, rule, match }) => `${file}:${line} ${rule} ${match}`)).toEqual([]);
    expect(EXCEPTIONS).toEqual([]);
  });

  it("nenhum componente exclusivo do espécime sai do grupo app/(dev)", () => {
    // Os componentes de seção ficam junto da rota; components/ e features/ não importam o espécime.
    const outside = ["components", "features", "lib"].flatMap((dir) => walk(dir));
    const importers = outside
      .filter((file) => /\.tsx?$/.test(file))
      .filter((file) => readFileSync(join(WEB_ROOT, file), "utf8").includes("design-system/"));
    expect(importers).toEqual([]);
  });
});
