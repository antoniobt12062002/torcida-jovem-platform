import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Button } from "@/components/ui/button";

import { BrandMark } from "./brand";
import { PageContainer } from "./page-container";
import { PageHeader } from "./page-header";

// Conferido pelo `tsc` (gate Web), sem execução: as escalas são fechadas.
export function recusadoPeloCompilador() {
  return [
    // @ts-expect-error Só 32, 40, 64 e 96 (mínimo de 32 px, DSF-05 AC3).
    <BrandMark key="marca" size={16} />,
    // @ts-expect-error Só "narrow", "default" e "wide" (DSP-01 AC1).
    <PageContainer key="pagina" width="huge">
      x
    </PageContainer>,
  ];
}

// DSF-05 AC2 a AC5: o escudo oficial, sempre 1:1, com largura e altura
// declaradas e texto alternativo conforme haja ou não o nome ao lado.
describe("BrandMark", () => {
  it("sozinho, a imagem tem o texto alternativo da marca", () => {
    render(<BrandMark />);
    const image = screen.getByRole("img", { name: "Torcida Jovem" });
    expect(image.getAttribute("src")).toBe("/brand/torcida-jovem-logo.svg");
    expect(screen.queryByText("Torcida Jovem")).toBeNull();
  });

  it("com o nome, a imagem tem texto alternativo vazio e o nome é texto", () => {
    const { container } = render(<BrandMark withName />);
    const image = container.querySelector("img");
    expect(image?.getAttribute("alt")).toBe("");
    expect(screen.queryByRole("img")).toBeNull();
    expect(screen.getByText("Torcida Jovem")).toBeTruthy();
  });

  it.each([32, 40, 64, 96] as const)("com %i px, declara largura e altura iguais", (size) => {
    const { container } = render(<BrandMark size={size} />);
    const image = container.querySelector("img");
    expect(image?.getAttribute("width")).toBe(String(size));
    expect(image?.getAttribute("height")).toBe(String(size));
  });

  it("sem tamanho informado, usa o mínimo de 32 px", () => {
    const { container } = render(<BrandMark withName />);
    expect(container.querySelector("img")?.getAttribute("width")).toBe("32");
    expect(container.querySelector("img")?.getAttribute("height")).toBe("32");
  });
});

// Revisão da rodada 4: HTML válido com o nome e carregamento acima da dobra.
describe("BrandMark: estrutura e carregamento", () => {
  it("com o nome, nenhum parágrafo fica dentro de elemento inline", () => {
    const { container } = render(<BrandMark withName />);
    const root = container.querySelector("[data-slot=brand-mark]") as HTMLElement;
    expect(root.tagName).toBe("DIV");
    expect(root.className.split(/\s+/)).toContain("inline-flex");
    for (const block of container.querySelectorAll("p, div")) {
      expect(block.closest("span"), block.outerHTML).toBeNull();
    }
  });

  it("com priority, o escudo carrega sem espera e com prioridade alta", () => {
    const { container } = render(<BrandMark priority />);
    const image = container.querySelector("img");
    expect(image?.getAttribute("loading")).toBe("eager");
    expect(image?.getAttribute("fetchpriority")).toBe("high");
  });

  it("sem priority, o escudo segue o carregamento sob demanda", () => {
    const { container } = render(<BrandMark />);
    const image = container.querySelector("img");
    expect(image?.getAttribute("loading")).toBe("lazy");
    expect(image?.getAttribute("fetchpriority")).toBeNull();
  });
});

