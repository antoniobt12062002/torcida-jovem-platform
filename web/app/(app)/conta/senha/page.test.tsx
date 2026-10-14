import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext } from "@/test/msw/fixtures";

import AppLayout from "../../layout";
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

// DSS-03: na troca obrigatória, /conta/senha usa o layout de acesso, sem o
// shell, com "Sair" no canto; fora dela, é página estreita do shell.
describe("contêiner de /conta/senha", () => {
  function renderInGate() {
    setLocation("/conta/senha");
    renderWithSession(
      <AppLayout>
        <ContaSenhaPage />
      </AppLayout>,
    );
  }

  it("na troca obrigatória: layout de acesso com a marca e Sair, sem navegação do shell", async () => {
    mockMe(authContext({ must_change_password: true }));
    renderInGate();
    await screen.findByRole("heading", { level: 1, name: "Trocar senha" });
    expect(document.querySelector("[data-slot='access-layout'] [data-slot='brand-mark']")).not.toBeNull();
    expect(screen.queryByRole("navigation")).toBeNull();
    expect(document.querySelectorAll("header:not(main header)")).toHaveLength(0);
    expect(screen.getAllByRole("main")).toHaveLength(1);
    expect(screen.getByRole("button", { name: "Sair" })).toBeTruthy();
    expect(screen.getByLabelText("Senha atual").getAttribute("autocomplete")).toBe("current-password");
    expect(screen.getByLabelText("Nova senha").getAttribute("autocomplete")).toBe("new-password");
  });

  it("fora da troca obrigatória: página estreita dentro do shell", async () => {
    mockMe(authContext());
    renderInGate();
    await screen.findByRole("heading", { level: 1, name: "Trocar senha" });
    expect(screen.getByRole("navigation", { name: "Navegação principal" })).toBeTruthy();
    expect(document.querySelector("[data-slot='access-layout']")).toBeNull();
    const mains = screen.getAllByRole("main");
    expect(mains).toHaveLength(1);
    expect(mains[0].className).toContain("max-w-content-narrow");
  });

  it("enviar vazio leva o foco à senha atual", async () => {
    mockMe(authContext());
    renderInGate();
    fireEvent.click(await screen.findByRole("button", { name: "Trocar senha" }));
    await waitFor(() => expect(document.activeElement).toBe(screen.getByLabelText("Senha atual")));
  });
});
