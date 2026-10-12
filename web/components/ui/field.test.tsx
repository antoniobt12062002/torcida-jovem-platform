import { cleanup, fireEvent, render, screen } from "@testing-library/react"
import { afterEach, describe, expect, it, vi } from "vitest"

import { Checkbox } from "@/components/ui/checkbox"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { NativeSelect } from "@/components/ui/native-select"
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

// Acrescentados na revisão do kit (DSC-03, DSF-03 AC7).

function classes(el: Element) {
  return el.getAttribute("class")?.split(/\s+/) ?? []
}

function renderControls(props: { disabled?: boolean; readOnly?: boolean } = {}) {
  render(
    <>
      <Input aria-label="Nome" {...props} />
      <Textarea aria-label="Motivo" {...props} />
      <NativeSelect aria-label="Tipo" disabled={props.disabled}>
        <option value="a">A</option>
      </NativeSelect>
    </>
  )
  return [
    screen.getByRole("textbox", { name: "Nome" }),
    screen.getByRole("textbox", { name: "Motivo" }),
    screen.getByRole("combobox", { name: "Tipo" }),
  ]
}

describe("associação de rótulo em Field", () => {
  it("getByLabelText encontra o Input", () => {
    render(
      <Field>
        <FieldLabel>Buscar por código ou nome</FieldLabel>
        <Input type="search" />
      </Field>
    )
    expect(screen.getByLabelText("Buscar por código ou nome").tagName).toBe("INPUT")
  })

  it("getByLabelText encontra a Textarea", () => {
    render(
      <Field>
        <FieldLabel>Motivo</FieldLabel>
        <Textarea />
      </Field>
    )
    expect(screen.getByLabelText("Motivo").tagName).toBe("TEXTAREA")
  })

  it("getByLabelText encontra o NativeSelect", () => {
    render(
      <Field>
        <FieldLabel>Tipo</FieldLabel>
        <NativeSelect>
          <option value="RECEITA">Receita</option>
        </NativeSelect>
      </Field>
    )
    expect(screen.getByLabelText("Tipo").tagName).toBe("SELECT")
  })

  it("o rótulo é um label nativo ligado ao campo, portanto clicável", () => {
    render(
      <Field>
        <FieldLabel>E-mail</FieldLabel>
        <Input />
      </Field>
    )
    const label = screen.getByText("E-mail") as HTMLLabelElement
    expect(label.tagName).toBe("LABEL")
    expect(label.control).toBe(screen.getByRole("textbox", { name: "E-mail" }))
  })
})

describe("FieldLabel opcional", () => {
  it("marca o campo opcional com (opcional)", () => {
    render(
      <Field>
        <FieldLabel optional>Observação</FieldLabel>
        <Input />
      </Field>
    )
    expect(screen.getByRole("textbox", { name: "Observação (opcional)" })).toBeTruthy()
  })

  it("campo obrigatório não leva marca", () => {
    render(
      <Field>
        <FieldLabel>Descrição</FieldLabel>
        <Input required />
      </Field>
    )
    expect(screen.getByText("Descrição").textContent).toBe("Descrição")
    expect(screen.queryByText(/opcional/)).toBeNull()
  })
})

describe("FieldError", () => {
  it("mostra a mensagem com ícone decorativo e texto", () => {
    render(
      <Field invalid>
        <FieldLabel>E-mail</FieldLabel>
        <Input />
        <FieldError>Informe um e-mail válido.</FieldError>
      </Field>
    )
    const error = screen.getByRole("alert")
    expect(error.textContent).toBe("Informe um e-mail válido.")
    const icon = error.querySelector("svg")
    expect(icon).toBeTruthy()
    expect(icon?.getAttribute("aria-hidden")).toBe("true")
  })

  it("aceita a lista de erros do formulário", () => {
    render(
      <Field invalid>
        <FieldLabel>Senha</FieldLabel>
        <Input />
        <FieldError errors={[{ message: "Mínimo de 12 caracteres." }]} />
      </Field>
    )
    expect(screen.getByRole("alert").textContent).toBe("Mínimo de 12 caracteres.")
    expect(describedByText(screen.getByRole("textbox", { name: "Senha" }))).toContain(
      "Mínimo de 12 caracteres."
    )
  })
})

