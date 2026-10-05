import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";

import vectors from "../../../api/internal/platform/money/testdata/vectors.json";
import { MoneyField } from "./money-field";

// FND-05 AC4: entrada pt-BR convertida em centavos inteiros por parseBRL,
// nunca número de ponto flutuante. Vetores de lib/money.test.ts.

function Harness({ onValue, initial = null }: { onValue: (v: number | null) => void; initial?: number | null }) {
  const [value, setValue] = useState<number | null>(initial);
  return (
    <MoneyField
      label="Valor"
      value={value}
      onChange={(cents) => {
        setValue(cents);
        onValue(cents);
      }}
    />
  );
}

function input() {
  return screen.getByLabelText("Valor") as HTMLInputElement;
}

function type(text: string) {
  fireEvent.change(input(), { target: { value: text } });
}

describe("entradas válidas (vetores parse_valid)", () => {
  it.each(vectors.parse_valid)("$text vira $cents centavos inteiros", ({ text, cents }) => {
    const onValue = vi.fn();
    render(<Harness onValue={onValue} />);
    type(text);
    const last = onValue.mock.lastCall?.[0];
    expect(last).toBe(cents);
    expect(Number.isSafeInteger(last)).toBe(true);
    fireEvent.blur(input());
    expect(screen.queryByRole("alert")).toBeNull();
    expect(input().value).toBe(text);
  });
});

describe("entradas inválidas (vetores parse_invalid)", () => {
  it.each(vectors.parse_invalid.filter((text) => text !== ""))("%j não produz valor e mostra erro", (text) => {
    const onValue = vi.fn();
    render(<Harness onValue={onValue} initial={100} />);
    type(text);
    expect(onValue).toHaveBeenLastCalledWith(null);
    fireEvent.blur(input());
    expect(screen.getByRole("alert").textContent).toBe("Valor inválido. Use o formato 1.234,56.");
    expect(input().getAttribute("aria-invalid")).toBe("true");
  });

  it("acima do maior inteiro seguro não produz valor", () => {
    const onValue = vi.fn();
    render(<Harness onValue={onValue} />);
    type("90.071.992.547.409,92");
    expect(onValue).toHaveBeenLastCalledWith(null);
    fireEvent.blur(input());
    expect(screen.getByRole("alert").textContent).toBe("Valor grande demais.");
  });

  it("campo vazio vira null, sem erro de formato", () => {
    const onValue = vi.fn();
    render(<Harness onValue={onValue} initial={100} />);
    type("");
    expect(onValue).toHaveBeenLastCalledWith(null);
    fireEvent.blur(input());
    expect(screen.queryByRole("alert")).toBeNull();
  });
});

describe("exibição", () => {
  it("mostra o valor em centavos no formato pt-BR", () => {
    render(<MoneyField label="Valor" value={123456} onChange={() => {}} />);
    expect(input().value).toBe("1.234,56");
  });

  it("mostra o erro recebido do formulário associado ao campo", () => {
    render(<MoneyField label="Valor" value={null} onChange={() => {}} error="Informe o valor." />);
    expect(screen.getByRole("alert").textContent).toBe("Informe o valor.");
    expect(input().getAttribute("aria-invalid")).toBe("true");
  });
});

describe("com React Hook Form", () => {
  function Form({ onSubmit }: { onSubmit: (data: { valor: number | null }) => void }) {
    const form = useForm<{ valor: number | null }>({ defaultValues: { valor: null } });
    return (
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <Controller
          name="valor"
          control={form.control}
          rules={{ validate: (v) => v !== null || "Informe um valor válido." }}
          render={({ field, fieldState }) => (
            <MoneyField label="Valor" {...field} error={fieldState.error?.message} />
          )}
        />
        <button type="submit">Salvar</button>
        <button type="button" onClick={() => form.reset({ valor: 500 })}>
          Restaurar
        </button>
      </form>
    );
  }

  it("envia centavos inteiros à API", async () => {
    const onSubmit = vi.fn();
    render(<Form onSubmit={onSubmit} />);
    type("1.234,56");
    fireEvent.click(screen.getByRole("button", { name: "Salvar" }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit.mock.calls[0][0]).toEqual({ valor: 123456 });
  });

  it("não envia com entrada inválida", async () => {
    const onSubmit = vi.fn();
    render(<Form onSubmit={onSubmit} />);
    type("1,234");
    fireEvent.click(screen.getByRole("button", { name: "Salvar" }));
    expect(await screen.findByRole("alert")).toBeTruthy();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("acompanha o valor quando o formulário é restaurado", async () => {
    render(<Form onSubmit={() => {}} />);
    type("10");
    fireEvent.click(screen.getByRole("button", { name: "Restaurar" }));
    await waitFor(() => expect(input().value).toBe("5,00"));
  });
});
