import { render, screen, within } from "@testing-library/react";
import { expect, it } from "vitest";

import SemAcessoPage from "./page";

// FND-04 AC4: a página "Sem acesso".
it("mostra Sem acesso e o caminho de volta à página inicial", () => {
  render(<SemAcessoPage />);
  expect(screen.getByRole("heading", { name: "Sem acesso" })).toBeTruthy();
  expect(screen.getByText(/não tem permissão/)).toBeTruthy();
  expect(screen.getByRole("link", { name: "Voltar ao início" }).getAttribute("href")).toBe("/inicio");
});

// DSP-02 AC5, DSS-01 AC6: o estado padrão "Sem acesso" no contêiner de página,
// com uma única região principal e um único `h1`.
it("usa o contêiner de página, com uma região principal e um h1", () => {
  render(<SemAcessoPage />);
  const mains = screen.getAllByRole("main");
  expect(mains).toHaveLength(1);
  expect(mains[0].id).toBe("conteudo");
  expect(within(mains[0]).getAllByRole("heading", { level: 1 }).map((h) => h.textContent)).toEqual(["Sem acesso"]);
  expect(mains[0].querySelector("[data-slot='no-access']")).not.toBeNull();
});
