import { cleanup, fireEvent, render, screen, within } from "@testing-library/react"
import { afterEach, describe, expect, it, vi } from "vitest"

import { Breadcrumb } from "@/components/ui/breadcrumb"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ShowMore } from "@/components/ui/show-more"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { DisplayTitle, Eyebrow, PageTitle, SectionTitle } from "@/components/ui/typography"
import { findRawValues } from "@/test/design/guard"

afterEach(cleanup)

function classes(el: Element) {
  return el.getAttribute("class")?.split(/\s+/) ?? []
}

function Lancamentos(props: React.ComponentProps<typeof Table>) {
  return (
    <Table aria-label="Lançamentos" {...props}>
      <TableHeader>
        <TableRow>
          <TableHead>Descrição</TableHead>
          <TableHead numeric>Valor</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>Mensalidade de março</TableCell>
          <TableCell numeric>R$ 10,00</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  )
}

describe("Table: semântica", () => {
  it("tem nome acessível e cabeçalhos de coluna com escopo", () => {
    render(<Lancamentos />)
    expect(screen.getByRole("table", { name: "Lançamentos" })).toBeTruthy()
    const headers = screen.getAllByRole("columnheader")
    expect(headers.map((header) => header.textContent)).toEqual(["Descrição", "Valor"])
    for (const header of headers) {
      expect(header.tagName).toBe("TH")
      expect(header.getAttribute("scope")).toBe("col")
    }
  })

  it("a coluna numeric alinha à direita e usa algarismos tabulares, no cabeçalho e na célula", () => {
    render(<Lancamentos />)
    for (const cell of [
      screen.getByRole("columnheader", { name: "Valor" }),
      screen.getByRole("cell", { name: "R$ 10,00" }),
    ]) {
      expect(classes(cell)).toEqual(expect.arrayContaining(["text-right", "tabular-nums"]))
    }
    for (const cell of [
      screen.getByRole("columnheader", { name: "Descrição" }),
      screen.getByRole("cell", { name: "Mensalidade de março" }),
    ]) {
      expect(classes(cell)).not.toContain("text-right")
      expect(classes(cell)).not.toContain("tabular-nums")
    }
  })

  it("o cabeçalho fica em muted, com régua de 2 px e a altura do token", () => {
    render(<Lancamentos />)
    const header = screen.getByRole("columnheader", { name: "Descrição" })
    expect(classes(header)).toEqual(
      expect.arrayContaining(["h-table-head", "text-xs", "font-semibold", "text-muted-foreground"])
    )
    expect(classes(header.closest("thead") as Element)).toEqual(
      expect.arrayContaining(["bg-muted", "[&_tr]:border-b-2", "[&_tr]:border-border-strong"])
    )
  })

  it("a linha tem a altura mínima do token, divisória e hover em accent", () => {
    render(<Lancamentos />)
    const cell = screen.getByRole("cell", { name: "Mensalidade de março" })
    expect(classes(cell)).toEqual(expect.arrayContaining(["h-table-row", "px-3", "py-2"]))
    expect(classes(cell.closest("tr") as Element)).toEqual(expect.arrayContaining(["border-b", "hover:bg-accent"]))
  })
})

describe("Table: quebra de linha por coluna", () => {
  function cell(wrap?: "nowrap" | "truncate") {
    render(
      <Table aria-label="Produtos">
        <TableBody>
          <TableRow>
            <TableCell wrap={wrap}>Camisa oficial da temporada, tamanho G</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    )
    return classes(screen.getByRole("cell"))
  }

  it("por padrão a célula quebra a linha", () => {
    const names = cell()
    expect(names).not.toContain("whitespace-nowrap")
    expect(names).not.toContain("truncate")
  })

  it("nowrap mantém o texto numa linha", () => {
    expect(cell("nowrap")).toContain("whitespace-nowrap")
  })

  it("truncate corta com reticências sem alargar a tabela", () => {
    expect(cell("truncate")).toEqual(expect.arrayContaining(["truncate", "max-w-0"]))
  })
})

