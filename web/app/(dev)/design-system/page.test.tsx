import { render, screen, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { contrastRatio } from "@/lib/design/contrast";
import { CONTRAST_PAIRS } from "@/lib/design/contrast-pairs";
import { parseCss, readGlobalsCss, resolveToken as resolveFromTest, rootTokens } from "@/test/design/css";

const notFound = vi.hoisted(() =>
  vi.fn(() => {
    throw new Error("NEXT_HTTP_ERROR_FALLBACK;404");
  }),
);

vi.mock("next/navigation", async () => ({
  ...(await import("@/lib/session/navigation-mock")).nextNavigationMock,
  notFound,
}));

import { BG, COLOR_GROUP_ORDER, COLOR_GROUPS, TEXT } from "./color-classes";
import DesignSystemPage, { generateMetadata } from "./page";
import { SECTIONS } from "./sections";
import { paletteOf, parseTokenSheet, readTokenSheet } from "./tokens";

afterEach(() => {
  vi.unstubAllEnvs();
  notFound.mockClear();
});

const RATIO = new Intl.NumberFormat("pt-BR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

function renderPage() {
  return render(<DesignSystemPage />);
}

// DSF-06 AC5: em produção a rota responde como não encontrada, antes de
// renderizar qualquer conteúdo ou ler qualquer arquivo.
describe("indisponível em produção", () => {
  it("com NODE_ENV=production chama notFound e não devolve nada", () => {
    vi.stubEnv("NODE_ENV", "production");
    let result: unknown = "não chamado";
    expect(() => {
      result = DesignSystemPage();
    }).toThrow("NEXT_HTTP_ERROR_FALLBACK;404");
    expect(notFound).toHaveBeenCalledTimes(1);
    expect(result).toBe("não chamado");
  });

  it("com NODE_ENV=production os metadados também chamam notFound, sem o título do espécime", () => {
    vi.stubEnv("NODE_ENV", "production");
    expect(() => generateMetadata()).toThrow("NEXT_HTTP_ERROR_FALLBACK;404");
    expect(notFound).toHaveBeenCalledTimes(1);
  });

  it("fora de produção não chama notFound e renderiza o espécime", () => {
    vi.stubEnv("NODE_ENV", "development");
    renderPage();
    expect(notFound).not.toHaveBeenCalled();
    expect(screen.getByRole("heading", { level: 1, name: "Espécime do design system" })).toBeTruthy();
  });
});

describe("estrutura da página", () => {
  it("tem um só h1, um só main e nenhum id repetido", () => {
    const { container } = renderPage();
    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    expect(container.querySelectorAll("main")).toHaveLength(1);
    expect(container.querySelectorAll("main main")).toHaveLength(0);
    const ids = [...container.querySelectorAll("[id]")].map((element) => element.id);
    expect(ids.filter((id, index) => ids.indexOf(id) !== index)).toEqual([]);
  });

  it.each(SECTIONS.map((section) => [section.title, section] as const))(
    "a seção %s existe, com âncora, título h2 e link no índice",
    (_title, section) => {
      const { container } = renderPage();
      const element = container.querySelector(`section#${section.id}`);
      expect(element).not.toBeNull();
      const heading = within(element as HTMLElement).getAllByRole("heading", { level: 2 })[0];
      expect(heading.textContent).toBe(section.title);
      const index = screen.getByRole("navigation", { name: "Seções do espécime" });
      expect(within(index).getByRole("link", { name: section.title }).getAttribute("href")).toBe(`#${section.id}`);
    },
  );

  it("as seções de fundação vêm na ordem do índice", () => {
    const { container } = renderPage();
    const ids = [...container.querySelectorAll("main > section")].map((element) => element.id);
    expect(ids).toEqual(SECTIONS.map((section) => section.id));
  });
});

describe("marca", () => {
  it("mostra o escudo nos quatro tamanhos sobre o claro e sobre o shell, e o nome como texto", () => {
    const { container } = renderPage();
    const section = container.querySelector("section#marca") as HTMLElement;
    const shields = within(section).getAllByRole("img", { name: "Torcida Jovem" });
    expect(shields.map((image) => image.getAttribute("width"))).toEqual(["32", "40", "64", "96", "32", "40", "64", "96"]);
    expect(within(section).getAllByText("Torcida Jovem").length).toBeGreaterThanOrEqual(2);
  });
});

describe("cores", () => {
  const sheet = parseCss(readGlobalsCss());
  const tokens = rootTokens(sheet);

  it("os grupos cobrem todas as cores do tema, cada uma com classe de fundo", () => {
    const { themeColors } = readTokenSheet();
    const grouped = COLOR_GROUP_ORDER.flatMap((group) => COLOR_GROUPS[group]);
    expect([...grouped].sort()).toEqual([...themeColors].sort());
    expect(themeColors.filter((token) => !BG[token])).toEqual([]);
  });

  it("todo par tem amostra: classe de texto no par de texto e de fundo nos dois lados", () => {
    for (const pair of CONTRAST_PAIRS) {
      expect(BG[pair.background], pair.background).toBeTruthy();
      if (pair.kind === "text") expect(TEXT[pair.foreground], pair.foreground).toBeTruthy();
      else expect(BG[pair.foreground], pair.foreground).toBeTruthy();
    }
  });

  it("a leitura dos tokens dá os mesmos valores que a do teste de contraste", () => {
    const { declarations } = parseTokenSheet(readGlobalsCss());
    for (const [name, value] of Object.entries(tokens)) {
      expect(declarations[name.slice(2)], name).toBe(value);
    }
  });

  it.each(CONTRAST_PAIRS.map((pair) => [`${pair.foreground} sobre ${pair.background} (${pair.kind})`, pair] as const))(
    "%s aparece com a razão calculada",
    (_name, pair) => {
      const { container } = renderPage();
      const section = container.querySelector("section#cores") as HTMLElement;
      const ratio = contrastRatio(resolveFromTest(tokens, pair.foreground), resolveFromTest(tokens, pair.background));
      const rows = within(section)
        .getAllByText(`${pair.foreground} sobre ${pair.background}`)
        .map((cell) => cell.closest("tr") as HTMLElement);
      expect(rows.some((row) => row.textContent?.includes(`${RATIO.format(ratio)}:1`))).toBe(true);
    },
  );

  it("mostra todos os primitivos com o valor do arquivo", () => {
    const { container } = renderPage();
    const section = container.querySelector("section#cores") as HTMLElement;
    const { primitives } = paletteOf(readTokenSheet());
    expect(primitives.length).toBeGreaterThanOrEqual(24);
    for (const { name, value } of primitives) {
      const label = within(section).getByText(`--${name}`);
      expect(label.parentElement?.textContent).toContain(value);
    }
  });
});

describe("tipografia", () => {
  it("cada papel aparece com texto real, com acentos e caixa alta pelos papéis de identidade", () => {
    const { container } = renderPage();
    const section = container.querySelector("section#tipografia") as HTMLElement;
    for (const role of [
      "Destaque",
      "Título de página",
      "Título de seção",
      "Rótulo de grupo e sobretítulo",
      "Título de bloco",
      "Corpo",
      "Corpo maior",
      "Rótulo",
      "Legenda",
      "Dado",
      "Valor de destaque",
      "Identificador",
    ]) {
      expect(within(section).getByText(role, { selector: "dt > span" })).toBeTruthy();
    }
    expect(section.querySelector("[data-slot=display-title]")?.textContent).toMatch(/[ÁÂÃÉÍÓÕÚÇáâãéíóõúç]/);
    expect(section.querySelector("[data-slot=eyebrow]")?.textContent).toMatch(/[ÁÂÃÉÍÓÕÚÇáâãéíóõúç]/);
    expect(section.querySelector("[data-slot=money-value]")).not.toBeNull();
  });
});

describe("espaço, raio, borda e sombra", () => {
  it("mostra as amostras com o nome do token e as medidas dos tokens de componente", () => {
    const { container } = renderPage();
    const section = container.querySelector("section#espaco-raio-borda-sombra") as HTMLElement;
    for (const name of ["spacing-4", "rounded-md (--radius-md)", "Nível 2 · shadow-elevation-2", "--control-height-md"]) {
      expect(within(section).getByText(name)).toBeTruthy();
    }
    const { declarations } = readTokenSheet();
    const row = within(section).getByText("--control-height-md").parentElement as HTMLElement;
    expect(row.textContent).toContain(declarations["control-height-md"]);
  });
});
