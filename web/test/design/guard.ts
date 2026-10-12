import { readdirSync, readFileSync, statSync } from "node:fs";
import { dirname, isAbsolute, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

// Guarda de design: lê arquivos de `web/` e aponta valores visuais soltos,
// isto é, que não vêm de um token. Segue o modelo de
// `app/server-client-boundary.test.ts`: só leitura de arquivos, sem navegador.
// Cada unidade do ciclo aplica a guarda aos caminhos de que é dona.

export const WEB_ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..", "..");

export const GUARD_RULES = [
  "literal-color",
  "raw-palette",
  "arbitrary-color",
  "arbitrary-text-size",
  "arbitrary-spacing",
  "arbitrary-radius",
  "font-display",
  "uppercase",
  "transition-all",
  "dark-variant",
] as const;

export type GuardRule = (typeof GUARD_RULES)[number];

export type GuardViolation = {
  /** Caminho relativo à raiz, com barras normais. */
  file: string;
  /** Linha, começando em 1. */
  line: number;
  rule: GuardRule;
  /** Trecho que violou a regra. */
  match: string;
};

export type GuardException = {
  /** Caminho relativo à raiz, com barras normais. */
  file: string;
  rules: readonly GuardRule[];
  /** Obrigatório: por que este arquivo pode violar estas regras. */
  reason: string;
};

export type GuardOptions = {
  /** Raiz para resolver caminhos relativos e relatar arquivos. Padrão: `web/`. */
  root?: string;
  exceptions?: readonly GuardException[];
};

// Prefixos de utilitário do Tailwind que recebem cor.
const COLOR_PREFIX =
  "(?:bg|text|border(?:-[xytrblse])?|ring|ring-offset|outline|fill|stroke|from|via|to|shadow|inset-shadow|divide|decoration|caret|accent|placeholder)";
const PALETTE =
  "(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)";
const COLOR_FUNCTION = "(?:rgba?|hsla?|oklch|oklab|lab|lch|hwb)\\(";
const SPACING_PREFIX =
  "(?:p[xytrblse]?|m[xytrblse]?|gap(?:-[xy])?|space-[xy]|inset(?:-[xy])?|top|right|bottom|left|start|end|scroll-[mp][xytrblse]?)";
// Início de token: não continua uma palavra nem um nome de variável CSS.
const START = "(?<![\\w-])";

const RULES: Record<GuardRule, RegExp> = {
  // #RGB, #RGBA, #RRGGBB, #RRGGBBAA e funções de cor.
  "literal-color": new RegExp(
    `#(?:[0-9a-fA-F]{8}|[0-9a-fA-F]{6}|[0-9a-fA-F]{3,4})(?![\\w-])|${START}${COLOR_FUNCTION}`,
    "g",
  ),
  // `text-red-500`, `bg-black`, `border-white/10`.
  "raw-palette": new RegExp(
    `${START}${COLOR_PREFIX}-(?:${PALETTE}-(?:50|[1-9]00|950)|black|white)(?![\\w-])`,
    "g",
  ),
  // `bg-[#fff]`, `text-[rgb(0,0,0)]`, `border-[color:red]`, `bg-[var(--x)]`, `bg-(--x)`.
  "arbitrary-color": new RegExp(
    `${START}${COLOR_PREFIX}-(?:\\[(?:#|${COLOR_FUNCTION}|color:|var\\(--|--)|\\(--)[^\\s"'\`]*`,
    "g",
  ),
  // `text-[13px]`, `text-[0.8rem]`, `text-[length:1rem]`.
  "arbitrary-text-size": new RegExp(`${START}text-\\[(?:length:|[\\d.])[^\\s"'\`]*`, "g"),
  // `p-[13px]`, `-mt-[2px]`, `gap-[5px]`.
  "arbitrary-spacing": new RegExp(`${START}-?${SPACING_PREFIX}-\\[[^\\s"'\`]*`, "g"),
  // `rounded-[10px]`, `rounded-t-[min(...)]`.
  "arbitrary-radius": new RegExp(`${START}rounded(?:-[a-z]{1,2})?-\\[[^\\s"'\`]*`, "g"),
  // Classe `font-display`; a variável `--font-display` não conta.
  "font-display": new RegExp(`${START}font-display(?![\\w-])`, "g"),
  // Classe `uppercase` e `text-transform: uppercase`.
  uppercase: new RegExp(`${START}uppercase(?![\\w-])`, "g"),
  // Classe `transition-all` e `transition: all` em CSS.
  "transition-all": new RegExp(`${START}transition-all(?![\\w-])|transition(?:-property)?:\\s*all\\b`, "g"),
  // Variante `dark:` em qualquer posição da cadeia de variantes.
  "dark-variant": new RegExp(`${START}dark:`, "g"),
};

const SCANNED_EXTENSIONS = /\.(?:tsx?|jsx?|mts|css)$/;

function toPosix(path: string) {
  return path.replaceAll("\\", "/");
}

function walk(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    if (name === "node_modules" || name === ".next") return [];
    const full = join(dir, name);
    if (statSync(full).isDirectory()) return walk(full);
    return SCANNED_EXTENSIONS.test(name) ? [full] : [];
  });
}

/** Aplica as regras a um texto. `file` só serve para o relato. */
export function findRawValuesInSource(source: string, file: string): GuardViolation[] {
  const found: GuardViolation[] = [];
  source.split(/\r?\n/).forEach((text, index) => {
    for (const rule of GUARD_RULES) {
      for (const hit of text.matchAll(RULES[rule])) {
        found.push({ file, line: index + 1, rule, match: hit[0] });
      }
    }
  });
  return found;
}

/**
 * Lê os caminhos dados e devolve os valores visuais soltos encontrados.
 * Um diretório é percorrido (arquivos `.ts`, `.tsx`, `.js`, `.jsx`, `.mts` e
 * `.css`); um arquivo citado diretamente é lido qualquer que seja a extensão.
 * As exceções são explícitas e cada uma exige o motivo.
 */
export function findRawValues(paths: readonly string[], options: GuardOptions = {}): GuardViolation[] {
  const root = options.root ?? WEB_ROOT;
  const exceptions = options.exceptions ?? [];
  for (const exception of exceptions) {
    if (!exception.reason.trim()) {
      throw new Error(`Exceção da guarda sem motivo: ${exception.file}`);
    }
  }
  const allowed = (file: string, rule: GuardRule) =>
    exceptions.some((exception) => toPosix(exception.file) === file && exception.rules.includes(rule));

  return paths.flatMap((path) => {
    const full = isAbsolute(path) ? path : join(root, path);
    const files = statSync(full).isDirectory() ? walk(full) : [full];
    return files.flatMap((file) => {
      const name = toPosix(relative(root, file));
      return findRawValuesInSource(readFileSync(file, "utf8"), name).filter(
        (violation) => !allowed(name, violation.rule),
      );
    });
  });
}