describe("Table: rolagem", () => {
  it("usada como hoje, sem propriedades novas, mantém a rolagem embutida (modo legado)", () => {
    render(
      <Table>
        <TableBody>
          <TableRow>
            <TableCell>Mensalidade</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    )
    const container = screen.getByRole("table").parentElement as HTMLElement
    expect(container.getAttribute("data-slot")).toBe("table-container")
    expect(classes(container)).toContain("overflow-x-auto")
    expect(container.hasAttribute("tabindex")).toBe(false)
    expect(container.hasAttribute("role")).toBe(false)
  })

  it('scroll="none" não envolve a tabela em contêiner de rolagem', () => {
    const { container } = render(<Lancamentos scroll="none" />)
    expect(container.firstElementChild).toBe(screen.getByRole("table", { name: "Lançamentos" }))
    expect(container.querySelector("[data-slot=table-container]")).toBeNull()
  })

  it('scroll="scrollable" cria uma região rolável com nome, alcançável pelo teclado, com a primeira coluna fixa', () => {
    render(<Lancamentos scroll="scrollable" />)
    const region = screen.getByRole("region", { name: "Lançamentos" })
    expect(region.getAttribute("tabindex")).toBe("0")
    expect(classes(region)).toEqual(
      expect.arrayContaining(["overflow-x-auto", "focus-visible:ring-2", "focus-visible:ring-ring"])
    )
    region.focus()
    expect(document.activeElement).toBe(region)
    const table = within(region).getByRole("table", { name: "Lançamentos" })
    expect(classes(table)).toEqual(
      expect.arrayContaining(["[&_tr>*:first-child]:sticky", "[&_tr>*:first-child]:left-0"])
    )
  })

  it("a região rolável aceita o nome por aria-labelledby", () => {
    render(
      <>
        <h2 id="titulo-movimentacoes">Movimentações</h2>
        <Table scroll="scrollable" aria-labelledby="titulo-movimentacoes">
          <TableBody>
            <TableRow>
              <TableCell>Entrada</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </>
    )
    expect(screen.getByRole("region", { name: "Movimentações" })).toBeTruthy()
    expect(screen.getByRole("table", { name: "Movimentações" })).toBeTruthy()
  })
})

describe("Card", () => {
  it("agrupa conteúdo com título h3, borda e sem sombra", () => {
    render(
      <Card data-testid="cartao">
        <CardHeader>
          <CardTitle>Saldo em caixa</CardTitle>
        </CardHeader>
        <CardContent>R$ 198,50</CardContent>
      </Card>
    )
    expect(screen.getByRole("heading", { level: 3, name: "Saldo em caixa" })).toBeTruthy()
    const card = screen.getByTestId("cartao")
    expect(card.textContent).toBe("Saldo em caixaR$ 198,50")
    expect(classes(card)).toEqual(expect.arrayContaining(["border", "rounded-lg", "bg-card", "p-4", "md:p-6"]))
    expect(classes(card).filter((name) => /shadow/.test(name))).toEqual([])
  })
})

