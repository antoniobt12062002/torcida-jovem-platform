import { render, screen } from "@testing-library/react";
import { expect, it } from "vitest";

import { RoleBadge } from "./role-badge";

// DSC-06 AC5: o nome do papel em português, em tom neutro. O rótulo vem pronto
// da feature dona do mapa de papéis.
it("mostra o rótulo recebido, em selo de contorno neutro", () => {
  render(<RoleBadge label="Conselho fiscal" />);
  const badge = screen.getByText("Conselho fiscal");
  expect(badge.getAttribute("data-slot")).toBe("badge");
  const classes = badge.className.split(/\s+/);
  expect(classes).toEqual(expect.arrayContaining(["border-border", "text-foreground"]));
  expect(classes.some((name) => /^(bg|text)-(success|warning|destructive)/.test(name))).toBe(false);
});
