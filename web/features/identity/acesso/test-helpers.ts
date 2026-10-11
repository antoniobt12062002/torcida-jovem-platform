import { fireEvent, screen } from "@testing-library/react";

// Ajudantes dos testes da feature de acesso.

/** Textos ligados ao campo por aria-describedby (dica e erro do campo). */
export function describedByText(el: HTMLElement): string[] {
  const ids = (el.getAttribute("aria-describedby") ?? "").split(" ").filter(Boolean);
  return ids.map((id) => document.getElementById(id)?.textContent ?? "");
}

export function isInvalid(el: HTMLElement): boolean {
  return el.getAttribute("aria-invalid") === "true";
}

export function type(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}
