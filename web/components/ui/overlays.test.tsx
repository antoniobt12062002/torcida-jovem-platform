import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react"
import { afterEach, describe, expect, it, vi } from "vitest"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLinkItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"

afterEach(cleanup)

describe("Dialog", () => {
  it("abre pelo gatilho com papel dialog nomeado pelo título e fecha em Fechar", async () => {
    render(
      <Dialog>
        <DialogTrigger render={<Button />}>Abrir</DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Editar item</DialogTitle>
            <DialogDescription>Altere os dados.</DialogDescription>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    )
    expect(screen.queryByRole("dialog")).toBeNull()
    fireEvent.click(screen.getByRole("button", { name: "Abrir" }))
    const dialog = await screen.findByRole("dialog", { name: "Editar item" })
    expect(dialog).toBeTruthy()
    fireEvent.click(screen.getByRole("button", { name: "Fechar" }))
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull())
  })
})

describe("AlertDialog", () => {
  it("só executa a ação ao confirmar; cancelar não a executa", async () => {
    const onConfirm = vi.fn()
    render(
      <AlertDialog>
        <AlertDialogTrigger render={<Button />}>Excluir</AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirmar exclusão</AlertDialogTitle>
            <AlertDialogDescription>Não pode ser desfeito.</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancelar</AlertDialogCancel>
            <AlertDialogAction onClick={onConfirm}>Confirmar</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    )
    fireEvent.click(screen.getByRole("button", { name: "Excluir" }))
    await screen.findByRole("alertdialog", { name: "Confirmar exclusão" })
    fireEvent.click(screen.getByRole("button", { name: "Cancelar" }))
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull())
    expect(onConfirm).not.toHaveBeenCalled()

    fireEvent.click(screen.getByRole("button", { name: "Excluir" }))
    await screen.findByRole("alertdialog")
    fireEvent.click(screen.getByRole("button", { name: "Confirmar" }))
    expect(onConfirm).toHaveBeenCalledTimes(1)
  })
})

describe("Sheet", () => {
  it("abre como diálogo nomeado e fecha em Fechar", async () => {
    render(
      <Sheet>
        <SheetTrigger render={<Button />}>Menu</SheetTrigger>
        <SheetContent side="left">
          <SheetHeader>
            <SheetTitle>Navegação</SheetTitle>
            <SheetDescription>Seções do sistema.</SheetDescription>
          </SheetHeader>
        </SheetContent>
      </Sheet>
    )
    fireEvent.click(screen.getByRole("button", { name: "Menu" }))
    await screen.findByRole("dialog", { name: "Navegação" })
    fireEvent.click(screen.getByRole("button", { name: "Fechar" }))
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull())
  })
})

describe("DropdownMenu", () => {
  it("abre o menu e expõe itens com papel menuitem", async () => {
    const onSelect = vi.fn()
    render(
      <DropdownMenu>
        <DropdownMenuTrigger render={<Button />}>Conta</DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem onClick={onSelect}>Sair</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    )
    fireEvent.click(screen.getByRole("button", { name: "Conta" }))
    expect(await screen.findByRole("menu")).toBeTruthy()
    fireEvent.click(screen.getByRole("menuitem", { name: "Sair" }))
    expect(onSelect).toHaveBeenCalledTimes(1)
  })
})

// Acrescentados na revisão do kit (DSC-04 AC1-5 e AC9, DSC-01 AC2).

function classes(el: Element) {
  return el.getAttribute("class")?.split(/\s+/) ?? []
}

function pressEscape() {
  fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" })
}

function clickOutside(target: Element) {
  fireEvent.pointerDown(target)
  fireEvent.mouseDown(target)
  fireEvent.pointerUp(target)
  fireEvent.mouseUp(target)
  fireEvent.click(target)
}

