import { screen } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";

import RecuperarAcessoPage from "./page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

// ACS-02: /recuperar-acesso monta o formulário de pedido.
it("monta o formulário de pedido de recuperação", async () => {
  setLocation("/recuperar-acesso");
  mockMe(null);
  renderWithSession(<RecuperarAcessoPage />);
  expect(await screen.findByRole("heading", { name: "Recuperar acesso" })).toBeTruthy();
  expect(screen.getByLabelText("E-mail")).toBeTruthy();
  expect(screen.getByRole("button", { name: "Enviar link" })).toBeTruthy();
});