describe("Breadcrumb", () => {
  const ITEMS = [
    { label: "Financeiro", href: "/financeiro" },
    { label: "Lançamentos", href: "/financeiro/lancamentos" },
    { label: "Mensalidade de março" },
  ]

  it("é uma navegação nomeada com lista ordenada", () => {
    render(<Breadcrumb items={ITEMS} />)
    const nav = screen.getByRole("navigation", { name: "Trilha de navegação" })
    const list = within(nav).getByRole("list")
    expect(list.tagName).toBe("OL")
    expect(within(list).getAllByRole("listitem").map((item) => item.textContent)).toEqual([
      "Financeiro",
      "Lançamentos",
      "Mensalidade de março",
    ])
  })

  it("os itens anteriores são links e o último é a página atual, sem link", () => {
    render(<Breadcrumb items={ITEMS} />)
    const list = within(screen.getByRole("navigation")).getByRole("list")
    const links = within(list).getAllByRole("link")
    expect(links.map((link) => [link.textContent, link.getAttribute("href")])).toEqual([
      ["Financeiro", "/financeiro"],
      ["Lançamentos", "/financeiro/lancamentos"],
    ])
    const current = within(list).getByText("Mensalidade de março")
    expect(current.getAttribute("aria-current")).toBe("page")
    expect(current.closest("a")).toBeNull()
    expect(list.querySelectorAll("[aria-current]")).toHaveLength(1)
  })

  it("o último item não vira link mesmo que receba href", () => {
    render(
      <Breadcrumb
        items={[
          { label: "Produtos", href: "/estoque/produtos" },
          { label: "Camisa oficial", href: "/estoque/produtos/1" },
        ]}
      />
    )
    const current = screen.getByText("Camisa oficial")
    expect(current.getAttribute("aria-current")).toBe("page")
    expect(current.closest("a")).toBeNull()
  })

  it("os separadores são decorativos", () => {
    render(<Breadcrumb items={ITEMS} />)
    const list = within(screen.getByRole("navigation")).getByRole("list")
    const separators = list.querySelectorAll("svg")
    expect(separators).toHaveLength(2)
    for (const separator of separators) {
      expect(separator.getAttribute("aria-hidden")).toBe("true")
    }
  })

  it("abaixo de 768 px mostra só o item anterior, como voltar; a trilha completa aparece a partir daí", () => {
    render(<Breadcrumb items={ITEMS} />)
    const nav = screen.getByRole("navigation")
    const back = nav.querySelector("[data-slot=breadcrumb-back]") as HTMLAnchorElement
    expect(back.tagName).toBe("A")
    expect(back.getAttribute("href")).toBe("/financeiro/lancamentos")
    expect(back.textContent).toBe("Lançamentos")
    expect(back.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true")
    expect(classes(back)).toEqual(expect.arrayContaining(["md:hidden", "min-h-control-sm"]))
    expect(classes(within(nav).getByRole("list"))).toEqual(expect.arrayContaining(["hidden", "md:flex"]))
  })

  it("com um único item não há para onde voltar", () => {
    render(<Breadcrumb items={[{ label: "Início" }]} />)
    const nav = screen.getByRole("navigation")
    expect(nav.querySelector("[data-slot=breadcrumb-back]")).toBeNull()
    expect(within(nav).queryByRole("link")).toBeNull()
    expect(screen.getByText("Início").getAttribute("aria-current")).toBe("page")
  })
})

describe("ShowMore", () => {
  const NOUN = ["lançamento", "lançamentos"] as const

  it("informa quantos itens estão visíveis e o total", () => {
    render(<ShowMore shown={50} total={137} noun={NOUN} onShowMore={() => {}} />)
    expect(screen.getByText("Mostrando 50 de 137 lançamentos")).toBeTruthy()
  })

  it("sem total (paginação por cursor) mostra só a quantidade visível", () => {
    render(<ShowMore shown={50} noun={["usuário", "usuários"]} onShowMore={() => {}} />)
    expect(screen.getByText("Mostrando 50 usuários")).toBeTruthy()
    expect(screen.queryByText(/ de /)).toBeNull()
  })

  it("usa o singular quando a contagem é um", () => {
    render(
      <>
        <ShowMore shown={1} total={1} noun={NOUN} onShowMore={() => {}} />
        <ShowMore shown={1} noun={["usuário", "usuários"]} onShowMore={() => {}} />
      </>
    )
    expect(screen.getByText("Mostrando 1 de 1 lançamento")).toBeTruthy()
    expect(screen.getByText("Mostrando 1 usuário")).toBeTruthy()
  })

  it("formata os números em português", () => {
    render(<ShowMore shown={1050} total={2300} noun={NOUN} onShowMore={() => {}} />)
    expect(screen.getByText("Mostrando 1.050 de 2.300 lançamentos")).toBeTruthy()
  })

  it("a contagem fica numa região viva e anuncia a mudança", () => {
    const { rerender } = render(<ShowMore shown={50} total={137} noun={NOUN} onShowMore={() => {}} />)
    const live = screen.getByText("Mostrando 50 de 137 lançamentos")
    expect(live.getAttribute("aria-live")).toBe("polite")
    rerender(<ShowMore shown={100} total={137} noun={NOUN} onShowMore={() => {}} />)
    expect(live.isConnected).toBe(true)
    expect(live.textContent).toBe("Mostrando 100 de 137 lançamentos")
  })

  it("o botão secundário Mostrar mais chama onShowMore", () => {
    const onShowMore = vi.fn()
    render(<ShowMore shown={50} total={137} noun={NOUN} onShowMore={onShowMore} />)
    const button = screen.getByRole("button", { name: "Mostrar mais" })
    expect(classes(button)).toContain("border-input")
    button.focus()
    expect(document.activeElement).toBe(button)
    fireEvent.click(button)
    expect(onShowMore).toHaveBeenCalledTimes(1)
  })

  it("com tudo visível não oferece o botão, mas mantém a contagem", () => {
    render(<ShowMore shown={137} total={137} noun={NOUN} onShowMore={() => {}} />)
    expect(screen.getByText("Mostrando 137 de 137 lançamentos")).toBeTruthy()
    expect(screen.queryByRole("button")).toBeNull()
  })

  it("em andamento o botão fica desabilitado, com aria-busy e rótulo de progresso", () => {
    const onShowMore = vi.fn()
    render(<ShowMore shown={50} noun={NOUN} onShowMore={onShowMore} loading />)
    const button = screen.getByRole("button", { name: "Carregando…" }) as HTMLButtonElement
    expect(button.disabled).toBe(true)
    expect(button.getAttribute("aria-busy")).toBe("true")
    fireEvent.click(button)
    expect(onShowMore).not.toHaveBeenCalled()
  })
})

describe("Tipografia", () => {
  it("PageTitle é o h1, em caixa alta, com o filete de marca", () => {
    render(<PageTitle>Lançamentos</PageTitle>)
    const title = screen.getByRole("heading", { level: 1, name: "Lançamentos" })
    expect(classes(title)).toEqual(
      expect.arrayContaining([
        "font-display",
        "font-bold",
        "uppercase",
        "text-2xl",
        "md:text-3xl",
        "text-balance",
        "after:bg-brand",
      ])
    )
  })

  it("SectionTitle é o h2, em frase comum, com a régua forte", () => {
    render(<SectionTitle>Comprovantes</SectionTitle>)
    const title = screen.getByRole("heading", { level: 2, name: "Comprovantes" })
    expect(classes(title)).toEqual(
      expect.arrayContaining([
        "font-display",
        "font-semibold",
        "text-xl",
        "text-balance",
        "border-b-2",
        "border-border-strong",
      ])
    )
    expect(classes(title)).not.toContain("uppercase")
  })

  it("DisplayTitle é o destaque em caixa alta e não cria um título na estrutura da página", () => {
    render(<DisplayTitle>Torcida Jovem</DisplayTitle>)
    const title = screen.getByText("Torcida Jovem")
    expect(classes(title)).toEqual(
      expect.arrayContaining(["font-display", "font-bold", "uppercase", "text-4xl", "text-balance"])
    )
    expect(screen.queryByRole("heading")).toBeNull()
  })

  it("Eyebrow é o rótulo de grupo em caixa alta e não cria um título", () => {
    render(<Eyebrow>Financeiro</Eyebrow>)
    const eyebrow = screen.getByText("Financeiro")
    expect(classes(eyebrow)).toEqual(
      expect.arrayContaining(["font-display", "font-semibold", "uppercase", "text-sm", "text-balance"])
    )
    expect(screen.queryByRole("heading")).toBeNull()
  })

  it("os componentes aceitam className e atributos de quem chama", () => {
    render(
      <PageTitle id="titulo" className="text-sidebar-foreground">
        Entrar
      </PageTitle>
    )
    const title = screen.getByRole("heading", { level: 1 })
    expect(title.id).toBe("titulo")
    expect(classes(title)).toContain("text-sidebar-foreground")
  })

  it("são o único lugar de components/ui com font-display e uppercase", () => {
    const hits = findRawValues(["components/ui"]).filter(
      (hit) => (hit.rule === "font-display" || hit.rule === "uppercase") && !/\.test\.tsx?$/.test(hit.file)
    )
    expect(new Set(hits.map((hit) => hit.rule))).toEqual(new Set(["font-display", "uppercase"]))
    expect([...new Set(hits.map((hit) => hit.file))]).toEqual(["components/ui/typography.tsx"])
  })
})
