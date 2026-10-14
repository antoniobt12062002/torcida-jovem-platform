import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { afterEach, describe, expect, it } from "vitest";

import { FormErrorAlert } from "./form-error-alert";

afterEach(cleanup);

// N2 (smoke T13): o erro que não pertence a campo recebe o foco ao surgir,
// porque os controles ficam desabilitados durante o envio.
describe("FormErrorAlert", () => {
  function Harness() {
    const [error, setError] = useState<string | null>(null);
    const [text, setText] = useState("");
    return (
      <div>
        <input aria-label="Campo" value={text} onChange={(event) => setText(event.target.value)} />
        <button type="button" onClick={() => setError("Saldo insuficiente.")}>
          Enviar
        </button>
        {error ? <FormErrorAlert>{error}</FormErrorAlert> : null}
      </div>
    );
  }

  it("é um alerta de erro focável por script e recebe o foco ao surgir", () => {
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Enviar" }));
    const alert = screen.getByRole("alert");
    expect(alert.textContent).toContain("Saldo insuficiente.");
    expect(alert.getAttribute("tabindex")).toBe("-1");
    expect(alert.className).toContain("bg-destructive-subtle");
    expect(document.activeElement).toBe(alert);
  });

  it("não toma o foco de volta a cada renderização", () => {
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Enviar" }));
    const input = screen.getByRole("textbox", { name: "Campo" });
    input.focus();
    fireEvent.change(input, { target: { value: "a" } });
    expect(document.activeElement).toBe(input);
  });
});