function FormDialog() {
  return (
    <Dialog>
      <DialogTrigger render={<Button />}>Nova conta</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Nova conta</DialogTitle>
          <DialogDescription>Informe o nome da conta.</DialogDescription>
        </DialogHeader>
        <input aria-label="Nome" />
        <DialogFooter>
          <DialogClose render={<Button variant="secondary" />}>Cancelar</DialogClose>
          <Button>Criar conta</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function Confirmation({ onConfirm }: { onConfirm?: () => void }) {
  return (
    <AlertDialog>
      <AlertDialogTrigger render={<Button variant="secondary" />}>Desativar conta</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Desativar a conta Bilheteria?</AlertDialogTitle>
          <AlertDialogDescription>A conta deixa de receber lançamentos.</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancelar</AlertDialogCancel>
          <AlertDialogAction variant="destructive" onClick={onConfirm}>
            Desativar conta
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

function FilterSheet({ side }: { side?: "left" | "bottom" }) {
  return (
    <Sheet>
      <SheetTrigger render={<Button variant="secondary" />}>Filtros</SheetTrigger>
      <SheetContent side={side}>
        <SheetHeader>
          <SheetTitle>Filtrar lançamentos</SheetTitle>
          <SheetDescription>Escolha os filtros.</SheetDescription>
        </SheetHeader>
        <input aria-label="Busca" />
      </SheetContent>
    </Sheet>
  )
}

function ActionsMenu({ onDeactivate }: { onDeactivate?: () => void }) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button variant="ghost" />}>Ações de Ana</DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuLinkItem href="/usuarios/1">Ver detalhes</DropdownMenuLinkItem>
        <DropdownMenuItem>Editar papéis</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem variant="destructive" onClick={onDeactivate}>
          Desativar
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

describe("Dialog: foco, Esc e clique fora", () => {
  it("abrir leva o foco ao primeiro campo e fechar devolve o foco ao gatilho", async () => {
    render(<FormDialog />)
    const trigger = screen.getByRole("button", { name: "Nova conta" })
    trigger.focus()
    fireEvent.click(trigger)
    const dialog = await screen.findByRole("dialog", { name: "Nova conta" })
    await waitFor(() => expect(document.activeElement).toBe(within(dialog).getByRole("textbox", { name: "Nome" })))
    fireEvent.click(within(dialog).getByRole("button", { name: "Cancelar" }))
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull())
    await waitFor(() => expect(document.activeElement).toBe(trigger))
  })

  it("Esc fecha e devolve o foco ao gatilho", async () => {
    render(<FormDialog />)
    const trigger = screen.getByRole("button", { name: "Nova conta" })
    trigger.focus()
    fireEvent.click(trigger)
    const dialog = await screen.findByRole("dialog")
    await waitFor(() => expect(dialog.contains(document.activeElement)).toBe(true))
    pressEscape()
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull())
    await waitFor(() => expect(document.activeElement).toBe(trigger))
  })

  it("clique fora fecha", async () => {
    render(<FormDialog />)
    fireEvent.click(screen.getByRole("button", { name: "Nova conta" }))
    await screen.findByRole("dialog")
    clickOutside(document.querySelector("[data-slot=dialog-overlay]") as Element)
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull())
  })

  it("o botão de fechar é um botão de ícone com nome acessível", async () => {
    render(<FormDialog />)
    fireEvent.click(screen.getByRole("button", { name: "Nova conta" }))
    const close = within(await screen.findByRole("dialog")).getByRole("button", { name: "Fechar" })
    expect(close.getAttribute("aria-label")).toBe("Fechar")
    expect(classes(close)).toContain("size-control-sm")
    expect(close.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true")
  })

  it("o título é um h2 e descreve o diálogo junto com a descrição", async () => {
    render(<FormDialog />)
    fireEvent.click(screen.getByRole("button", { name: "Nova conta" }))
    const dialog = await screen.findByRole("dialog", { name: "Nova conta" })
    expect(within(dialog).getByRole("heading", { level: 2, name: "Nova conta" })).toBeTruthy()
    const describedBy = dialog.getAttribute("aria-describedby") ?? ""
    expect(document.getElementById(describedBy)?.textContent).toBe("Informe o nome da conta.")
  })
})

describe("Dialog: tela estreita, rolagem e profundidade", () => {
  async function open() {
    render(<FormDialog />)
    fireEvent.click(screen.getByRole("button", { name: "Nova conta" }))
    return screen.findByRole("dialog")
  }

  it("abaixo de 640 px ocupa a tela inteira; a partir de 640 px é uma caixa centrada de 480 px", async () => {
    const names = classes(await open())
    expect(names).toEqual(expect.arrayContaining(["fixed", "inset-0", "sm:inset-auto", "sm:max-w-120", "sm:rounded-lg"]))
    expect(names.filter((name) => /^rounded/.test(name))).toEqual([])
  })

  it("aceita a largura de 640 px", async () => {
    render(
      <Dialog defaultOpen>
        <DialogContent size="wide">
          <DialogTitle>Detalhe</DialogTitle>
        </DialogContent>
      </Dialog>
    )
    expect(classes(await screen.findByRole("dialog"))).toContain("sm:max-w-content-narrow")
  })

  it("contém a rolagem e usa a sombra de nível 2 sobre o véu", async () => {
    const dialog = await open()
    expect(classes(dialog)).toEqual(
      expect.arrayContaining(["overflow-y-auto", "overscroll-contain", "shadow-elevation-2"])
    )
    expect(classes(document.querySelector("[data-slot=dialog-overlay]") as Element)).toContain("bg-overlay")
  })

  it("o rodapé fica preso à base do diálogo, não à janela", async () => {
    const dialog = await open()
    const footer = dialog.querySelector("[data-slot=dialog-footer]") as Element
    expect(classes(footer)).toEqual(expect.arrayContaining(["sticky", "bottom-0", "mt-auto"]))
    expect(classes(footer)).not.toContain("fixed")
    expect(within(footer as HTMLElement).getAllByRole("button").map((button) => button.textContent)).toEqual([
      "Cancelar",
      "Criar conta",
    ])
  })
})

describe("AlertDialog: foco, Esc e clique fora", () => {
  it("abrir leva o foco ao botão de cancelar e cancelar devolve o foco ao gatilho", async () => {
    render(<Confirmation />)
    const trigger = screen.getByRole("button", { name: "Desativar conta" })
    trigger.focus()
    fireEvent.click(trigger)
    const dialog = await screen.findByRole("alertdialog", { name: "Desativar a conta Bilheteria?" })
    const cancel = within(dialog).getByRole("button", { name: "Cancelar" })
    await waitFor(() => expect(document.activeElement).toBe(cancel))
    fireEvent.click(cancel)
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull())
    await waitFor(() => expect(document.activeElement).toBe(trigger))
  })

  it("Esc fecha sem executar a ação", async () => {
    const onConfirm = vi.fn()
    render(<Confirmation onConfirm={onConfirm} />)
    fireEvent.click(screen.getByRole("button", { name: "Desativar conta" }))
    const dialog = await screen.findByRole("alertdialog")
    await waitFor(() => expect(dialog.contains(document.activeElement)).toBe(true))
    pressEscape()
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull())
    expect(onConfirm).not.toHaveBeenCalled()
  })

  it("clique fora não fecha", async () => {
    render(<Confirmation />)
    fireEvent.click(screen.getByRole("button", { name: "Desativar conta" }))
    const dialog = await screen.findByRole("alertdialog")
    await waitFor(() => expect(dialog.contains(document.activeElement)).toBe(true))
    clickOutside(document.querySelector("[data-slot=alert-dialog-overlay]") as Element)
    clickOutside(document.body)
    await new Promise((resolve) => setTimeout(resolve, 50))
    expect(screen.getByRole("alertdialog")).toBeTruthy()
  })

  it("o botão que confirma a ação destrutiva é sólido no tom de erro e o de cancelar é secundário", async () => {
    render(<Confirmation />)
    fireEvent.click(screen.getByRole("button", { name: "Desativar conta" }))
    const dialog = await screen.findByRole("alertdialog")
    expect(classes(within(dialog).getByRole("button", { name: "Desativar conta" }))).toContain("bg-destructive")
    expect(classes(within(dialog).getByRole("button", { name: "Cancelar" }))).toContain("border-input")
  })

  it("contém a rolagem e usa a sombra de nível 2 sobre o véu", async () => {
    render(<Confirmation />)
    fireEvent.click(screen.getByRole("button", { name: "Desativar conta" }))
    expect(classes(await screen.findByRole("alertdialog"))).toEqual(
      expect.arrayContaining(["overflow-y-auto", "overscroll-contain", "shadow-elevation-2"])
    )
    expect(classes(document.querySelector("[data-slot=alert-dialog-overlay]") as Element)).toContain("bg-overlay")
  })
})

