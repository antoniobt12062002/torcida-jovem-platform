import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react"
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
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
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
