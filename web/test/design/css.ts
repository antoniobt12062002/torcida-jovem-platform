import { readFileSync } from "node:fs";
import { join } from "node:path";

import { WEB_ROOT } from "./guard";

// Leitura mínima da folha global para os testes de tokens: blocos aninhados e
// declarações. Não é um parser de CSS completo; cobre o que a folha usa.

export type CssBlock = {
  /** Seletor ou regra-at, sem a chave. */
  prelude: string;
  declarations: Record<string, string>;
  children: CssBlock[];
};

export function readGlobalsCss() {
  return readFileSync(join(WEB_ROOT, "app", "globals.css"), "utf8");
}

export function stripComments(css: string) {
  return css.replace(/\/\*[\s\S]*?\*\//g, "");
}

export function parseCss(css: string): CssBlock {
  const root: CssBlock = { prelude: "", declarations: {}, children: [] };
  const stack = [root];
  let buffer = "";
  const flushDeclaration = () => {
    const text = buffer.trim();
    buffer = "";
    const colon = text.indexOf(":");
    if (!text || text.startsWith("@") || colon < 0) return;
    stack[stack.length - 1].declarations[text.slice(0, colon).trim()] = text.slice(colon + 1).trim();
  };
  for (const char of stripComments(css)) {
    if (char === "{") {
      const block: CssBlock = { prelude: buffer.trim().replace(/\s+/g, " "), declarations: {}, children: [] };
      buffer = "";
      stack[stack.length - 1].children.push(block);
      stack.push(block);
    } else if (char === "}") {
      flushDeclaration();
      stack.pop();
    } else if (char === ";") {
      flushDeclaration();
    } else {
      buffer += char;
    }
  }
  return root;
}

/** Todos os blocos com o prelúdio dado, em qualquer profundidade. */
export function findBlocks(block: CssBlock, prelude: string): CssBlock[] {
  return block.children.flatMap((child) => [
    ...(child.prelude === prelude ? [child] : []),
    ...findBlocks(child, prelude),
  ]);
}

/** Blocos `:root` de primeiro nível (fora de media query). */
export function rootBlocks(sheet: CssBlock) {
  return sheet.children.filter((block) => block.prelude === ":root");
}

/** Propriedades personalizadas dos blocos `:root` de primeiro nível. */
export function rootTokens(sheet: CssBlock): Record<string, string> {
  return Object.assign({}, ...rootBlocks(sheet).map((block) => block.declarations));
}

/** Segue `var(--x)` até um valor que não é só uma referência. */
export function resolveToken(tokens: Record<string, string>, name: string, seen: string[] = []): string {
  const key = name.startsWith("--") ? name : `--${name}`;
  if (seen.includes(key)) throw new Error(`Referência circular: ${[...seen, key].join(" -> ")}`);
  const value = tokens[key];
  if (value === undefined) throw new Error(`Token sem valor: ${key}`);
  const reference = /^var\((--[\w-]+)\)$/.exec(value);
  return reference ? resolveToken(tokens, reference[1], [...seen, key]) : value;
}

/** Converte `rem`, `px` e `ms` em número (1 rem = 16 px). */
export function toNumber(value: string) {
  const match = /^(-?[\d.]+)(rem|px|ms)$/.exec(value.trim());
  if (!match) throw new Error(`Medida não reconhecida: ${value}`);
  return match[2] === "rem" ? Number(match[1]) * 16 : Number(match[1]);
}
