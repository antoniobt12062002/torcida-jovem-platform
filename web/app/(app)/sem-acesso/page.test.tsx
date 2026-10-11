import { render, screen } from "@testing-library/react";
import { expect, it } from "vitest";

import SemAcessoPage from "./page";

// FND-04 AC4: a página "Sem acesso".
it("mostra Sem acesso e o caminho de volta à página inicial", () => {
  render(<SemAcessoPage />);
  expect(screen.getByRole("heading", { name: "Sem acesso" })).toBeTruthy();
  expect(screen.getByText(/não tem permissão/)).toBeTruthy();
  expect(screen.getByRole("link", { name: "Voltar ao início" }).getAttribute("href")).toBe("/inicio");
});
