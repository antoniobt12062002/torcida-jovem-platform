import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";

import { describe, expect, it } from "vitest";

import { WEB_ROOT } from "./guard";

// Integridade do escudo oficial. Os hashes abaixo foram calculados a partir do
// arquivo recebido (commit 7c30c6a, `web/public/brand/torcida-jovem-logo.svg`),
// com as quebras de linha normalizadas. O único ajuste permitido é o `viewBox`.
const RECEIVED = {
  paths: 62,
  d: "6377665492ccd5e3e5873826dbd16a210f44b00379560f55d76545e2b15b8897",
  classes: "ce4c9cce46a064a31286c7e0279d560da1cf481836f16e62973b82905cfbffdf",
  style: "2b7a4dea35d8146c4ed0121385cc09502ffc2f47371c3b6b734547071105635c",
  // Arquivo inteiro com o valor do `viewBox` esvaziado.
  everythingButViewBox: "8f238ff745d73f3b675a2bbcbbb0f42fda1bcda764b796d08b0c5fb67bbacf14",
};

// Círculo externo do escudo no desenho recebido (primeiro traçado).
const CIRCLE = { cx: 295.58, cy: 413.86, r: 293.86 };

const sha256 = (text: string) => createHash("sha256").update(text).digest("hex");
const read = (...parts: string[]) => readFileSync(join(WEB_ROOT, ...parts), "utf8").replace(/\r\n/g, "\n");

const LOGO = ["public", "brand", "torcida-jovem-logo.svg"];
const logo = read(...LOGO);
const paths = [...logo.matchAll(/<path\b[^>]*>/g)].map((match) => match[0]);

function walk(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const full = join(dir, name);
    return statSync(full).isDirectory() ? walk(full) : [full];
  });
}

describe("escudo oficial", () => {
  it("os traçados são idênticos aos do arquivo recebido", () => {
    expect(paths).toHaveLength(RECEIVED.paths);
    expect(sha256(JSON.stringify(paths.map((path) => /\sd="([^"]*)"/.exec(path)?.[1])))).toBe(RECEIVED.d);
  });

  it("as cores são idênticas: classes de cada traçado e bloco de estilos", () => {
    expect(sha256(JSON.stringify(paths.map((path) => /\sclass="([^"]*)"/.exec(path)?.[1])))).toBe(RECEIVED.classes);
    expect(sha256(/<style>([\s\S]*?)<\/style>/.exec(logo)?.[1] ?? "")).toBe(RECEIVED.style);
  });

  it("nada além do `viewBox` mudou no arquivo", () => {
    expect(logo.match(/viewBox="/g)).toHaveLength(1);
    expect(sha256(logo.replace(/viewBox="[^"]*"/, 'viewBox=""'))).toBe(RECEIVED.everythingButViewBox);
  });

  it("o `viewBox` é um quadrado centrado no círculo, sem cortar o escudo", () => {
    const [x, y, width, height] = (/viewBox="([^"]*)"/.exec(logo)?.[1] ?? "").split(/\s+/).map(Number);
    expect(width).toBe(height);
    expect(width).toBeGreaterThan(0);
    expect(x + width / 2).toBeCloseTo(CIRCLE.cx, 1);
    expect(y + height / 2).toBeCloseTo(CIRCLE.cy, 1);
    expect(x).toBeLessThanOrEqual(CIRCLE.cx - CIRCLE.r);
    expect(y).toBeLessThanOrEqual(CIRCLE.cy - CIRCLE.r);
    expect(x + width).toBeGreaterThanOrEqual(CIRCLE.cx + CIRCLE.r);
    expect(y + height).toBeGreaterThanOrEqual(CIRCLE.cy + CIRCLE.r);
    // Respiro mínimo: o quadrado não sobra mais de 1% além do círculo.
    expect(width).toBeLessThanOrEqual(CIRCLE.r * 2 * 1.01);
  });
});

describe("ícones da aplicação", () => {
  it("`app/icon.svg` é cópia do escudo versionado", () => {
    expect(read("app", "icon.svg")).toBe(logo);
  });

  it("`app/apple-icon.png` é um PNG de 180 por 180 px", () => {
    const png = readFileSync(join(WEB_ROOT, "app", "apple-icon.png"));
    expect(png.subarray(0, 8).toString("hex")).toBe("89504e470d0a1a0a");
    expect(png.subarray(12, 16).toString("latin1")).toBe("IHDR");
    expect([png.readUInt32BE(16), png.readUInt32BE(20)]).toEqual([180, 180]);
  });

  it("o `favicon.ico` padrão do create-next-app foi removido", () => {
    expect(existsSync(join(WEB_ROOT, "app", "favicon.ico"))).toBe(false);
  });
});

describe("limpeza de `public/`", () => {
  const EXAMPLES = ["file.svg", "globe.svg", "next.svg", "vercel.svg", "window.svg"];

  it("só o escudo fica em `public/`", () => {
    const files = walk(join(WEB_ROOT, "public")).map((file) => relative(WEB_ROOT, file).replaceAll("\\", "/"));
    expect(files).toEqual(["public/brand/torcida-jovem-logo.svg"]);
  });

  it("nenhum código referencia os arquivos de exemplo", () => {
    const sources = ["app", "components", "features", "lib"]
      .flatMap((dir) => walk(join(WEB_ROOT, dir)))
      .filter((file) => /\.(?:tsx?|css|mts)$/.test(file));
    expect(sources.length).toBeGreaterThan(50);
    const pattern = new RegExp(`/(?:${EXAMPLES.map((name) => name.replace(".", "\\.")).join("|")})\\b`);
    expect(sources.filter((file) => pattern.test(readFileSync(file, "utf8")))).toEqual([]);
  });
});
