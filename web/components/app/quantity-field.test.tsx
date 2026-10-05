import { fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { QuantityField, type QuantityMode } from "./quantity-field";

// FND-05 AC5: só inteiros, com o sinal configurável (só positivos, ou
// diferente de zero com sinal).

function Harness({ mode, onValue }: { mode: QuantityMode; onValue: (v: number | null) => void }) {
  const [value, setValue] = useState<number | null>(null);
  return (
    <QuantityField
      label="Quantidade"
      mode={mode}
      value={value}
      onChange={(n) => {
        setValue(n);
        onValue(n);
      }}
    />
  );
}

function setup(mode: QuantityMode) {
  const onValue = vi.fn();
  render(<Harness mode={mode} onValue={onValue} />);
  const input = screen.getByLabelText("Quantidade") as HTMLInputElement;
  return {
    onValue,
    type(text: string) {
      fireEvent.change(input, { target: { value: text } });
      fireEvent.blur(input);
    },
    error: () => screen.queryByRole("alert")?.textContent ?? null,
  };
}

describe("modo positive", () => {
  it.each([
    ["5", 5],
    ["120", 120],
    ["+7", 7],
  ])("aceita %s como %d", (text, expected) => {
    const field = setup("positive");
    field.type(text);
    expect(field.onValue).toHaveBeenLastCalledWith(expected);
    expect(field.error()).toBeNull();
  });

  it.each(["1,5", "1.5", "2,0", "abc", "1e3"])("recusa decimal ou texto %j", (text) => {
    const field = setup("positive");
    field.type(text);
    expect(field.onValue).toHaveBeenLastCalledWith(null);
    expect(field.error()).toBe("Informe um número inteiro.");
  });

  it.each(["0", "-3"])("recusa %j", (text) => {
    const field = setup("positive");
    field.type(text);
    expect(field.onValue).toHaveBeenLastCalledWith(null);
    expect(field.error()).toBe("Informe um número maior que zero.");
  });
});

describe("modo nonZero", () => {
  it.each([
    ["-3", -3],
    ["+4", 4],
    ["10", 10],
  ])("aceita %s como %d", (text, expected) => {
    const field = setup("nonZero");
    field.type(text);
    expect(field.onValue).toHaveBeenLastCalledWith(expected);
    expect(field.error()).toBeNull();
  });

  it.each(["0", "-0"])("recusa %j", (text) => {
    const field = setup("nonZero");
    field.type(text);
    expect(field.onValue).toHaveBeenLastCalledWith(null);
    expect(field.error()).toBe("Informe um número diferente de zero.");
  });

  it.each(["-2,5", "2.5"])("recusa decimal %j", (text) => {
    const field = setup("nonZero");
    field.type(text);
    expect(field.onValue).toHaveBeenLastCalledWith(null);
    expect(field.error()).toBe("Informe um número inteiro.");
  });
});

it("recusa inteiro acima do maior inteiro seguro", () => {
  const field = setup("positive");
  field.type("9007199254740993");
  expect(field.onValue).toHaveBeenLastCalledWith(null);
  expect(field.error()).toBe("Número grande demais.");
});
