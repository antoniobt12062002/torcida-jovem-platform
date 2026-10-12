import { act, cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react"
import { CircleDashedIcon } from "lucide-react"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Toaster, toast } from "@/components/ui/toast"

afterEach(cleanup)

describe("Table", () => {
  it("expõe tabela, cabeçalhos de coluna e células", () => {
    render(
      <Table>
        <TableCaption>Lançamentos</TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead>Descrição</TableHead>
            <TableHead>Valor</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow>
            <TableCell>Mensalidade</TableCell>
            <TableCell>R$ 10,00</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    )
    expect(screen.getByRole("table", { name: "Lançamentos" })).toBeTruthy()
    expect(screen.getAllByRole("columnheader").length).toBe(2)
    expect(screen.getByRole("cell", { name: "Mensalidade" })).toBeTruthy()
  })
})

describe("Alert", () => {
  it("tem papel alert com título e descrição", () => {
    render(
      <Alert variant="destructive">
        <AlertTitle>Falha</AlertTitle>
        <AlertDescription>Não foi possível salvar.</AlertDescription>
      </Alert>
    )
    const alert = screen.getByRole("alert")
    expect(alert.textContent).toContain("Falha")
    expect(alert.textContent).toContain("Não foi possível salvar.")
  })
})

describe("Skeleton", () => {
  it("renderiza um marcador de carregamento oculto da árvore de leitura", () => {
    const { container } = render(<Skeleton className="h-4 w-10" />)
    const el = container.firstElementChild as HTMLElement
    expect(el).toBeTruthy()
    expect(el.className).toContain("h-4")
  })
})

describe("Toast", () => {
  it("dispara toast pela API e permite fechar", async () => {
    render(<Toaster />)
    act(() => {
      toast.success("Salvo com sucesso", { description: "Lançamento criado.", timeout: 0 })
    })
    expect(await screen.findByText("Salvo com sucesso")).toBeTruthy()
    expect(screen.getByText("Lançamento criado.")).toBeTruthy()
    fireEvent.click(screen.getByLabelText("Fechar aviso"))
    await waitFor(() => expect(screen.queryByText("Salvo com sucesso")).toBeNull())
  })
})

// Acrescentados na revisão do kit (DSC-01 AC7, DSC-04 AC6-7, DSC-05 AC2).

function classes(el: Element) {
  return el.getAttribute("class")?.split(/\s+/) ?? []
}

/** Nome do ícone do Lucide (`lucide-circle-check`), para distinguir os tons sem depender de cor. */
function iconName(el: Element | null) {
  const icon = el?.matches("svg") ? el : el?.querySelector("svg")
  return classes(icon as Element).find((name) => /^lucide-.+/.test(name))
}

const BADGE_TONES = [
  ["neutral", "bg-info-subtle", "text-info"],
  ["success", "bg-success-subtle", "text-success-emphasis"],
  ["warning", "bg-warning-subtle", "text-warning-emphasis"],
  ["error", "bg-destructive-subtle", "text-destructive-emphasis"],
  ["outline", "border-border", "text-foreground"],
] as const

