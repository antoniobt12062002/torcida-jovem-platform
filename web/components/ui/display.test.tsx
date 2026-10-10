import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react"
import { afterEach, describe, expect, it } from "vitest"

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
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
