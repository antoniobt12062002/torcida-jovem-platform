import { cleanup, fireEvent, render, screen } from "@testing-library/react"
import { PlusIcon, Trash2Icon } from "lucide-react"
import { afterEach, describe, expect, it, vi } from "vitest"

import { Button, buttonVariants } from "@/components/ui/button"
import { IconButton } from "@/components/ui/icon-button"

afterEach(cleanup)

const VARIANTS = ["primary", "secondary", "ghost", "destructive", "link"] as const
const SIZES = [
  ["sm", "h-control-sm"],
  ["md", "h-control-md"],
  ["lg", "h-control-lg"],
] as const

function classes(el: HTMLElement) {
  return el.className.split(/\s+/)
}

describe("Button: variantes", () => {
  it.each(VARIANTS)("%s tem papel de botão, nome e aceita ação", (variant) => {
    const onClick = vi.fn()
    render(
      <Button variant={variant} onClick={onClick}>
        Salvar
      </Button>
    )
    const button = screen.getByRole("button", { name: "Salvar" }) as HTMLButtonElement
    expect(button.tagName).toBe("BUTTON")
    expect(button.type).toBe("button")
    expect(button.disabled).toBe(false)
    button.focus()
    expect(document.activeElement).toBe(button)
    fireEvent.click(button)
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it.each(VARIANTS)("%s desabilitada não dispara a ação", (variant) => {
    const onClick = vi.fn()
    render(
      <Button variant={variant} disabled onClick={onClick}>
        Salvar
      </Button>
    )
    const button = screen.getByRole("button", { name: "Salvar" }) as HTMLButtonElement
    expect(button.disabled).toBe(true)
    fireEvent.click(button)
    expect(onClick).not.toHaveBeenCalled()
  })

  it("sem variante é a primária, de fundo preto", () => {
    render(<Button>Salvar</Button>)
    const button = screen.getByRole("button", { name: "Salvar" })
    expect(classes(button)).toContain("bg-primary")
    expect(classes(button)).toContain("text-primary-foreground")
  })

  it("a destrutiva é sólida no tom de erro, com texto de contraste", () => {
    render(<Button variant="destructive">Cancelar lançamento</Button>)
    const button = screen.getByRole("button", { name: "Cancelar lançamento" })
    expect(classes(button)).toContain("bg-destructive")
    expect(classes(button)).toContain("text-destructive-foreground")
  })

  it("a secundária tem contorno e a discreta não tem fundo", () => {
    expect(buttonVariants({ variant: "secondary" }).split(/\s+/)).toEqual(
      expect.arrayContaining(["border-input", "bg-background"])
    )
    const ghost = buttonVariants({ variant: "ghost" }).split(/\s+/)
    expect(ghost.filter((name) => /^bg-/.test(name))).toEqual([])
  })

  it("a de link é texto sublinhado", () => {
    expect(buttonVariants({ variant: "link" }).split(/\s+/)).toContain("underline")
  })
})

// Revisão da rodada 4: o hover do link não reduz o contraste do texto.
describe("Button: hover do link", () => {
  it("engrossa o sublinhado e não clareia o texto", () => {
    const link = buttonVariants({ variant: "link" }).split(/\s+/)
    expect(link).toContain("hover:decoration-2")
    expect(link.filter((name) => name.startsWith("hover:text-") && name !== "hover:text-foreground")).toEqual([])
  })
})

describe("Button: tamanhos", () => {
  it.each(SIZES)("%s usa o token de altura %s", (size, token) => {
    render(<Button size={size}>Salvar</Button>)
    const button = screen.getByRole("button", { name: "Salvar" }) as HTMLButtonElement
    expect(classes(button)).toContain(token)
    expect(button.disabled).toBe(false)
  })

  it.each(SIZES)("%s desabilitado continua com nome e fica desabilitado", (size) => {
    render(
      <Button size={size} disabled>
        Salvar
      </Button>
    )
    expect((screen.getByRole("button", { name: "Salvar" }) as HTMLButtonElement).disabled).toBe(true)
  })
})

describe("Button: estados comuns", () => {
  it.each(VARIANTS)("%s tem foco por anel de 2 px afastado, nunca vermelho", (variant) => {
    const names = buttonVariants({ variant }).split(/\s+/)
    expect(names).toEqual(
      expect.arrayContaining([
        "focus-visible:ring-2",
        "focus-visible:ring-ring",
        "focus-visible:ring-offset-2",
        "focus-visible:ring-offset-ring-offset",
      ])
    )
    expect(names.filter((name) => /focus-visible:.*(destructive|brand)/.test(name))).toEqual([])
  })

  it("anima só cor, sem deslocamento ao pressionar, e evita o atraso de toque", () => {
    const names = buttonVariants().split(/\s+/)
    expect(names).toContain("transition-colors")
    expect(names).toContain("touch-manipulation")
    expect(names.filter((name) => /translate/.test(name))).toEqual([])
  })

  it("o ícone ao lado do texto fica oculto de tecnologias assistivas", () => {
    render(
      <Button>
        <PlusIcon data-testid="icone" />
        Novo lançamento
      </Button>
    )
    expect(screen.getByRole("button", { name: "Novo lançamento" })).toBeTruthy()
    expect(screen.getByTestId("icone").getAttribute("aria-hidden")).toBe("true")
  })
})

describe("Button: em andamento", () => {
  it("desabilita, expõe aria-busy e troca o rótulo pelo de progresso", () => {
    const onClick = vi.fn()
    render(
      <Button loading loadingLabel="Salvando…" onClick={onClick}>
        Salvar
      </Button>
    )
    const button = screen.getByRole("button", { name: "Salvando…" }) as HTMLButtonElement
    expect(button.disabled).toBe(true)
    expect(button.getAttribute("aria-busy")).toBe("true")
    expect(button.textContent).toBe("Salvando…")
    expect(screen.queryByText("Salvar")).toBeNull()
    expect(button.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true")
    fireEvent.click(button)
    expect(onClick).not.toHaveBeenCalled()
  })

  it("fora de andamento não expõe aria-busy e mostra o rótulo normal", () => {
    render(
      <Button loading={false} loadingLabel="Salvando…">
        Salvar
      </Button>
    )
    const button = screen.getByRole("button", { name: "Salvar" }) as HTMLButtonElement
    expect(button.disabled).toBe(false)
    expect(button.hasAttribute("aria-busy")).toBe(false)
  })

  it("exige no tipo o rótulo de progresso terminado em reticências", () => {
    // @ts-expect-error `loading` sem `loadingLabel` não compila.
    render(<Button loading>Salvar</Button>)
    // @ts-expect-error o rótulo de progresso termina em "…".
    render(<Button loading loadingLabel="Salvando">Salvar</Button>)
    expect(screen.getAllByRole("button")).toHaveLength(2)
  })
})

describe("Button: alias de compatibilidade", () => {
  it.each([
    [{ variant: "default" }, { variant: "primary" }],
    [{ variant: "outline" }, { variant: "secondary" }],
    [{ size: "default" }, { size: "md" }],
  ] as const)("%o renderiza igual a %o", (alias, canonical) => {
    render(
      <>
        <Button {...alias}>Alias</Button>
        <Button {...canonical}>Nome novo</Button>
      </>
    )
    expect(screen.getByRole("button", { name: "Alias" }).className).toBe(
      screen.getByRole("button", { name: "Nome novo" }).className
    )
  })

  it.each([
    ["icon", "size-control-md"],
    ["icon-sm", "size-control-sm"],
  ] as const)("o tamanho %s continua renderizando um botão quadrado", (size, token) => {
    render(
      <Button variant="ghost" size={size} aria-label="Abrir menu">
        <PlusIcon />
      </Button>
    )
    const button = screen.getByRole("button", { name: "Abrir menu" })
    expect(classes(button)).toContain(token)
  })
})

describe("IconButton", () => {
  it("expõe o label como nome acessível e oculta o ícone", () => {
    const onClick = vi.fn()
    render(
      <IconButton label="Excluir comprovante recibo.pdf" onClick={onClick}>
        <Trash2Icon data-testid="icone" />
      </IconButton>
    )
    const button = screen.getByRole("button", { name: "Excluir comprovante recibo.pdf" }) as HTMLButtonElement
    expect(button.tagName).toBe("BUTTON")
    expect(screen.getByTestId("icone").getAttribute("aria-hidden")).toBe("true")
    button.focus()
    expect(document.activeElement).toBe(button)
    fireEvent.click(button)
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it("sem label não compila", () => {
    render(
      // @ts-expect-error `label` é obrigatório: é o nome acessível do botão.
      <IconButton>
        <Trash2Icon />
      </IconButton>
    )
    expect(screen.getByRole("button")).toBeTruthy()
  })

  it("é discreto por padrão e aceita a variante secundária", () => {
    render(
      <>
        <IconButton label="Mais ações">
          <PlusIcon />
        </IconButton>
        <IconButton label="Filtros" variant="secondary">
          <PlusIcon />
        </IconButton>
      </>
    )
    expect(classes(screen.getByRole("button", { name: "Mais ações" }))).not.toContain("border-input")
    expect(classes(screen.getByRole("button", { name: "Filtros" }))).toContain("border-input")
  })

  it.each([
    ["sm", "size-control-sm"],
    ["md", "size-control-md"],
    ["lg", "size-control-lg"],
  ] as const)("no tamanho %s é quadrado pelo token %s", (size, token) => {
    render(
      <IconButton label="Mais ações" size={size}>
        <PlusIcon />
      </IconButton>
    )
    expect(classes(screen.getByRole("button", { name: "Mais ações" }))).toContain(token)
  })

  it("sem tamanho é o md", () => {
    render(
      <IconButton label="Mais ações">
        <PlusIcon />
      </IconButton>
    )
    expect(classes(screen.getByRole("button", { name: "Mais ações" }))).toContain("size-control-md")
  })

  it("desabilitado não dispara a ação", () => {
    const onClick = vi.fn()
    render(
      <IconButton label="Mais ações" disabled onClick={onClick}>
        <PlusIcon />
      </IconButton>
    )
    const button = screen.getByRole("button", { name: "Mais ações" }) as HTMLButtonElement
    expect(button.disabled).toBe(true)
    fireEvent.click(button)
    expect(onClick).not.toHaveBeenCalled()
  })
})