describe("Badge", () => {
  it.each(BADGE_TONES)("tom %s mostra o rótulo e o ícone, com fundo e texto do tom", (tone, surface, text) => {
    render(
      <Badge variant={tone}>
        <CircleDashedIcon />
        Criada
      </Badge>
    )
    const badge = screen.getByText("Criada")
    expect(badge.tagName).toBe("SPAN")
    expect(classes(badge)).toEqual(expect.arrayContaining([surface, text]))
    expect(badge.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true")
  })

  it("sem variante é o tom neutro", () => {
    render(
      <>
        <Badge>Padrão</Badge>
        <Badge variant="neutral">Neutro</Badge>
      </>
    )
    expect(screen.getByText("Padrão").className).toBe(screen.getByText("Neutro").className)
  })

  it("é uma pílula de texto, não uma ação, e não usa transition-all", () => {
    render(<Badge variant="success">Recebida</Badge>)
    const badge = screen.getByText("Recebida")
    expect(classes(badge)).toEqual(expect.arrayContaining(["rounded-full", "text-xs", "font-medium"]))
    expect(classes(badge)).not.toContain("transition-all")
    expect(classes(badge).filter((name) => /hover:|focus-visible:/.test(name))).toEqual([])
    expect(badge.getAttribute("tabindex")).toBeNull()
    expect(screen.queryByRole("button")).toBeNull()
  })

  it.each([
    ["destructive", "error"],
    ["secondary", "neutral"],
  ] as const)("o alias %s renderiza igual ao tom %s", (alias, tone) => {
    render(
      <>
        <Badge variant={alias}>Alias</Badge>
        <Badge variant={tone}>Tom</Badge>
      </>
    )
    expect(screen.getByText("Alias").className).toBe(screen.getByText("Tom").className)
  })

  it("o alias default continua renderizando, em preto", () => {
    render(<Badge variant="default">Paga</Badge>)
    expect(classes(screen.getByText("Paga"))).toEqual(
      expect.arrayContaining(["bg-primary", "text-primary-foreground"])
    )
  })
})

const ALERT_TONES = [
  ["neutral", "status", "lucide-info", "bg-info-subtle", "border-info"],
  ["success", "status", "lucide-circle-check", "bg-success-subtle", "border-success"],
  ["warning", "alert", "lucide-triangle-alert", "bg-warning-subtle", "border-warning"],
  ["error", "alert", "lucide-circle-alert", "bg-destructive-subtle", "border-destructive"],
] as const

describe("Alert: tons", () => {
  it.each(ALERT_TONES)("tom %s tem papel %s, ícone próprio e texto", (tone, role, icon, surface, border) => {
    render(
      <Alert variant={tone}>
        <AlertTitle>Título</AlertTitle>
        <AlertDescription>Descrição do aviso.</AlertDescription>
      </Alert>
    )
    const alert = screen.getByRole(role)
    expect(alert.textContent).toBe("TítuloDescrição do aviso.")
    expect(iconName(alert)).toBe(icon)
    expect(alert.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true")
    expect(classes(alert)).toEqual(expect.arrayContaining(["border", surface, border]))
  })

  it("os quatro tons têm ícones diferentes entre si", () => {
    expect(new Set(ALERT_TONES.map(([, , icon]) => icon)).size).toBe(4)
  })

  it("o alias destructive é o tom de erro", () => {
    render(
      <>
        <Alert variant="destructive">
          <AlertDescription>Alias</AlertDescription>
        </Alert>
        <Alert variant="error">
          <AlertDescription>Tom</AlertDescription>
        </Alert>
      </>
    )
    const [alias, tone] = screen.getAllByRole("alert")
    expect(alias.className).toBe(tone.className)
    expect(iconName(alias)).toBe("lucide-circle-alert")
  })

  it("sem variante (alias default) tem a aparência neutra e mantém o papel alert", () => {
    render(
      <>
        <Alert>
          <AlertDescription>Sua sessão foi encerrada.</AlertDescription>
        </Alert>
        <Alert variant="default">
          <AlertDescription>Senha alterada com sucesso.</AlertDescription>
        </Alert>
        <Alert variant="neutral">
          <AlertDescription>Neutro</AlertDescription>
        </Alert>
      </>
    )
    const alerts = screen.getAllByRole("alert")
    expect(alerts.map((alert) => alert.textContent)).toEqual([
      "Sua sessão foi encerrada.",
      "Senha alterada com sucesso.",
    ])
    for (const alert of alerts) {
      expect(alert.className).toBe(screen.getByRole("status").className)
      expect(iconName(alert)).toBe("lucide-info")
    }
  })

  it("aceita um papel explícito de quem chama", () => {
    render(
      <Alert variant="error" role="status">
        <AlertDescription>Só lançamentos em aberto podem ser editados.</AlertDescription>
      </Alert>
    )
    expect(screen.getByRole("status").textContent).toBe("Só lançamentos em aberto podem ser editados.")
  })
})

describe("Skeleton: formas", () => {
  it.each([
    ["text", "h-4"],
    ["row", "h-table-row"],
    ["block", "h-10"],
  ] as const)("a forma %s tem a altura %s", (shape, height) => {
    const { container } = render(<Skeleton shape={shape} />)
    expect(classes(container.firstElementChild as Element)).toContain(height)
  })

  it("é decorativo, em muted, com pulso que para sob redução de movimento", () => {
    const { container } = render(<Skeleton shape="text" />)
    const el = container.firstElementChild as Element
    expect(el.getAttribute("aria-hidden")).toBe("true")
    expect(classes(el)).toEqual(expect.arrayContaining(["bg-muted", "animate-pulse", "motion-reduce:animate-none"]))
  })
})

describe("Toast: API e tons", () => {
  it.each([
    ["show", "lucide-info"],
    ["success", "lucide-circle-check"],
    ["warning", "lucide-triangle-alert"],
    ["error", "lucide-circle-alert"],
  ] as const)("toast.%s(título, opções) devolve o id e mostra título, descrição e ícone do tom", async (method, icon) => {
    render(<Toaster />)
    let id = ""
    act(() => {
      id = toast[method]("Título do aviso", { description: "Detalhe do aviso.", timeout: 0 })
    })
    expect(typeof id).toBe("string")
    expect(id).not.toBe("")
    const item = (await screen.findByText("Título do aviso", { ignore: "[role=alert] *" })).closest("[data-slot=toast]")
    expect(item?.textContent).toContain("Detalhe do aviso.")
    expect(iconName(item)).toBe(icon)
    expect(item?.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true")
    act(() => toast.dismiss(id))
    await waitFor(() => expect(document.querySelector("[data-slot=toast]")).toBeNull())
  })

  it("aceita só o título", async () => {
    render(<Toaster />)
    act(() => {
      toast.success("Conta criada.")
    })
    expect(await screen.findByText("Conta criada.")).toBeTruthy()
  })
})

describe("Toast: anúncio e teclado", () => {
  it("aparece dentro de uma região viva nomeada em português", async () => {
    render(<Toaster />)
    act(() => {
      toast.success("Salvo com sucesso", { timeout: 0 })
    })
    const region = screen.getByRole("region", { name: "Avisos" })
    expect(region.getAttribute("aria-live")).toBe("polite")
    expect(await within(region).findByText("Salvo com sucesso")).toBeTruthy()
  })

  // SPEC_DEVIATION: o design pede o erro como `assertive`. No Base UI isso
  // duplica o texto do aviso na página, e os testes das telas ainda não
  // migradas procuram a mensagem de erro como texto único. Até DS-INT, todos
  // os tons, inclusive o erro, são anunciados pela região viva, uma só vez.
  it.each(["show", "success", "warning", "error"] as const)(
    "toast.%s é anunciado pela região viva, com o texto uma única vez na página",
    async (method) => {
      render(<Toaster />)
      act(() => {
        toast[method]("Não foi possível sair.", { timeout: 0 })
      })
      const region = screen.getByRole("region", { name: "Avisos" })
      expect(await within(region).findByText("Não foi possível sair.")).toBeTruthy()
      expect(screen.getAllByText("Não foi possível sair.")).toHaveLength(1)
    }
  )

  it("o botão de fechar, ao receber foco, é um botão com nome e alvo pelo token de altura", async () => {
    render(<Toaster />)
    act(() => {
      toast.success("Salvo com sucesso", { timeout: 0 })
    })
    // O Base UI só expõe o botão de fechar quando o aviso tem foco ou está expandido.
    const close = await screen.findByLabelText("Fechar aviso")
    act(() => close.focus())
    expect(screen.getByRole("button", { name: "Fechar aviso" })).toBe(close)
    expect(close.tagName).toBe("BUTTON")
    expect(classes(close)).toContain("size-control-sm")
    expect(close.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true")
  })

  it("Esc com o foco no aviso dispensa o aviso", async () => {
    render(<Toaster />)
    act(() => {
      toast.success("Salvo com sucesso", { timeout: 0 })
    })
    const close = await screen.findByLabelText("Fechar aviso")
    act(() => close.focus())
    expect(document.activeElement).toBe(close)
    fireEvent.keyDown(close, { key: "Escape" })
    await waitFor(() => expect(screen.queryByText("Salvo com sucesso")).toBeNull())
  })
})

describe("Toast: duração", () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  function shown() {
    return document.querySelector("[data-slot=toast]") !== null
  }

  it.each([
    ["show", 5000],
    ["success", 5000],
    ["warning", 8000],
    ["error", 8000],
  ] as const)("toast.%s some depois de %i ms", (method, duration) => {
    render(<Toaster />)
    act(() => {
      toast[method]("Título do aviso")
    })
    act(() => {
      vi.advanceTimersByTime(duration - 100)
    })
    expect(shown()).toBe(true)
    act(() => {
      vi.advanceTimersByTime(200)
    })
    expect(shown()).toBe(false)
  })

  it("o tempo informado por quem chama prevalece", () => {
    render(<Toaster />)
    act(() => {
      toast.error("Título do aviso", { timeout: 1000 })
    })
    act(() => {
      vi.advanceTimersByTime(1100)
    })
    expect(shown()).toBe(false)
  })

  it("a contagem pausa enquanto o ponteiro está sobre os avisos", () => {
    render(<Toaster />)
    act(() => {
      toast.success("Título do aviso")
    })
    fireEvent.mouseEnter(screen.getByRole("region", { name: "Avisos" }))
    act(() => {
      vi.advanceTimersByTime(20000)
    })
    expect(shown()).toBe(true)
  })

  it("a contagem pausa enquanto o foco está no aviso", () => {
    render(<Toaster />)
    act(() => {
      toast.success("Título do aviso")
    })
    act(() => {
      ;(document.querySelector("[data-slot=toast]") as HTMLElement).focus()
    })
    act(() => {
      vi.advanceTimersByTime(20000)
    })
    expect(shown()).toBe(true)
  })
})
