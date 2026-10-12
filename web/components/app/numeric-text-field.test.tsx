import { fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { NumericTextField, type ParseResult } from "./numeric-text-field";

// Base dos campos de dinheiro e quantidade (FND-05 AC4-5, DSC-06 AC4): rótulo
// associado, erro junto do campo e modo de entrada.

const MESSAGE = "Informe um número par.";

function parseEven(text: string): ParseResult {
  if (text === "") return { value: null, error: null };
  const value = Number(text);
  return Number.isInteger(value) && value % 2 === 0 ? { value, error: null } : { value: null, error: MESSAGE };
}

type HarnessProps = {
  onValue?: (value: number | null) => void;
  initial?: number | null;
  error?: string;
  inputMode?: "decimal" | "numeric" | "text";
  prefix?: string;
  description?: string;
};

function Harness({ onValue, initial = null, inputMode = "numeric", ...rest }: HarnessProps) {
  const [value, setValue] = useState<number | null>(initial);
  return (
    <>
      <NumericTextField
        label="Peso"
        value={value}
        onChange={(next) => {
          setValue(next);
          onValue?.(next);
        }}
        parse={parseEven}
        format={String}
        inputMode={inputMode}
        {...rest}
      />
      <button type="button" onClick={() => setValue(42)}>
        definir 42
      </button>
    </>
  );
}

function input() {
  return screen.getByLabelText("Peso") as HTMLInputElement;
}

describe("rótulo e modo de entrada", () => {
  it("o rótulo visível fica acima do campo e associado a ele", () => {
    render(<Harness />);
    const label = screen.getByText("Peso");
    expect(input().tagName).toBe("INPUT");
    expect(label.compareDocumentPosition(input()) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it.each(["decimal", "numeric", "text"] as const)("declara o modo de entrada %s em campo de texto", (inputMode) => {
    render(<Harness inputMode={inputMode} />);
    expect(input().getAttribute("inputmode")).toBe(inputMode);
    expect(input().getAttribute("type")).toBe("text");
    expect(input().getAttribute("autocomplete")).toBe("off");
  });

  it("mostra o prefixo e a descrição junto do campo", () => {
    render(<Harness prefix="kg" description="Peso em quilos inteiros." />);
    expect(screen.getByText("kg")).toBeTruthy();
    expect(screen.getByText("Peso em quilos inteiros.")).toBeTruthy();
  });
});

describe("valor", () => {
  it("entrega o inteiro lido e mantém o texto digitado", () => {
    const onValue = vi.fn();
    render(<Harness onValue={onValue} />);
    fireEvent.change(input(), { target: { value: "08" } });
    expect(onValue).toHaveBeenLastCalledWith(8);
    expect(input().value).toBe("08");
  });

  it("mostra o valor inicial formatado e acompanha o valor trocado por fora", () => {
    render(<Harness initial={10} />);
    expect(input().value).toBe("10");
    fireEvent.click(screen.getByRole("button", { name: "definir 42" }));
    expect(input().value).toBe("42");
  });

  it("vazio entrega null, sem erro", () => {
    const onValue = vi.fn();
    render(<Harness onValue={onValue} initial={10} />);
    fireEvent.change(input(), { target: { value: "" } });
    fireEvent.blur(input());
    expect(onValue).toHaveBeenLastCalledWith(null);
    expect(screen.queryByRole("alert")).toBeNull();
  });
});

describe("erro", () => {
  it("o erro de formato só aparece depois de sair do campo", () => {
    const onValue = vi.fn();
    render(<Harness onValue={onValue} />);
    fireEvent.change(input(), { target: { value: "3" } });
    expect(onValue).toHaveBeenLastCalledWith(null);
    expect(screen.queryByRole("alert")).toBeNull();
    expect(input().getAttribute("aria-invalid")).not.toBe("true");

    fireEvent.blur(input());
    const alert = screen.getByRole("alert");
    expect(alert.textContent).toBe(MESSAGE);
    expect(input().getAttribute("aria-invalid")).toBe("true");
    // Erro com ícone e ligado ao campo.
    expect(alert.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true");
    expect(input().getAttribute("aria-describedby")?.split(" ")).toContain(alert.id);
  });

  it("mostra o erro vindo do formulário, ligado ao campo", () => {
    render(<Harness error="Informe o peso." />);
    const alert = screen.getByRole("alert");
    expect(alert.textContent).toBe("Informe o peso.");
    expect(input().getAttribute("aria-invalid")).toBe("true");
    expect(input().getAttribute("aria-describedby")?.split(" ")).toContain(alert.id);
  });

  it("o erro de formato tem precedência sobre o do formulário", () => {
    render(<Harness error="Informe o peso." />);
    fireEvent.change(input(), { target: { value: "3" } });
    fireEvent.blur(input());
    expect(screen.getByRole("alert").textContent).toBe(MESSAGE);
  });

  it("sem erro, o campo não é marcado como inválido", () => {
    render(<Harness initial={10} />);
    fireEvent.blur(input());
    expect(screen.queryByRole("alert")).toBeNull();
    expect(input().getAttribute("aria-invalid")).not.toBe("true");
  });
});
