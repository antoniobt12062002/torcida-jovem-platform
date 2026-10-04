import { cleanup, render, screen } from "@testing-library/react"
import { afterEach, describe, expect, it } from "vitest"

import { Checkbox } from "@/components/ui/checkbox"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"

afterEach(cleanup)

function describedByText(el: HTMLElement) {
  const ids = (el.getAttribute("aria-describedby") ?? "").split(" ").filter(Boolean)
  return ids.map((id) => document.getElementById(id)?.textContent ?? "")
}

describe("Input em Field", () => {
  it("associa o rótulo ao campo", () => {
    render(
      <Field>
        <FieldLabel>E-mail</FieldLabel>
        <Input />
      </Field>
    )
    expect(screen.getByRole("textbox", { name: "E-mail" })).toBeTruthy()
  })

  it("com erro marca aria-invalid e aponta aria-describedby para a mensagem", () => {
    render(
      <Field invalid>
        <FieldLabel>E-mail</FieldLabel>
        <Input />
        <FieldError>Informe um e-mail válido.</FieldError>
      </Field>
    )
    const input = screen.getByRole("textbox", { name: "E-mail" })
    expect(input.getAttribute("aria-invalid")).toBe("true")
    expect(describedByText(input)).toContain("Informe um e-mail válido.")
    expect(screen.getByRole("alert").textContent).toBe(
      "Informe um e-mail válido."
    )
  })

  it("sem erro não marca aria-invalid", () => {
    render(
      <Field>
        <FieldLabel>E-mail</FieldLabel>
        <Input />
        <FieldError />
      </Field>
    )
    const input = screen.getByRole("textbox", { name: "E-mail" })
    expect(input.getAttribute("aria-invalid")).not.toBe("true")
    expect(screen.queryByRole("alert")).toBeNull()
  })

  it("inclui a descrição em aria-describedby", () => {
    render(
      <Field>
        <FieldLabel>Nome</FieldLabel>
        <Input />
        <FieldDescription>Como aparece nos relatórios.</FieldDescription>
      </Field>
    )
    const input = screen.getByRole("textbox", { name: "Nome" })
    expect(describedByText(input)).toContain("Como aparece nos relatórios.")
  })
})

describe("Textarea em Field", () => {
  it("com erro marca aria-invalid e aria-describedby", () => {
    render(
      <Field invalid>
        <FieldLabel>Motivo</FieldLabel>
        <Textarea />
        <FieldError>Informe o motivo.</FieldError>
      </Field>
    )
    const area = screen.getByRole("textbox", { name: "Motivo" })
    expect(area.tagName).toBe("TEXTAREA")
    expect(area.getAttribute("aria-invalid")).toBe("true")
    expect(describedByText(area)).toContain("Informe o motivo.")
  })
})

describe("Checkbox em Field", () => {
  it("expõe o papel checkbox com o rótulo e o erro associados", () => {
    render(
      <Field invalid>
        <FieldLabel>
          <Checkbox />
          Aceito os termos
        </FieldLabel>
        <FieldError>É obrigatório aceitar.</FieldError>
      </Field>
    )
    const box = screen.getByRole("checkbox", { name: /Aceito os termos/ })
    expect(box.getAttribute("aria-invalid")).toBe("true")
    expect(describedByText(box)).toContain("É obrigatório aceitar.")
  })
})

describe("Select em Field", () => {
  it("o gatilho tem papel combobox, rótulo e erro associados", () => {
    render(
      <Field invalid>
        <FieldLabel>Categoria</FieldLabel>
        <Select items={[{ value: "a", label: "A" }]}>
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="a">A</SelectItem>
          </SelectContent>
        </Select>
        <FieldError>Escolha uma categoria.</FieldError>
      </Field>
    )
    const trigger = screen.getByRole("combobox")
    expect(trigger.getAttribute("aria-invalid")).toBe("true")
    expect(describedByText(trigger)).toContain("Escolha uma categoria.")
    expect(trigger.getAttribute("aria-labelledby")).toBeTruthy()
  })
})
