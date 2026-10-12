import { readFileSync } from "node:fs";
import { join } from "node:path";

// Lê os tokens de `app/globals.css` para o espécime mostrar os valores reais,
// sem nenhum hexadecimal escrito no código da página. Só servidor e só fora de
// produção (a página chama `notFound()` antes). Não é um parser de CSS
// completo: cobre os blocos de primeiro nível que a folha usa.

export type TokenSheet = {
  /** Propriedades dos blocos `:root` de primeiro nível, na ordem do arquivo (sem `--`). */
  declarations: Record<string, string>;
  /** Nomes das cores expostas ao Tailwind no `@theme inline` (sem `--color-`). */
  themeColors: string[];
};

type Block = { prelude: string; body: string };

function topLevelBlocks(css: string): Block[] {
  const text = css.replace(/\/\*[\s\S]*?\*\//g, "");
  const blocks: Block[] = [];
  let depth = 0;
  let segmentStart = 0;
  let bodyStart = 0;
  let prelude = "";
  for (let index = 0; index < text.length; index++) {
    const char = text[index];
    if (char === "{") {
      if (depth === 0) {
        prelude = text.slice(segmentStart, index).trim().replace(/\s+/g, " ");
        bodyStart = index + 1;
      }
      depth++;
    } else if (char === "}") {
      depth--;
      if (depth === 0) {
        blocks.push({ prelude, body: text.slice(bodyStart, index) });
        segmentStart = index + 1;
      }
    } else if (char === ";" && depth === 0) {
      segmentStart = index + 1;
    }
  }
  return blocks;
}

function declarationsOf(body: string): [string, string][] {
  return body
    .split(";")
    .map((part) => part.trim())
    .filter((part) => part.startsWith("--"))
    .map((part) => {
      const colon = part.indexOf(":");
      return [part.slice(2, colon).trim(), part.slice(colon + 1).trim()];
    });
}

export function parseTokenSheet(css: string): TokenSheet {
  const blocks = topLevelBlocks(css);
  const declarations: Record<string, string> = {};
  for (const block of blocks.filter((item) => item.prelude === ":root")) {
    for (const [name, value] of declarationsOf(block.body)) declarations[name] = value;
  }
  const theme = blocks.find((item) => item.prelude === "@theme inline");
  const themeColors = theme
    ? declarationsOf(theme.body)
        .map(([name]) => name)
        .filter((name) => name.startsWith("color-"))
        .map((name) => name.slice("color-".length))
    : [];
  return { declarations, themeColors };
}

/** Segue `var(--x)` até um valor que não é só uma referência. */
export function resolveToken(declarations: Record<string, string>, name: string, seen: string[] = []): string {
  if (seen.includes(name)) throw new Error(`Referência circular: ${[...seen, name].join(" -> ")}`);
  const value = declarations[name];
  if (value === undefined) throw new Error(`Token sem valor: --${name}`);
  const reference = /^var\(--([\w-]+)\)$/.exec(value);
  return reference ? resolveToken(declarations, reference[1], [...seen, name]) : value;
}

export function readTokenSheet(): TokenSheet {
  return parseTokenSheet(readFileSync(join(process.cwd(), "app", "globals.css"), "utf8"));
}

/** Dados de cor prontos para a seção de cores (objetos simples, serializáveis). */
export type Palette = {
  primitives: { name: string; value: string }[];
  /** Cada cor semântica: a referência declarada e o valor final. */
  semantic: Record<string, { reference: string; value: string }>;
  themeColors: string[];
};

export function paletteOf(sheet: TokenSheet): Palette {
  const primitives = Object.entries(sheet.declarations)
    .filter(([name]) => name.startsWith("tj-"))
    .map(([name, value]) => ({ name, value }));
  const semantic: Palette["semantic"] = {};
  for (const name of sheet.themeColors) {
    semantic[name] = { reference: sheet.declarations[name], value: resolveToken(sheet.declarations, name) };
  }
  return { primitives, semantic, themeColors: sheet.themeColors };
}