// DSP-01 AC1: três larguras nomeadas. DSF-04 AC7: alvo do link de pular.
describe("PageContainer", () => {
  it("é o main da página e o alvo do link de pular", () => {
    render(<PageContainer>conteúdo</PageContainer>);
    const main = screen.getByRole("main");
    expect(main.id).toBe("conteudo");
    expect(main.getAttribute("tabindex")).toBe("-1");
    expect(main.textContent).toBe("conteúdo");
  });

  it.each([
    ["narrow", "max-w-content-narrow"],
    ["default", "max-w-content-default"],
    ["wide", "max-w-content-wide"],
  ] as const)("largura %s usa o token %s", (width, token) => {
    render(<PageContainer width={width}>x</PageContainer>);
    const classes = screen.getByRole("main").className.split(/\s+/);
    expect(classes).toContain(token);
    expect(classes.filter((name) => name.startsWith("max-w-"))).toEqual([token]);
    expect(classes).toContain("px-gutter");
  });

  it("sem largura informada, usa a padrão", () => {
    render(<PageContainer>x</PageContainer>);
    expect(screen.getByRole("main").className.split(/\s+/)).toContain("max-w-content-default");
  });
});

describe("PageHeader", () => {
  it("renderiza exatamente um h1, com o título (DSP-01 AC2)", () => {
    render(
      <PageHeader
        title="Lançamentos"
        description="Receitas e despesas da torcida."
        breadcrumb={[{ label: "Financeiro", href: "/financeiro" }, { label: "Lançamentos" }]}
        actions={<Button>Novo lançamento</Button>}
        meta={<span>Paga</span>}
      />,
    );
    const headings = screen.getAllByRole("heading");
    expect(headings).toHaveLength(1);
    expect(headings[0].tagName).toBe("H1");
    expect(headings[0].textContent).toBe("Lançamentos");
  });

  it("mostra a descrição, as ações e o meta (DSP-01 AC2)", () => {
    render(
      <PageHeader
        title="Lançamento"
        description="Receitas e despesas da torcida."
        actions={<Button>Novo lançamento</Button>}
        meta={<span>Paga</span>}
      />,
    );
    expect(screen.getByText("Receitas e despesas da torcida.")).toBeTruthy();
    expect(screen.getByRole("button", { name: "Novo lançamento" })).toBeTruthy();
    expect(screen.getByText("Paga")).toBeTruthy();
  });

  it("marca a trilha como navegação, com link até a origem e a página atual (DSP-01 AC4)", () => {
    render(
      <PageHeader
        title="Camisa oficial"
        breadcrumb={[{ label: "Produtos", href: "/estoque/produtos" }, { label: "Camisa oficial" }]}
      />,
    );
    const trail = screen.getByRole("navigation", { name: "Trilha de navegação" });
    const links = within(trail).getAllByRole("link", { name: "Produtos" });
    expect(links.length).toBeGreaterThan(0);
    for (const link of links) expect(link.getAttribute("href")).toBe("/estoque/produtos");
    expect(within(trail).getByText("Camisa oficial").getAttribute("aria-current")).toBe("page");
  });

  it("sem trilha e sem ações, não cria navegação nem área de ações", () => {
    const { container } = render(<PageHeader title="Saldo" />);
    expect(screen.queryByRole("navigation")).toBeNull();
    expect(container.querySelector("[data-slot=page-header-actions]")).toBeNull();
  });

  it("abaixo de 768 px as ações ficam empilhadas em largura total; a partir daí, em linha (DSP-01 AC6)", () => {
    const { container } = render(
      <PageHeader
        title="Contas"
        actions={
          <>
            <Button variant="secondary">Exportar</Button>
            <Button>Nova conta</Button>
          </>
        }
      />,
    );
    const actions = container.querySelector("[data-slot=page-header-actions]") as HTMLElement;
    const classes = actions.className.split(/\s+/);
    expect(classes).toEqual(expect.arrayContaining(["flex-col", "*:w-full", "md:flex-row", "md:*:w-auto"]));
    // A primária vem por último na ordem de leitura.
    const buttons = within(actions).getAllByRole("button");
    expect(buttons.map((button) => button.textContent)).toEqual(["Exportar", "Nova conta"]);
    // Título antes das ações no documento.
    const title = screen.getByRole("heading", { level: 1 });
    expect(title.compareDocumentPosition(actions) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });
});
