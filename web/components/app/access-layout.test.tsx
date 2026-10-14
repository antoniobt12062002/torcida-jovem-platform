import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AccessLayout } from "./access-layout";

function classes(el: Element) {
  return el.getAttribute("class")?.split(/\s+/) ?? [];
}

function renderLayout() {
  return render(
    <AccessLayout>
      <h1>Entrar</h1>
      <input aria-label="E-mail" />
    </AccessLayout>,
  );
}

const panel = (container: HTMLElement) => container.querySelector("[data-slot=access-brand]") as HTMLElement;

// DSS-03 AC1 e AC5; WDS-D-020 e WDS-D-027.
describe("marca", () => {
  it("mostra o escudo oficial e o nome Torcida Jovem no painel de marca", () => {
    const { container } = renderLayout();
    const brand = panel(container);
    const images = [...brand.querySelectorAll("img")];
    expect(images.length).toBeGreaterThan(0);
    for (const image of images) expect(image.getAttribute("src")).toBe("/brand/torcida-jovem-logo.svg");
    expect(within(brand).getAllByText("Torcida Jovem").length).toBeGreaterThan(0);
    expect(classes(brand)).toEqual(expect.arrayContaining(["bg-sidebar", "text-sidebar-foreground"]));
  });

  it("escudo de 64 px na faixa do topo e de 96 px no painel lateral, trocados só por CSS", () => {
    const { container } = renderLayout();
    const marks = [...panel(container).querySelectorAll("[data-slot=brand-mark]")];
    expect(marks.map((mark) => mark.querySelector("img")?.getAttribute("width"))).toEqual(["64", "96"]);
    expect(classes(marks[0])).toContain("lg:hidden");
    expect(classes(marks[1])).toEqual(expect.arrayContaining(["hidden", "lg:inline-flex"]));
    // O nome fica abaixo do escudo.
    for (const mark of marks) expect(classes(mark)).toContain("flex-col");
  });

  it("não repete Desde 2013 nem exibe TJ Platform", () => {
    const { container } = renderLayout();
    expect(container.textContent).not.toContain("Desde 2013");
    expect(container.textContent).not.toContain("2013");
    expect(container.textContent).not.toContain("TJ Platform");
  });

  it("não usa ilustração nem fotografia: as únicas imagens são o escudo", () => {
    const { container } = renderLayout();
    const images = [...container.querySelectorAll("img, svg, picture, video")];
    expect(images.every((image) => image.getAttribute("src") === "/brand/torcida-jovem-logo.svg")).toBe(true);
    expect(container.innerHTML).not.toContain("gradient");
  });
});

// DSS-03 AC2 e AC6.
describe("layout", () => {
  it("o conteúdo fica na única região principal, que é o alvo do link de pular", () => {
    renderLayout();
    const main = screen.getByRole("main");
    expect(screen.getAllByRole("main")).toHaveLength(1);
    expect(main.id).toBe("conteudo");
    expect(main.getAttribute("tabindex")).toBe("-1");
    expect(within(main).getByRole("heading", { level: 1, name: "Entrar" })).toBeTruthy();
    expect(within(main).getByRole("textbox", { name: "E-mail" })).toBeTruthy();
    expect(main.textContent).not.toContain("Torcida Jovem");
  });

  it("coluna única abaixo de 1024 px e duas colunas a partir daí, só por CSS", () => {
    const { container } = renderLayout();
    const root = container.firstElementChild as HTMLElement;
    expect(classes(root)).toEqual(expect.arrayContaining(["flex", "min-h-svh", "flex-col", "lg:flex-row"]));
    expect(classes(panel(container))).toContain("lg:w-2/5");
    // A marca vem antes do formulário.
    expect(panel(container).compareDocumentPosition(screen.getByRole("main")) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it("o formulário ocupa a largura disponível até o máximo estreito de 400 px", () => {
    renderLayout();
    const column = screen.getByRole("heading", { name: "Entrar" }).parentElement as HTMLElement;
    expect(classes(column)).toEqual(expect.arrayContaining(["w-full", "max-w-100"]));
    expect(classes(screen.getByRole("main"))).toContain("px-4");
  });

  it("aceita uma ação no canto, fora do formulário", () => {
    render(
      <AccessLayout action={<button type="button">Sair</button>}>
        <h1>Trocar senha</h1>
      </AccessLayout>,
    );
    const action = screen.getByRole("button", { name: "Sair" });
    expect(screen.getByRole("main").contains(action)).toBe(true);
    expect(screen.getByRole("heading", { name: "Trocar senha" }).parentElement?.contains(action)).toBe(false);
  });
});
