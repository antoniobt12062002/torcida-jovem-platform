import { fireEvent, render, screen, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const notFound = vi.hoisted(() =>
  vi.fn(() => {
    throw new Error("NEXT_HTTP_ERROR_FALLBACK;404");
  }),
);

vi.mock("next/navigation", async () => ({
  ...(await import("@/lib/session/navigation-mock")).nextNavigationMock,
  notFound,
}));

import { resetNavigation } from "@/lib/session/navigation-mock";

import AccessPage, { generateMetadata as accessMetadata } from "./acesso/page";
import ExamplePage, { generateMetadata as exampleMetadata } from "./pagina-de-exemplo/page";
import NoAccessPage, { generateMetadata as noAccessMetadata } from "./sem-acesso/page";

// Páginas próprias do espécime: as composições que trazem o próprio `main` ou
// `h1` (AccessLayout, PageHeader, Sem acesso). Cada uma também some em produção.

beforeEach(() => resetNavigation());
afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
  notFound.mockClear();
});

const PAGES = [
  { name: "acesso", Page: AccessPage, metadata: accessMetadata, h1: "Entrar" },
  { name: "pagina-de-exemplo", Page: ExamplePage, metadata: exampleMetadata, h1: "Lançamentos" },
  { name: "sem-acesso", Page: NoAccessPage, metadata: noAccessMetadata, h1: "Sem acesso" },
] as const;

describe.each(PAGES)("/design-system/$name", ({ Page, metadata, h1 }) => {
  it("com NODE_ENV=production chama notFound na página e nos metadados", () => {
    vi.stubEnv("NODE_ENV", "production");
    expect(() => Page()).toThrow("NEXT_HTTP_ERROR_FALLBACK;404");
    expect(() => metadata()).toThrow("NEXT_HTTP_ERROR_FALLBACK;404");
    expect(notFound).toHaveBeenCalledTimes(2);
  });

  it("fora de produção tem um só main, um só h1, nenhum id repetido e nenhuma requisição", () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const { container } = render(<Page />);
    expect(notFound).not.toHaveBeenCalled();
    expect(container.querySelectorAll("main")).toHaveLength(1);
    expect(screen.getAllByRole("heading", { level: 1 }).map((heading) => heading.textContent)).toEqual([h1]);
    const ids = [...container.querySelectorAll("[id]")].map((element) => element.id);
    expect(ids.filter((id, index) => ids.indexOf(id) !== index)).toEqual([]);
    expect(fetchSpy).not.toHaveBeenCalled();
  });
});

describe("tela de acesso de exemplo", () => {
  it("usa o AccessLayout com o escudo e valida o formulário sem enviar nada", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const { container } = render(<AccessPage />);
    expect(container.querySelector("[data-slot=access-layout]")).not.toBeNull();
    expect(screen.getAllByText("Torcida Jovem").length).toBeGreaterThanOrEqual(1);
    fireEvent.click(screen.getByRole("button", { name: "Entrar" }));
    expect(await screen.findByText("Informe o e-mail.")).toBeTruthy();
    expect(screen.getByText("Informe a senha.")).toBeTruthy();
    fireEvent.change(screen.getByRole("textbox", { name: "E-mail" }), { target: { value: "pessoa@exemplo.com" } });
    fireEvent.change(screen.getByLabelText("Senha"), { target: { value: "exemplo" } });
    fireEvent.click(screen.getByRole("button", { name: "Entrar" }));
    expect(screen.queryByText("Informe o e-mail.")).toBeNull();
    expect(fetchSpy).not.toHaveBeenCalled();
  });
});

describe("página de exemplo completa", () => {
  it("junta shell, trilha, cabeçalho de página, barra de filtros e listagem", () => {
    render(<ExamplePage />);
    expect(screen.getByRole("navigation", { name: "Navegação principal" })).toBeTruthy();
    expect(screen.getByRole("navigation", { name: "Trilha de navegação" })).toBeTruthy();
    expect(screen.getByRole("group", { name: "Filtros" })).toBeTruthy();
    const table = screen.getByRole("table", { name: "Lançamentos" });
    expect(within(table).getAllByRole("row")).toHaveLength(13);
    expect(screen.getByRole("main").querySelector("[data-slot=page-header]")).not.toBeNull();
  });
});
