import { screen } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";

import RedefinirSenhaPage from "./page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

// ACS-03: /redefinir-senha monta o formulário de redefinição a partir do link.
it("monta o formulário de redefinição com o token do link", async () => {
  setLocation("/redefinir-senha#token=tok-pagina-123");
  mockMe(null);
  renderWithSession(<RedefinirSenhaPage />);
  expect(await screen.findByRole("heading", { name: "Redefinir senha" })).toBeTruthy();
  expect(screen.getByLabelText("Nova senha")).toBeTruthy();
  expect(screen.getByLabelText("Confirmar nova senha")).toBeTruthy();
  expect(screen.getByRole("button", { name: "Redefinir senha" })).toBeTruthy();
});
