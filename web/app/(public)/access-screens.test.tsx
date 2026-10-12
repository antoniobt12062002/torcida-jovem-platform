import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";

import EntrarPage from "./entrar/page";
import PublicLayout from "./layout";
import RecuperarAcessoPage from "./recuperar-acesso/page";
import RedefinirSenhaPage from "./redefinir-senha/page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

// DSS-03 e DSP-04 aplicados às telas públicas de acesso, montadas dentro do
// layout público real.

function renderPublic(path: string, page: React.ReactNode) {
  setLocation(path);
  mockMe(null);
  return renderWithSession(<PublicLayout>{page}</PublicLayout>);
}

function expectBrand() {
  // DSS-03 AC1, WDS-D-020: escudo e "Torcida Jovem"; "TJ Platform" não é marca.
  expect(document.querySelector("[data-slot='access-layout'] [data-slot='brand-mark']")).not.toBeNull();
  expect(screen.getAllByText("Torcida Jovem").length).toBeGreaterThan(0);
  expect(document.body.textContent).not.toContain("TJ Platform");
  expect(document.body.textContent).not.toContain("Desde 2013");
  expect(screen.getAllByRole("main")).toHaveLength(1);
  expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
}

function expectEmailField(input: HTMLElement) {
  // DSS-03 AC4: autocomplete de autenticação e sem correção ortográfica.
  expect(input.getAttribute("type")).toBe("email");
  expect(input.getAttribute("autocomplete")).toBe("username");
  expect(input.getAttribute("spellcheck")).toBe("false");
  expect(input.getAttribute("autocapitalize")).toBe("none");
}

function expectFullWidthSubmit(name: string) {
  // DSS-03 AC6: abaixo de 768 px, botão primário em largura total, no rodapé do
  // FormLayout. A partir de 768 px vale o alinhamento do rodapé do FormLayout
  // (igual ao espécime aprovado).
  const button = screen.getByRole("button", { name });
  expect(button.className).toMatch(/(?:^|\s)w-full(?:\s|$)/);
  expect(button.closest("[data-slot='form-layout']")).not.toBeNull();
}

describe("/entrar", () => {
  it("tem a marca, o e-mail sem correção ortográfica e a senha com autocomplete", async () => {
    renderPublic("/entrar", <EntrarPage />);
    await screen.findByRole("button", { name: "Entrar" });
    expectBrand();
    expect(screen.getByRole("heading", { level: 1, name: "Entrar" })).toBeTruthy();
    expectEmailField(screen.getByLabelText("E-mail"));
    expect(screen.getByLabelText("Senha").getAttribute("autocomplete")).toBe("current-password");
    expectFullWidthSubmit("Entrar");
  });

  it("enviar vazio leva o foco ao primeiro campo inválido", async () => {
    renderPublic("/entrar", <EntrarPage />);
    fireEvent.click(await screen.findByRole("button", { name: "Entrar" }));
    await waitFor(() => expect(document.activeElement).toBe(screen.getByLabelText("E-mail")));
  });
});

describe("/recuperar-acesso", () => {
  it("tem a marca e o e-mail sem correção ortográfica", async () => {
    renderPublic("/recuperar-acesso", <RecuperarAcessoPage />);
    await screen.findByRole("button", { name: "Enviar link" });
    expectBrand();
    expectEmailField(screen.getByLabelText("E-mail"));
    expectFullWidthSubmit("Enviar link");
  });

  it("enviar vazio leva o foco ao e-mail", async () => {
    renderPublic("/recuperar-acesso", <RecuperarAcessoPage />);
    fireEvent.click(await screen.findByRole("button", { name: "Enviar link" }));
    await waitFor(() => expect(document.activeElement).toBe(screen.getByLabelText("E-mail")));
  });
});

describe("/redefinir-senha", () => {
  it("tem a marca e as senhas com autocomplete de nova senha", async () => {
    renderPublic("/redefinir-senha#token=tok-layout-123", <RedefinirSenhaPage />);
    await screen.findByRole("button", { name: "Redefinir senha" });
    expectBrand();
    expect(screen.getByLabelText("Nova senha").getAttribute("autocomplete")).toBe("new-password");
    expect(screen.getByLabelText("Confirmar nova senha").getAttribute("autocomplete")).toBe("new-password");
    expectFullWidthSubmit("Redefinir senha");
  });

  it("enviar vazio leva o foco à nova senha", async () => {
    renderPublic("/redefinir-senha#token=tok-layout-123", <RedefinirSenhaPage />);
    fireEvent.click(await screen.findByRole("button", { name: "Redefinir senha" }));
    await waitFor(() => expect(document.activeElement).toBe(screen.getByLabelText("Nova senha")));
  });
});
