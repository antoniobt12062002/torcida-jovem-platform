import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

// Fronteira servidor/cliente do App Router. Um Server Component que importa
// de um módulo "use client" um valor que não é componente recebe uma
// referência opaca: `CONSTANTE.campo` vira `undefined` no servidor e a página
// quebra só no build de produção (o jsdom não tem essa fronteira). Foi o que
// derrubou /financeiro/contas. Este teste impede a volta do padrão.

const WEB = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const APP = join(WEB, "app");

function walk(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const full = join(dir, name);
    return statSync(full).isDirectory() ? walk(full) : [full];
  });
}

function isClientModule(source: string) {
  return /^\s*(["'])use client\1/.test(source.replace(/^﻿/, ""));
}

function resolveModule(specifier: string, from: string): string | null {
  const base = specifier.startsWith("@/")
    ? join(WEB, specifier.slice(2))
    : specifier.startsWith(".")
      ? resolve(dirname(from), specifier)
      : null;
  if (!base) return null;
  const candidates = [".tsx", ".ts", "/index.tsx", "/index.ts"].map((ext) => base + ext);
  return candidates.find((file) => existsSync(file)) ?? null;
}

type Violation = { file: string; name: string; from: string };

function violationsOf(file: string): Violation[] {
  const source = readFileSync(file, "utf8");
  if (isClientModule(source)) return [];
  const found: Violation[] = [];
  const imports = source.matchAll(/import\s+(type\s+)?\{([^}]*)\}\s+from\s+["']([^"']+)["']/g);
  for (const [, typeOnly, names, specifier] of imports) {
    if (typeOnly) continue;
    const target = resolveModule(specifier, file);
    if (!target || !isClientModule(readFileSync(target, "utf8"))) continue;
    for (const raw of names.split(",")) {
      const name = raw.trim().split(/\s+as\s+/)[0].trim();
      if (!name || name.startsWith("type ")) continue;
      // Componentes são PascalCase. Constantes, objetos e funções não cruzam a fronteira.
      if (!/^[A-Z][a-z0-9]/.test(name)) {
        found.push({ file: relative(WEB, file).replaceAll("\\", "/"), name, from: specifier });
      }
    }
  }
  return found;
}

const serverFiles = walk(APP).filter((file) => /\.tsx?$/.test(file) && !/\.test\.tsx?$/.test(file));

describe("fronteira servidor/cliente em app/", () => {
  it("há arquivos de servidor para conferir", () => {
    expect(serverFiles.filter((file) => !isClientModule(readFileSync(file, "utf8"))).length).toBeGreaterThan(5);
  });

  it("nenhum arquivo de servidor importa valor que não é componente de um módulo \"use client\"", () => {
    expect(serverFiles.flatMap(violationsOf)).toEqual([]);
  });

  it("o detector reconhece o padrão que quebrou /financeiro/contas", () => {
    // Módulo cliente real da feature; o import abaixo é o que existia na página.
    const page = join(APP, "(app)", "financeiro", "contas", "page.tsx");
    const original = readFileSync(page, "utf8");
    expect(original).not.toMatch(/import\s+\{[^}]*\bPERMISSION\b[^}]*\}\s+from\s+["']@\/features\/financeiro\/contas\/contas-page["']/);
    expect(isClientModule(readFileSync(join(WEB, "features", "financeiro", "contas", "contas-page.tsx"), "utf8"))).toBe(true);
  });
});