describe("Sheet: foco, Esc e lados", () => {
  it("abrir leva o foco para dentro e fechar devolve o foco ao gatilho", async () => {
    render(<FilterSheet side="bottom" />)
    const trigger = screen.getByRole("button", { name: "Filtros" })
    trigger.focus()
    fireEvent.click(trigger)
    const sheet = await screen.findByRole("dialog", { name: "Filtrar lançamentos" })
    await waitFor(() => expect(document.activeElement).toBe(within(sheet).getByRole("textbox", { name: "Busca" })))
    fireEvent.click(within(sheet).getByRole("button", { name: "Fechar" }))
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull())
    await waitFor(() => expect(document.activeElement).toBe(trigger))
  })

  it("Esc fecha e devolve o foco ao gatilho", async () => {
    render(<FilterSheet />)
    const trigger = screen.getByRole("button", { name: "Filtros" })
    trigger.focus()
    fireEvent.click(trigger)
    const sheet = await screen.findByRole("dialog")
    await waitFor(() => expect(sheet.contains(document.activeElement)).toBe(true))
    pressEscape()
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull())
    await waitFor(() => expect(document.activeElement).toBe(trigger))
  })

  it.each([
    ["left", ["inset-y-0", "left-0", "border-r"]],
    ["bottom", ["inset-x-0", "bottom-0", "border-t"]],
  ] as const)("o lado %s prende a gaveta à borda certa", async (side, expected) => {
    render(<FilterSheet side={side} />)
    fireEvent.click(screen.getByRole("button", { name: "Filtros" }))
    const sheet = await screen.findByRole("dialog")
    expect(sheet.getAttribute("data-side")).toBe(side)
    expect(classes(sheet)).toEqual(expect.arrayContaining([...expected]))
  })

  it("sem lado informado abre à esquerda", async () => {
    render(<FilterSheet />)
    fireEvent.click(screen.getByRole("button", { name: "Filtros" }))
    expect((await screen.findByRole("dialog")).getAttribute("data-side")).toBe("left")
  })

  it("contém a rolagem, usa a sombra de nível 2 e o botão de fechar é um botão de ícone", async () => {
    render(<FilterSheet side="bottom" />)
    fireEvent.click(screen.getByRole("button", { name: "Filtros" }))
    const sheet = await screen.findByRole("dialog")
    expect(classes(sheet)).toEqual(
      expect.arrayContaining(["overflow-y-auto", "overscroll-contain", "shadow-elevation-2"])
    )
    expect(classes(document.querySelector("[data-slot=sheet-overlay]") as Element)).toContain("bg-overlay")
    const close = within(sheet).getByRole("button", { name: "Fechar" })
    expect(close.getAttribute("aria-label")).toBe("Fechar")
    expect(classes(close)).toContain("size-control-sm")
  })
})

