import { screen } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext } from "@/test/msw/fixtures";

import ContaSenhaPage from "./page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

// ACS-01: /conta/senha monta o formulário de troca de senha.
it("monta o formulário de troca de senha", async () => {
  setLocation("/conta/senha");
  mockMe(authContext());
  renderWithSession(<ContaSenhaPage />);
  expect(await screen.findByRole("heading", { name: "Trocar senha" })).toBeTruthy();
  expect(screen.getByLabelText("Senha atual")).toBeTruthy();
  expect(screen.getByLabelText("Nova senha")).toBeTruthy();
  expect(screen.getByLabelText("Confirmar nova senha")).toBeTruthy();
  expect(screen.getByRole("button", { name: "Trocar senha" })).toBeTruthy();
});
