import { fireEvent, render, screen, within } from "@testing-library/react";
import { type ReactNode, useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import { FormLayout } from "./form-layout";

function Fields() {
  return (
    <Field>
      <FieldLabel>Descrição</FieldLabel>
      <Input name="descricao" />
    </Field>
  );
}

function footer(extra?: ReactNode) {
  return (
    <>
      <Button type="button" variant="secondary">
        Cancelar
      </Button>
      {extra ?? <Button type="submit">Salvar</Button>}
    </>
  );
}

describe("FormLayout", () => {
  // DSP-04 AC1.
  it("é um formulário em coluna única, na largura estreita, com o rótulo acima do campo", () => {
    render(
      <FormLayout aria-label="Novo lançamento" footer={footer()}>
        <Fields />
      </FormLayout>,
    );
    const form = screen.getByRole("form", { name: "Novo lançamento" });
    expect(form.tagName).toBe("FORM");
    expect(form.className.split(/\s+/)).toEqual(expect.arrayContaining(["flex", "flex-col", "max-w-content-narrow"]));
    const label = screen.getByText("Descrição");
    const input = screen.getByLabelText("Descrição");
    expect(label.compareDocumentPosition(input) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(input.closest("[data-slot=field]")?.getAttribute("data-orientation")).toBe("vertical");
  });

  it("repassa o envio a quem usa", () => {
    const onSubmit = vi.fn((event: React.FormEvent) => event.preventDefault());
    render(
      <FormLayout aria-label="Novo lançamento" onSubmit={onSubmit} footer={footer()}>
        <Fields />
      </FormLayout>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Salvar" }));
    expect(onSubmit).toHaveBeenCalledTimes(1);
  });

  // DSP-04 AC4.
  it("sem erro de formulário, não há alerta", () => {
    render(
      <FormLayout aria-label="Novo lançamento" footer={footer()}>
        <Fields />
      </FormLayout>,
    );
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("mostra o erro sem campo como alerta no topo do formulário", () => {
    render(
      <FormLayout aria-label="Novo lançamento" error="A conta está inativa." footer={footer()}>
        <Fields />
      </FormLayout>,
    );
    const form = screen.getByRole("form");
    const alert = within(form).getByRole("alert");
    expect(alert.textContent).toContain("A conta está inativa.");
    expect(alert.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true");
    // No topo: antes do primeiro campo.
    expect(
      alert.compareDocumentPosition(screen.getByLabelText("Descrição")) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  it("leva o foco ao alerta quando o erro surge, e não o rouba enquanto a pessoa digita", () => {
    function Harness() {
      const [error, setError] = useState<string>();
      const [, setText] = useState("");
      return (
        <FormLayout
          aria-label="Novo lançamento"
          error={error ? <span>{error}</span> : undefined}
          onSubmit={(event) => {
            event.preventDefault();
            setError("A conta está inativa.");
          }}
          footer={footer()}
        >
          <Field>
            <FieldLabel>Descrição</FieldLabel>
            <Input onChange={(event) => setText(event.target.value)} />
          </Field>
        </FormLayout>
      );
    }
    render(<Harness />);
    const input = screen.getByLabelText("Descrição");
    input.focus();
    fireEvent.click(screen.getByRole("button", { name: "Salvar" }));
    expect(document.activeElement).toBe(screen.getByRole("alert"));

    // Com o erro ainda exibido, digitar renderiza de novo e o foco fica no campo.
    input.focus();
    fireEvent.change(input, { target: { value: "mensalidade" } });
    expect(document.activeElement).toBe(input);
  });

  // DSP-04 AC9.
  it("rodapé: ações em largura total abaixo de 768 px e à direita a partir daí, com a principal por último", () => {
    const { container } = render(
      <FormLayout aria-label="Novo lançamento" footer={footer()}>
        <Fields />
      </FormLayout>,
    );
    const actions = container.querySelector("[data-slot=form-footer]") as HTMLElement;
    expect(actions.className.split(/\s+/)).toEqual(
      expect.arrayContaining(["flex-col", "*:w-full", "md:flex-row", "md:justify-end", "md:*:w-auto"]),
    );
    expect(within(actions).getAllByRole("button").map((button) => button.textContent)).toEqual(["Cancelar", "Salvar"]);
    // O rodapé vem depois dos campos.
    expect(
      screen.getByLabelText("Descrição").compareDocumentPosition(actions) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  // DSP-04 AC5.
  it("durante o envio, a ação de enviar fica desabilitada com o rótulo de progresso", () => {
    render(
      <FormLayout
        aria-label="Novo lançamento"
        footer={footer(
          <Button type="submit" loading loadingLabel="Salvando…">
            Salvar
          </Button>,
        )}
      >
        <Fields />
      </FormLayout>,
    );
    const submit = screen.getByRole("button", { name: "Salvando…" }) as HTMLButtonElement;
    expect(submit.disabled).toBe(true);
    expect(screen.queryByRole("button", { name: "Salvar" })).toBeNull();
  });
});