describe("DropdownMenu: foco, Esc e itens", () => {
  it("abrir leva o foco para dentro do menu; Esc fecha e devolve o foco ao gatilho", async () => {
    render(<ActionsMenu />)
    const trigger = screen.getByRole("button", { name: "Ações de Ana" })
    trigger.focus()
    fireEvent.click(trigger)
    const menu = await screen.findByRole("menu")
    await waitFor(() => expect(menu.contains(document.activeElement)).toBe(true))
    pressEscape()
    await waitFor(() => expect(screen.queryByRole("menu")).toBeNull())
    await waitFor(() => expect(document.activeElement).toBe(trigger))
  })

  it("as setas percorrem os itens", async () => {
    render(<ActionsMenu />)
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana" }))
    const menu = await screen.findByRole("menu")
    await waitFor(() => expect(menu.contains(document.activeElement)).toBe(true))
    fireEvent.keyDown(document.activeElement as Element, { key: "ArrowDown" })
    await waitFor(() => expect(document.activeElement?.getAttribute("role")).toBe("menuitem"))
    const first = document.activeElement
    fireEvent.keyDown(first as Element, { key: "ArrowDown" })
    await waitFor(() => expect(document.activeElement).not.toBe(first))
    expect(document.activeElement?.getAttribute("role")).toBe("menuitem")
  })

  it("o item que navega é um link com o papel de item de menu", async () => {
    render(<ActionsMenu />)
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana" }))
    const link = await screen.findByRole("menuitem", { name: "Ver detalhes" })
    expect(link.tagName).toBe("A")
    expect(link.getAttribute("href")).toBe("/usuarios/1")
  })

  it("o item que navega aceita o link da aplicação por render", async () => {
    render(
      <DropdownMenu>
        <DropdownMenuTrigger render={<Button />}>Conta</DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuLinkItem render={<a href="/conta/senha" data-testid="link-da-aplicacao" />}>
            Trocar senha
          </DropdownMenuLinkItem>
        </DropdownMenuContent>
      </DropdownMenu>
    )
    fireEvent.click(screen.getByRole("button", { name: "Conta" }))
    const link = await screen.findByRole("menuitem", { name: "Trocar senha" })
    expect(link.getAttribute("data-testid")).toBe("link-da-aplicacao")
    expect(link.getAttribute("href")).toBe("/conta/senha")
  })

  it("o item destrutivo usa o tom de erro e continua acionável", async () => {
    const onDeactivate = vi.fn()
    render(<ActionsMenu onDeactivate={onDeactivate} />)
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana" }))
    const item = await screen.findByRole("menuitem", { name: "Desativar" })
    expect(item.getAttribute("data-variant")).toBe("destructive")
    expect(classes(item)).toContain("data-[variant=destructive]:text-destructive")
    expect(screen.getByRole("separator")).toBeTruthy()
    fireEvent.click(item)
    expect(onDeactivate).toHaveBeenCalledTimes(1)
  })

  it("os itens e o link têm a altura do token de controle e o menu usa a sombra de nível 1", async () => {
    render(<ActionsMenu />)
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana" }))
    const menu = await screen.findByRole("menu")
    for (const item of within(menu).getAllByRole("menuitem")) {
      expect(classes(item), item.textContent ?? "").toContain("min-h-control-md")
    }
    expect(classes(menu)).toContain("shadow-elevation-1")
  })
})

