import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { HouseIcon, ReceiptIcon } from "lucide-react";
import type { ReactNode } from "react";
import { renderToString } from "react-dom/server";
import { describe, expect, it } from "vitest";

import { DropdownMenuItem, DropdownMenuLinkItem } from "@/components/ui/dropdown-menu";
import { mockViewport } from "@/test/media";

import type { NavEntry } from "./nav";
import { PageContainer } from "./page-container";
import { ShellFrame, type ShellFrameProps } from "./shell-frame";

// O ShellFrame é só apresentação: estes testes não montam sessão, API nem
// roteador (WDS-D-017: o espécime usa o shell real com dados de exemplo).

const entries: NavEntry[] = [
  { label: "Início", href: "/inicio", icon: HouseIcon },
  {
    label: "Financeiro",
    items: [
      { label: "Lançamentos", href: "/financeiro/lancamentos", icon: ReceiptIcon },
      { label: "Contas", href: "/financeiro/contas" },
    ],
  },
];

const userMenu = (
  <>
    <DropdownMenuLinkItem href="/conta/senha">Trocar senha</DropdownMenuLinkItem>
    <DropdownMenuItem>Sair</DropdownMenuItem>
  </>
);

function Shell({ children, ...props }: Partial<ShellFrameProps> & { children?: ReactNode }) {
  return (
    <ShellFrame
      entries={entries}
      currentPath="/inicio"
      userName="Maria Aparecida da Silva Sauro"
      userEmail="maria@example.com"
      userMenu={userMenu}
      showNav
      {...props}
    >
      {children ?? <PageContainer>conteúdo da página</PageContainer>}
    </ShellFrame>
  );
}

function classes(el: Element) {
  return el.getAttribute("class")?.split(/\s+/) ?? [];
}

function focusables(container: HTMLElement) {
  return [...container.querySelectorAll<HTMLElement>("a[href], button, input, select, textarea, [tabindex]")].filter(
    (el) => el.tabIndex >= 0,
  );
}

function linkNames(container: HTMLElement) {
  return within(container)
    .queryAllByRole("link")
    .map((link) => [link.textContent, link.getAttribute("href")]);
}

const nav = () => screen.getByRole("navigation", { name: "Navegação principal" });
const menuButton = () => screen.getByRole("button", { name: "Abrir menu" });

function pressEscape() {
  fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });
}

function clickOutside(target: Element) {
  fireEvent.pointerDown(target);
  fireEvent.mouseDown(target);
  fireEvent.pointerUp(target);
  fireEvent.mouseUp(target);
  fireEvent.click(target);
}

