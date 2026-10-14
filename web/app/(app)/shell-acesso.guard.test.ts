import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

import { findRawValues, WEB_ROOT, type GuardException } from "@/test/design/guard";

// Guarda da unidade DS-S (DSF-01 AC7, DSM-05 AC1): aplica a guarda de design
// de DS-F aos layouts, telas de acesso, páginas de erro e início. Os arquivos
// de teste ficam fora: citam classes nas próprias asserções.

const PATHS = [
  "app/(public)",
  "app/(app)/layout.tsx",
  "app/(app)/error.tsx",
  "app/(app)/inicio",
  "app/(app)/sem-acesso",
  "app/(app)/conta",
  "app/not-found.tsx",
  "app/global-error.tsx",
  "app/error-heading.tsx",
  "features/identity/acesso",
] as const;

// Nenhuma exceção: global-error importa a folha global, sem estilos locais.
const EXCEPTIONS: GuardException[] = [];

function files(path: string): string[] {
  const full = join(WEB_ROOT, path);
  if (!statSync(full).isDirectory()) return [path];
  return readdirSync(full).flatMap((name) => files(`${path}/${name}`));
}

const sources = PATHS.flatMap(files).filter((file) => /\.tsx?$/.test(file) && !/\.test\.tsx?$/.test(file));

describe("guarda de design de DS-S", () => {
  it("lê os arquivos da unidade", () => {
    expect(sources).toEqual(
      expect.arrayContaining([
        "app/(public)/layout.tsx",
        "app/(public)/entrar/login-form.tsx",
        "app/(app)/inicio/atalhos.tsx",
        "app/(app)/error.tsx",
        "app/global-error.tsx",
        "features/identity/acesso/change-password-form.tsx",
      ]),
    );
  });

  it("nenhum arquivo tem valor visual solto", () => {
    const found = findRawValues(sources, { exceptions: EXCEPTIONS });
    expect(found.map(({ file, line, rule, match }) => `${file}:${line} ${rule} ${match}`)).toEqual([]);
    expect(EXCEPTIONS).toEqual([]);
  });

  it("sem contêiner nem título montados à mão: `main` e `h1` vêm dos componentes", () => {
    const handmade = sources.filter((file) => /<(?:main|h1)[\s>]/.test(readFileSync(join(WEB_ROOT, file), "utf8")));
    expect(handmade).toEqual([]);
  });

  // WDS-D-020 e WDS-D-027: o nome do sistema só aparece no título da aba
  // ("Página | TJ Platform"); "Desde 2013" fica só dentro do escudo.
  it("nenhuma tela usa o nome do sistema como marca nem repete o ano do escudo", () => {
    const found = sources.filter((file) => /(?<!\| )TJ Platform|Desde 2013/.test(readFileSync(join(WEB_ROOT, file), "utf8")));
    expect(found).toEqual([]);
  });
});