// Revisão da rodada 4 (DSC-01 AC2): o item realçado por teclado tem indicador
// de foco de 2 px no token `ring`, não só o fundo de realce.
const ITEM_FOCUS_RING = [
  "focus-visible:ring-2",
  "focus-visible:ring-ring",
  "focus-visible:ring-inset",
  "data-highlighted:ring-2",
  "data-highlighted:ring-ring",
  "data-highlighted:ring-inset",
]
const ITEM_HIGHLIGHT_BG = ["focus:bg-accent", "data-highlighted:bg-accent"]

describe("DropdownMenu: indicador de foco do item", () => {
  it("o item realçado pelas setas recebe data-highlighted e tem anel de 2 px no token ring, além do fundo", async () => {
    render(<ActionsMenu />)
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana" }))
    const menu = await screen.findByRole("menu")
    await waitFor(() => expect(menu.contains(document.activeElement)).toBe(true))
    fireEvent.keyDown(document.activeElement as Element, { key: "ArrowDown" })
    await waitFor(() => expect(document.activeElement?.getAttribute("role")).toBe("menuitem"))
    const highlighted = document.activeElement as Element
    await waitFor(() => expect(highlighted.hasAttribute("data-highlighted")).toBe(true))
    expect(classes(highlighted)).toEqual(expect.arrayContaining([...ITEM_FOCUS_RING, ...ITEM_HIGHLIGHT_BG]))
  })

  it("item, link, caixa de seleção, opção de rádio e abertura de submenu têm o mesmo indicador", async () => {
    render(
      <DropdownMenu>
        <DropdownMenuTrigger render={<Button />}>Opções</DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem>Editar</DropdownMenuItem>
          <DropdownMenuLinkItem href="/x">Abrir</DropdownMenuLinkItem>
          <DropdownMenuCheckboxItem checked>Arquivados</DropdownMenuCheckboxItem>
          <DropdownMenuRadioGroup value="a">
            <DropdownMenuRadioItem value="a">Ordem A</DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>Mais</DropdownMenuSubTrigger>
          </DropdownMenuSub>
        </DropdownMenuContent>
      </DropdownMenu>
    )
    fireEvent.click(screen.getByRole("button", { name: "Opções" }))
    const menu = await screen.findByRole("menu")
    const items = [
      ...within(menu).getAllByRole("menuitem"),
      within(menu).getByRole("menuitemcheckbox"),
      within(menu).getByRole("menuitemradio"),
    ]
    expect(items).toHaveLength(5)
    for (const item of items) {
      expect(classes(item), item.textContent ?? "").toEqual(
        expect.arrayContaining([...ITEM_FOCUS_RING, ...ITEM_HIGHLIGHT_BG])
      )
    }
  })

  it("o item destrutivo realçado mantém o tom de erro", async () => {
    render(<ActionsMenu />)
    fireEvent.click(screen.getByRole("button", { name: "Ações de Ana" }))
    const item = await screen.findByRole("menuitem", { name: "Desativar" })
    expect(classes(item)).toEqual(
      expect.arrayContaining([
        "data-[variant=destructive]:data-highlighted:bg-destructive-subtle",
        "data-[variant=destructive]:data-highlighted:text-destructive-emphasis",
        ...ITEM_FOCUS_RING,
      ])
    )
  })
})

describe("Select: indicador de foco da opção", () => {
  it("a opção tem anel de 2 px no token ring quando realçada, além do fundo", async () => {
    render(
      <Select items={[{ value: "a", label: "Receita" }, { value: "b", label: "Despesa" }]} defaultOpen>
        <SelectTrigger aria-label="Tipo">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">Receita</SelectItem>
          <SelectItem value="b">Despesa</SelectItem>
        </SelectContent>
      </Select>
    )
    const options = await screen.findAllByRole("option")
    expect(options).toHaveLength(2)
    for (const option of options) {
      expect(classes(option), option.textContent ?? "").toEqual(
        expect.arrayContaining([...ITEM_FOCUS_RING, ...ITEM_HIGHLIGHT_BG])
      )
    }
  })
})