describe("NativeSelect", () => {
  it("em Field com erro marca aria-invalid e aria-describedby", () => {
    render(
      <Field invalid>
        <FieldLabel>Tipo</FieldLabel>
        <NativeSelect>
          <option value="">Selecione</option>
          <option value="RECEITA">Receita</option>
        </NativeSelect>
        <FieldError>Escolha o tipo.</FieldError>
      </Field>
    )
    const select = screen.getByRole("combobox", { name: "Tipo" })
    expect(select.getAttribute("aria-invalid")).toBe("true")
    expect(describedByText(select)).toContain("Escolha o tipo.")
  })

  it("fora de Field aceita invalid e não marca erro sem ele", () => {
    render(
      <>
        <NativeSelect aria-label="Status" invalid>
          <option value="a">A</option>
        </NativeSelect>
        <NativeSelect aria-label="Conta">
          <option value="a">A</option>
        </NativeSelect>
      </>
    )
    expect(screen.getByRole("combobox", { name: "Status" }).getAttribute("aria-invalid")).toBe("true")
    expect(screen.getByRole("combobox", { name: "Conta" }).getAttribute("aria-invalid")).not.toBe("true")
  })

  it("é um select nativo controlado, com opções e mudança de valor", () => {
    const onChange = vi.fn()
    render(
      <NativeSelect aria-label="Tipo" name="tipo" value="DESPESA" onChange={(event) => onChange(event.target.value)}>
        <option value="RECEITA">Receita</option>
        <option value="DESPESA">Despesa</option>
      </NativeSelect>
    )
    const select = screen.getByRole("combobox", { name: "Tipo" }) as HTMLSelectElement
    expect(select.tagName).toBe("SELECT")
    expect(select.name).toBe("tipo")
    expect(select.value).toBe("DESPESA")
    expect(screen.getAllByRole("option").map((option) => option.textContent)).toEqual(["Receita", "Despesa"])
    fireEvent.change(select, { target: { value: "RECEITA" } })
    expect(onChange).toHaveBeenCalledWith("RECEITA")
  })

  it("sem valor controlado guarda a escolha da pessoa", () => {
    render(
      <NativeSelect aria-label="Tipo" defaultValue="DESPESA">
        <option value="RECEITA">Receita</option>
        <option value="DESPESA">Despesa</option>
      </NativeSelect>
    )
    const select = screen.getByRole("combobox", { name: "Tipo" }) as HTMLSelectElement
    expect(select.value).toBe("DESPESA")
    fireEvent.change(select, { target: { value: "RECEITA" } })
    expect(select.value).toBe("RECEITA")
  })

  it("tem aparência própria, seta decorativa e cores explícitas, também nas opções", () => {
    const { container } = render(
      <NativeSelect aria-label="Tipo">
        <option value="a">A</option>
      </NativeSelect>
    )
    const select = screen.getByRole("combobox", { name: "Tipo" })
    expect(classes(select)).toEqual(
      expect.arrayContaining([
        "appearance-none",
        "bg-background",
        "[&_option]:bg-background",
        "[&_option]:text-foreground",
      ])
    )
    expect(container.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true")
  })
})

describe("Input, Textarea e NativeSelect: medidas e estados iguais", () => {
  it("Input e NativeSelect usam o mesmo token de altura; a Textarea parte de três linhas", () => {
    const [input, textarea, select] = renderControls()
    expect(classes(input)).toContain("h-control-md")
    expect(classes(select)).toContain("h-control-md")
    expect(classes(textarea)).toContain("min-h-20")
    expect(classes(textarea)).toContain("resize-y")
  })

  it.each([
    ["borda de controle", ["border", "border-input", "rounded-md"]],
    ["texto pelo token de controle, com cor explícita", ["text-control", "[&]:text-foreground", "bg-background"]],
    [
      "foco por anel de 2 px afastado",
      [
        "focus-visible:ring-2",
        "focus-visible:ring-ring",
        "focus-visible:ring-offset-2",
        "focus-visible:ring-offset-ring-offset",
      ],
    ],
    ["inválido com borda de erro", ["aria-invalid:border-destructive"]],
    ["desabilitado", ["disabled:bg-disabled", "disabled:text-disabled-foreground"]],
    ["somente leitura", ["[&[readonly]]:bg-muted", "aria-readonly:bg-muted"]],
  ])("os três têm o mesmo estado: %s", (_state, expected) => {
    for (const control of renderControls()) {
      expect(classes(control), control.tagName).toEqual(expect.arrayContaining(expected))
    }
  })

  it("nenhum usa vermelho no foco", () => {
    for (const control of renderControls()) {
      expect(classes(control).filter((name) => /focus-visible:.*(destructive|brand)/.test(name))).toEqual([])
    }
  })

  it("desabilitado e somente leitura chegam ao elemento nativo", () => {
    const [input, textarea, select] = renderControls({ disabled: true, readOnly: true })
    expect((input as HTMLInputElement).disabled).toBe(true)
    expect((input as HTMLInputElement).readOnly).toBe(true)
    expect((textarea as HTMLTextAreaElement).disabled).toBe(true)
    expect((textarea as HTMLTextAreaElement).readOnly).toBe(true)
    expect((select as HTMLSelectElement).disabled).toBe(true)
  })
})

describe("Checkbox: alvo único com o rótulo", () => {
  it("clicar no rótulo dentro de Field alterna o controle", () => {
    const onCheckedChange = vi.fn()
    render(
      <Field>
        <FieldLabel>
          <Checkbox onCheckedChange={(checked) => onCheckedChange(checked)} />
          Aceito os termos
        </FieldLabel>
      </Field>
    )
    const box = screen.getByRole("checkbox", { name: /Aceito os termos/ })
    expect(box.getAttribute("aria-checked")).toBe("false")
    fireEvent.click(screen.getByText("Aceito os termos"))
    expect(box.getAttribute("aria-checked")).toBe("true")
    expect(onCheckedChange).toHaveBeenLastCalledWith(true)
    fireEvent.click(screen.getByText("Aceito os termos"))
    expect(box.getAttribute("aria-checked")).toBe("false")
  })

  it("clicar no Label que envolve o controle alterna o controle", () => {
    render(
      <Label>
        <Checkbox />
        Lembrar neste aparelho
      </Label>
    )
    const box = screen.getByRole("checkbox", { name: /Lembrar neste aparelho/ })
    fireEvent.click(screen.getByText("Lembrar neste aparelho"))
    expect(box.getAttribute("aria-checked")).toBe("true")
  })

  it("o rótulo que envolve o controle tem a altura do alvo, que sobe a 44 px em toque", () => {
    render(
      <>
        <Field>
          <FieldLabel>
            <Checkbox />
            Aceito os termos
          </FieldLabel>
        </Field>
        <Label>
          <Checkbox />
          Lembrar neste aparelho
        </Label>
      </>
    )
    expect(classes(screen.getByText("Aceito os termos"))).toContain("has-data-[slot=checkbox]:min-h-control-sm")
    expect(classes(screen.getByText("Lembrar neste aparelho"))).toContain(
      "has-data-[slot=checkbox]:min-h-control-sm"
    )
  })

  it("alterna pelo teclado e respeita o estado desabilitado", () => {
    render(
      <>
        <Checkbox aria-label="Ativo" />
        <Checkbox aria-label="Bloqueado" disabled />
      </>
    )
    const box = screen.getByRole("checkbox", { name: "Ativo" })
    box.focus()
    expect(document.activeElement).toBe(box)
    fireEvent.keyDown(box, { key: " " })
    fireEvent.keyUp(box, { key: " " })
    expect(box.getAttribute("aria-checked")).toBe("true")
    const blocked = screen.getByRole("checkbox", { name: "Bloqueado" })
    fireEvent.click(blocked)
    expect(blocked.getAttribute("aria-checked")).toBe("false")
    expect(blocked.getAttribute("aria-disabled")).toBe("true")
  })

  it("a caixa tem 16 px, canto pequeno, marcado em primary e foco por anel", () => {
    render(<Checkbox aria-label="Ativo" />)
    expect(classes(screen.getByRole("checkbox", { name: "Ativo" }))).toEqual(
      expect.arrayContaining([
        "size-4",
        "rounded-sm",
        "data-checked:bg-primary",
        "focus-visible:ring-2",
        "focus-visible:ring-ring",
        "focus-visible:ring-offset-2",
      ])
    )
  })
})