// DSS-01 AC5.
describe("link de pular para o conteúdo", () => {
  it("é o primeiro elemento focável e aponta para #conteudo", () => {
    const { container } = render(<Shell />);
    const [first] = focusables(container);
    expect(first.tagName).toBe("A");
    expect(first.textContent).toBe("Pular para o conteúdo");
    expect(first.getAttribute("href")).toBe("#conteudo");
    // Antes do cabeçalho, no documento.
    const header = screen.getByRole("banner");
    expect(first.compareDocumentPosition(header) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it("fica fora da tela até receber foco", () => {
    render(<Shell />);
    expect(classes(screen.getByRole("link", { name: "Pular para o conteúdo" }))).toEqual(
      expect.arrayContaining(["fixed", "-translate-y-full", "focus:translate-y-0"]),
    );
  });

  it("o anel de foco usa o token do shell, visível sobre o cabeçalho escuro", () => {
    render(<Shell />);
    expect(classes(screen.getByRole("link", { name: "Pular para o conteúdo" }))).toEqual(
      expect.arrayContaining(["focus-visible:ring-2", "focus-visible:ring-sidebar-ring"]),
    );
  });

  it("leva o foco a main#conteudo", () => {
    render(<Shell />);
    fireEvent.click(screen.getByRole("link", { name: "Pular para o conteúdo" }));
    const main = screen.getByRole("main");
    expect(main.id).toBe("conteudo");
    expect(document.activeElement).toBe(main);
  });

  it("numa página ainda sem PageContainer, leva o foco ao main da página, sem id duplicado", () => {
    render(
      <Shell>
        <main>página antiga</main>
      </Shell>,
    );
    fireEvent.click(screen.getByRole("link", { name: "Pular para o conteúdo" }));
    const main = screen.getByRole("main");
    expect(document.activeElement).toBe(main);
    expect(main.textContent).toBe("página antiga");
    expect(document.querySelectorAll("#conteudo").length).toBeLessThanOrEqual(1);
  });

  it("sem main na página, leva o foco à área de conteúdo", () => {
    const { container } = render(
      <Shell>
        <p>só um parágrafo</p>
      </Shell>,
    );
    fireEvent.click(screen.getByRole("link", { name: "Pular para o conteúdo" }));
    const content = container.querySelector("[data-slot=shell-content]");
    expect(document.activeElement).toBe(content);
    expect(content?.textContent).toBe("só um parágrafo");
  });
});

// DSS-01 AC6.
describe("regiões", () => {
  it("o shell não cria main: a página tem uma única região principal, sem id duplicado", () => {
    render(<Shell />);
    expect(screen.getAllByRole("main")).toHaveLength(1);
    expect(document.querySelectorAll("#conteudo")).toHaveLength(1);
  });

  it("tem um cabeçalho e uma navegação nomeada; a gaveta não existe fechada", () => {
    render(<Shell />);
    expect(screen.getAllByRole("banner")).toHaveLength(1);
    expect(screen.getAllByRole("navigation")).toHaveLength(1);
    expect(nav()).toBeTruthy();
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(screen.queryByRole("navigation", { name: "Menu" })).toBeNull();
  });
});

// DSS-01 AC1 e AC2; WDS-D-020.
describe("identidade", () => {
  it("a marca é o escudo com o nome Torcida Jovem, com link para o início", () => {
    render(<Shell />);
    const brand = within(screen.getByRole("banner")).getByRole("link", { name: "Torcida Jovem" });
    expect(brand.getAttribute("href")).toBe("/inicio");
    const image = within(brand).getByRole("img", { name: "Torcida Jovem" });
    expect(image.getAttribute("src")).toBe("/brand/torcida-jovem-logo.svg");
    expect(image.getAttribute("width")).toBe("32");
    expect(within(brand).getByText("Torcida Jovem")).toBeTruthy();
  });

  it("o link da marca tem alvo de 44 px em toque e o escudo carrega com prioridade", () => {
    render(<Shell />);
    const brand = within(screen.getByRole("banner")).getByRole("link", { name: "Torcida Jovem" });
    expect(classes(brand)).toEqual(expect.arrayContaining(["min-h-touch", "min-w-touch", "items-center"]));
    expect(within(brand).getByRole("img").getAttribute("loading")).toBe("eager");
  });

  it("o nome ao lado do escudo some abaixo de 400 px e não é lido duas vezes", () => {
    render(<Shell />);
    const name = within(screen.getByRole("banner")).getByText("Torcida Jovem");
    expect(name.getAttribute("aria-hidden")).toBe("true");
    expect(classes(name)).toEqual(expect.arrayContaining(["hidden", "min-[400px]:block"]));
  });

  it("não exibe TJ Platform", () => {
    const { container } = render(<Shell />);
    expect(container.textContent).not.toContain("TJ Platform");
  });

  it("cabeçalho e navegação usam os tokens do shell escuro; o conteúdo, os do conteúdo claro", () => {
    const { container } = render(<Shell />);
    const header = screen.getByRole("banner");
    expect(classes(header)).toEqual(
      expect.arrayContaining(["sticky", "top-0", "border-sidebar-border", "bg-sidebar", "text-sidebar-foreground"]),
    );
    expect(header.querySelector(".h-shell-header")).not.toBeNull();
    const rail = container.querySelector("[data-slot=shell-nav]") as HTMLElement;
    expect(rail.contains(nav())).toBe(true);
    expect(classes(rail)).toEqual(
      expect.arrayContaining(["w-shell-nav", "border-sidebar-border", "bg-sidebar", "text-sidebar-foreground"]),
    );
    expect(classes(container.querySelector("[data-slot=shell-content]") as Element)).toContain("bg-background");
  });

  // DSS-02 AC6.
  // O jsdom descarta `env()` do atributo `style`; o HTML do servidor o preserva.
  it("o cabeçalho respeita as áreas seguras do aparelho", () => {
    const header = /<header[^>]*>/.exec(renderToString(<Shell />))?.[0] ?? "";
    expect(header).toContain("padding-top:env(safe-area-inset-top)");
    expect(header).toContain("padding-left:env(safe-area-inset-left)");
    expect(header).toContain("padding-right:env(safe-area-inset-right)");
  });
});

// DSS-01 AC3 e AC4.
describe("navegação", () => {
  it("agrupa os itens por módulo, com rótulo de grupo, na ordem recebida", () => {
    render(<Shell />);
    expect(linkNames(nav())).toEqual([
      ["Início", "/inicio"],
      ["Lançamentos", "/financeiro/lancamentos"],
      ["Contas", "/financeiro/contas"],
    ]);
    const group = within(nav()).getByText("Financeiro");
    expect(group.getAttribute("data-slot")).toBe("eyebrow");
    expect(classes(group)).toContain("text-sidebar-muted-foreground");
  });

  it("a lista de cada grupo é nomeada pelo rótulo do grupo", () => {
    render(<Shell />);
    const list = within(nav()).getByRole("list", { name: "Financeiro" });
    const label = document.getElementById(list.getAttribute("aria-labelledby") ?? "");
    expect(label?.getAttribute("data-slot")).toBe("eyebrow");
    expect(within(list).getAllByRole("link").map((link) => link.textContent)).toEqual(["Lançamentos", "Contas"]);
  });

  it("na gaveta, a lista do grupo também é nomeada, sem repetir os ids da navegação fixa", async () => {
    render(<Shell />);
    // Com a gaveta aberta (modal), o resto da página sai da árvore acessível.
    const railLabel = within(nav()).getByRole("list", { name: "Financeiro" }).getAttribute("aria-labelledby");
    fireEvent.click(menuButton());
    const drawer = await screen.findByRole("navigation", { name: "Menu" });
    const drawerList = within(drawer).getByRole("list", { name: "Financeiro" });
    expect(drawerList.getAttribute("aria-labelledby")).not.toBe(railLabel);
    expect(document.querySelectorAll(`[id="${drawerList.getAttribute("aria-labelledby")}"]`)).toHaveLength(1);
  });

  it("cada item tem ícone decorativo e texto; item sem ícone renderiza só o texto", () => {
    render(<Shell />);
    const withIcon = within(nav()).getByRole("link", { name: "Lançamentos" });
    const icon = withIcon.querySelector("svg");
    expect(icon?.getAttribute("aria-hidden")).toBe("true");
    expect(classes(icon as Element)).toContain("size-5");
    const withoutIcon = within(nav()).getByRole("link", { name: "Contas" });
    expect(withoutIcon.querySelector("svg")).toBeNull();
    expect(withoutIcon.textContent).toBe("Contas");
  });

  it.each([
    ["/financeiro/lancamentos", "a própria página"],
    ["/financeiro/lancamentos/123", "uma página abaixo dela"],
  ])("em %s (%s), marca Lançamentos com aria-current, peso maior e barra vermelha", (currentPath) => {
    render(<Shell currentPath={currentPath} />);
    const active = within(nav()).getByRole("link", { name: "Lançamentos" });
    expect(active.getAttribute("aria-current")).toBe("page");
    expect(classes(active)).toEqual(
      expect.arrayContaining(["font-semibold", "before:w-0.75", "before:bg-sidebar-primary"]),
    );
    // O indicador fica sempre sobre `sidebar`: o item ativo não troca de fundo.
    expect(classes(active).filter((c) => c.includes("bg-sidebar-accent"))).toEqual([]);

    const inactive = within(nav()).getByRole("link", { name: "Contas" });
    expect(inactive.getAttribute("aria-current")).toBeNull();
    expect(classes(inactive)).toEqual(expect.arrayContaining(["font-normal", "hover:bg-sidebar-accent"]));
    expect(classes(inactive).filter((c) => c.startsWith("before:"))).toEqual([]);
    expect(within(nav()).getAllByRole("link").filter((a) => a.getAttribute("aria-current"))).toHaveLength(1);
  });

  it("um caminho que só começa igual não marca o item", () => {
    render(<Shell currentPath="/financeiro/contas-antigas" />);
    expect(within(nav()).getByRole("link", { name: "Contas" }).getAttribute("aria-current")).toBeNull();
  });

  it("os itens têm o texto do shell, a altura de controle e o foco do shell", () => {
    render(<Shell />);
    for (const link of within(nav()).getAllByRole("link")) {
      expect(classes(link)).toEqual(
        expect.arrayContaining([
          "min-h-control-md",
          "text-sidebar-foreground",
          "focus-visible:ring-2",
          "focus-visible:ring-sidebar-ring",
        ]),
      );
    }
  });
});

// DSS-02 AC1 a AC4; WDS-D-019.
describe("navegação responsiva", () => {
  it("a navegação fixa aparece a partir de 1024 px (lg), só por CSS", () => {
    const { container } = render(<Shell />);
    const rail = classes(container.querySelector("[data-slot=shell-nav]") as Element);
    expect(rail).toEqual(expect.arrayContaining(["hidden", "lg:block"]));
    expect(rail.filter((c) => c.startsWith("md:") || c.startsWith("sm:"))).toEqual([]);
  });

  it("abaixo de 1024 px há o botão Abrir menu, com alvo de 44 px", () => {
    render(<Shell />);
    const button = classes(menuButton());
    expect(button).toEqual(expect.arrayContaining(["lg:hidden", "size-control-lg"]));
    expect(button.filter((c) => c.startsWith("md:") || c.startsWith("sm:"))).toEqual([]);
  });

  it("o botão abre a navegação numa gaveta à esquerda, com os mesmos itens e nome distinto", async () => {
    render(<Shell currentPath="/financeiro/contas" />);
    fireEvent.click(menuButton());
    const sheet = await screen.findByRole("dialog", { name: "Menu" });
    expect(sheet.getAttribute("data-side")).toBe("left");
    expect(classes(sheet)).toEqual(expect.arrayContaining(["bg-sidebar", "text-sidebar-foreground"]));
    const drawerNav = within(sheet).getByRole("navigation", { name: "Menu" });
    expect(linkNames(drawerNav)).toEqual([
      ["Início", "/inicio"],
      ["Lançamentos", "/financeiro/lancamentos"],
      ["Contas", "/financeiro/contas"],
    ]);
    expect(within(drawerNav).getByRole("link", { name: "Contas" }).getAttribute("aria-current")).toBe("page");
  });

  it("escolher um item fecha a gaveta", async () => {
    render(<Shell />);
    fireEvent.click(menuButton());
    const sheet = await screen.findByRole("dialog", { name: "Menu" });
    const link = within(sheet).getByRole("link", { name: "Contas" });
    link.addEventListener("click", (event) => event.preventDefault());
    fireEvent.click(link);
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("Esc fecha a gaveta e devolve o foco ao botão", async () => {
    render(<Shell />);
    const trigger = menuButton();
    trigger.focus();
    fireEvent.click(trigger);
    const sheet = await screen.findByRole("dialog", { name: "Menu" });
    await waitFor(() => expect(sheet.contains(document.activeElement)).toBe(true));
    pressEscape();
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
    await waitFor(() => expect(document.activeElement).toBe(trigger));
  });

  it("o botão de fechar fecha a gaveta e devolve o foco ao botão", async () => {
    render(<Shell />);
    const trigger = menuButton();
    trigger.focus();
    fireEvent.click(trigger);
    const sheet = await screen.findByRole("dialog", { name: "Menu" });
    fireEvent.click(within(sheet).getByRole("button", { name: "Fechar" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
    await waitFor(() => expect(document.activeElement).toBe(trigger));
  });

  it("clicar fora fecha a gaveta", async () => {
    render(<Shell />);
    fireEvent.click(menuButton());
    await screen.findByRole("dialog", { name: "Menu" });
    clickOutside(document.querySelector("[data-slot=sheet-overlay]") as Element);
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("se a janela passa a 1024 px com a gaveta aberta, ela fecha", async () => {
    const viewport = mockViewport({ width: 820 });
    render(<Shell />);
    fireEvent.click(menuButton());
    await screen.findByRole("dialog", { name: "Menu" });

    viewport.resize(1023);
    expect(screen.getByRole("dialog", { name: "Menu" })).toBeTruthy();

    viewport.resize(1024);
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());

    // Voltar a estreitar não reabre a gaveta.
    viewport.resize(820);
    expect(screen.queryByRole("dialog")).toBeNull();
  });
});

// DSS-01 AC7.
describe("menu da pessoa", () => {
  it("o gatilho mostra o nome e usa as cores do shell", () => {
    render(<Shell />);
    const trigger = screen.getByRole("button", { name: "Maria Aparecida da Silva Sauro" });
    expect(within(screen.getByRole("banner")).getByRole("button", { name: /Maria Aparecida/ })).toBe(trigger);
    expect(classes(trigger)).toEqual(
      expect.arrayContaining(["text-sidebar-foreground", "hover:bg-sidebar-accent", "focus-visible:ring-sidebar-ring"]),
    );
    expect(classes(within(trigger).getByText("Maria Aparecida da Silva Sauro"))).toContain("truncate");
  });

  it("aberto, mostra o nome completo e o e-mail como texto, antes dos itens recebidos", async () => {
    render(<Shell />);
    fireEvent.click(screen.getByRole("button", { name: "Maria Aparecida da Silva Sauro" }));
    const menu = await screen.findByRole("menu");
    const name = within(menu).getByText("Maria Aparecida da Silva Sauro");
    expect(within(menu).getByText("maria@example.com")).toBeTruthy();
    const items = within(menu).getAllByRole("menuitem");
    expect(items.map((item) => item.textContent)).toEqual(["Trocar senha", "Sair"]);
    expect(name.closest("[role=menuitem]")).toBeNull();
    expect(name.compareDocumentPosition(items[0]) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it("nome e e-mail são o rótulo de um grupo do menu, não um bloco solto", async () => {
    render(<Shell />);
    fireEvent.click(screen.getByRole("button", { name: "Maria Aparecida da Silva Sauro" }));
    const menu = await screen.findByRole("menu");
    const group = within(menu).getByRole("group");
    const label = document.getElementById(group.getAttribute("aria-labelledby") ?? "");
    expect(label?.getAttribute("data-slot")).toBe("dropdown-menu-label");
    expect(label?.textContent).toContain("Maria Aparecida da Silva Sauro");
    expect(label?.textContent).toContain("maria@example.com");
    expect(within(group).getAllByRole("menuitem").map((item) => item.textContent)).toEqual(["Trocar senha", "Sair"]);
    for (const child of menu.children) {
      expect(["group", "separator", "menuitem"], child.outerHTML).toContain(child.getAttribute("role"));
    }
  });

  it("sem e-mail, mostra só o nome", async () => {
    render(<Shell userEmail={undefined} />);
    fireEvent.click(screen.getByRole("button", { name: "Maria Aparecida da Silva Sauro" }));
    const menu = await screen.findByRole("menu");
    expect(within(menu).getByText("Maria Aparecida da Silva Sauro")).toBeTruthy();
    expect(within(menu).queryByText("maria@example.com")).toBeNull();
  });
});

// FND-05 AC8, na parte de apresentação.
describe("sem navegação (troca obrigatória de senha)", () => {
  it("não mostra navegação, botão de menu nem link; o menu da pessoa continua", () => {
    render(<Shell showNav={false} entries={[]} />);
    expect(screen.queryByRole("navigation")).toBeNull();
    expect(screen.queryByRole("button", { name: "Abrir menu" })).toBeNull();
    expect(screen.queryAllByRole("link")).toEqual([]);
    expect(within(screen.getByRole("banner")).getByRole("img", { name: "Torcida Jovem" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Maria Aparecida da Silva Sauro" })).toBeTruthy();
    expect(screen.getByText("conteúdo da página")).toBeTruthy();
  });
});
