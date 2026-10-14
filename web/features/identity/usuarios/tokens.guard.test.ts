import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

import { findRawValues, WEB_ROOT } from "@/test/design/guard";

// Guarda da frente DS-ID (DSM-02 AC1 e AC5; T2 de tasks/05-migracao-web-v1.md):
// nenhum valor visual solto, nenhum alias de variante do kit e nenhum
// `<select>`, contêiner ou título montado à mão nos diretórios de Identity
// migrados nesta rodada.

const OWNED = ["features/identity/usuarios", "features/identity/nav.ts", "app/(app)/admin"];

function sourcesOf(path: string): string[] {
  const full = join(WEB_ROOT, path);
  if (!statSync(full).isDirectory()) return [path];
  return readdirSync(full).flatMap((name) => sourcesOf(`${path}/${name}`));
}

// Código de tela, sem os testes (eles citam nomes nas asserções).
const screens = OWNED.flatMap(sourcesOf).filter((file) => /\.tsx?$/.test(file) && !/\.test\.tsx?$/.test(file));
const read = (file: string) => readFileSync(join(WEB_ROOT, file), "utf8");

// Alias de compatibilidade do kit, removidos em DS-INT: `Button` com
// `variant="default" | "outline"` ou `size="default" | "icon" | "icon-sm"`;
// `Alert` sem variante ou com `variant="default" | "destructive"`; status por texto.
const ALIASES: readonly { name: string; pattern: RegExp }[] = [
  { name: "variante de botão", pattern: /variant="(?:default|outline)"/ },
  { name: "tamanho de botão", pattern: /size="(?:default|icon|icon-sm)"/ },
  { name: "Alert sem tom", pattern: /<Alert(?:\s*>|\s+(?!\s*variant=)[^>]*>)/ },
  { name: "Alert com alias", pattern: /<Alert[^>]*variant="(?:default|destructive)"/ },
  { name: "status por texto", pattern: /\bactiveStatus\(/ },
];

describe("guarda da frente DS-ID", () => {
  it("encontra o código das telas de usuários", () => {
    expect(screens).toEqual(
      expect.arrayContaining([
        "features/identity/usuarios/users-page.tsx",
        "features/identity/usuarios/users-table.tsx",
        "features/identity/usuarios/dialogs/admin-dialogs.tsx",
        "features/identity/usuarios/dialogs/create-user-dialog.tsx",
        "features/identity/usuarios/dialogs/status-dialogs.tsx",
        "features/identity/usuarios/dialogs/temporary-password-dialog.tsx",
        "features/identity/nav.ts",
        "app/(app)/admin/usuarios/page.tsx",
      ]),
    );
  });

  it("nenhum valor visual solto", () => {
    const found = findRawValues(OWNED);
    expect(found.map(({ file, line, rule, match }) => `${file}:${line} ${rule} ${match}`)).toEqual([]);
  });

  it("nenhum alias de variante do kit", () => {
    const hits = screens.flatMap((file) =>
      ALIASES.filter(({ pattern }) => pattern.test(read(file))).map(({ name }) => `${file}: ${name}`),
    );
    expect(hits).toEqual([]);
  });

  it("nenhum <select>, <main> ou <h1> montado à mão", () => {
    const hits = screens.filter((file) => /<(?:select|main|h1)[\s>]/.test(read(file)));
    expect(hits).toEqual([]);
  });

  it("a verificação de alias aponta um uso real (controle)", () => {
    const sample = '<Button variant="outline">x</Button>\n<Alert>\n<Alert variant="destructive">';
    expect(ALIASES.filter(({ pattern }) => pattern.test(sample)).map(({ name }) => name)).toEqual([
      "variante de botão",
      "Alert sem tom",
      "Alert com alias",
    ]);
    const kit = '<Button variant="secondary">x</Button>\n<Alert\n  variant="warning">\n<AlertDescription>';
    expect(ALIASES.filter(({ pattern }) => pattern.test(kit))).toEqual([]);
  });
});
