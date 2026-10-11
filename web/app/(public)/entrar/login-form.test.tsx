import { fireEvent, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { navigations, resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, problem } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { LoginForm } from "./login-form";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

function mockLogin(respond: () => Response) {
  const bodies: unknown[] = [];
  server.use(
    http.post(apiUrl("/api/v1/auth/login"), async ({ request }) => {
      bodies.push(await request.json());
      return respond();
    }),
  );
  return bodies;
}

async function renderLogin(path = "/entrar") {
  setLocation(path);
  mockMe(null);
  renderWithSession(<LoginForm />);
  // O formulário fica disponível para quem está sem sessão.
  await screen.findByRole("button", { name: "Entrar" });
}

function fill(email: string, password: string) {
  fireEvent.change(screen.getByLabelText("E-mail"), { target: { value: email } });
  fireEvent.change(screen.getByLabelText("Senha"), { target: { value: password } });
}

function submit() {
  fireEvent.click(screen.getByRole("button", { name: "Entrar" }));
}

// FND-03 AC1 e caso de borda de `next`.
describe("login com sucesso", () => {
  it("envia e-mail e senha e leva ao destino pedido em next", async () => {
    const bodies = mockLogin(() => HttpResponse.json(authContext()));
    await renderLogin("/entrar?next=%2Ffinanceiro%2Fcontas%3Ftipo%3DRECEITA");
    fill("pessoa.teste@example.com", "senha-de-teste");
    submit();
    await waitFor(() => expect(navigations()).toEqual(["/financeiro/contas?tipo=RECEITA"]));
    expect(bodies).toEqual([{ email: "pessoa.teste@example.com", password: "senha-de-teste" }]);
  });

  it("sem next, leva à página inicial", async () => {
    mockLogin(() => HttpResponse.json(authContext()));
    await renderLogin();
    fill("pessoa.teste@example.com", "senha-de-teste");
    submit();
    await waitFor(() => expect(navigations()).toEqual(["/inicio"]));
  });

  it.each(["//exemplo.com", "https://exemplo.com/inicio", "/\\exemplo.com", "financeiro"])(
    "ignora next fora da origem (%s) e leva à página inicial",
    async (next) => {
      mockLogin(() => HttpResponse.json(authContext()));
      await renderLogin(`/entrar?next=${encodeURIComponent(next)}`);
      fill("pessoa.teste@example.com", "senha-de-teste");
      submit();
      await waitFor(() => expect(navigations()).toEqual(["/inicio"]));
    },
  );
});

// FND-03 AC4.
describe("troca de senha obrigatória", () => {
  it("leva a /conta/senha mesmo com next", async () => {
    mockLogin(() => HttpResponse.json(authContext({ must_change_password: true })));
    await renderLogin("/entrar?next=%2Ffinanceiro%2Fcontas");
    fill("pessoa.teste@example.com", "senha-temporaria");
    submit();
    await waitFor(() => expect(navigations()).toEqual(["/conta/senha"]));
  });
});

// FND-03 AC2.
describe("credenciais inválidas", () => {
  it("mostra a mensagem sem indicar qual campo está errado", async () => {
    mockLogin(() => problem(401, "invalid_credentials"));
    await renderLogin();
    fill("pessoa.teste@example.com", "errada");
    submit();
    expect(await screen.findByText("E-mail ou senha inválidos.")).toBeTruthy();
    expect(screen.getByLabelText("E-mail").getAttribute("aria-invalid")).not.toBe("true");
    expect(screen.getByLabelText("Senha").getAttribute("aria-invalid")).not.toBe("true");
    expect(navigations()).toEqual([]);
  });
});

// FND-03 AC3.
describe("acesso bloqueado", () => {
  it("informa o bloqueio e o tempo de espera a partir de Retry-After", async () => {
    mockLogin(() => problem(429, "login_blocked", {}, { "Retry-After": "900" }));
    await renderLogin();
    fill("pessoa.teste@example.com", "errada");
    submit();
    const alert = await screen.findByText(/temporariamente bloqueado/);
    expect(alert.textContent).toContain("15 minutos");
    expect(navigations()).toEqual([]);
  });

  it("sem Retry-After, informa o bloqueio sem inventar tempo", async () => {
    mockLogin(() => problem(429, "login_blocked"));
    await renderLogin();
    fill("pessoa.teste@example.com", "errada");
    submit();
    const alert = await screen.findByText(/temporariamente bloqueado/);
    expect(alert.textContent).toContain("mais tarde");
  });
});

// FND-06 aplicado ao login.
describe("outros erros", () => {
  it("API indisponível mostra a mensagem de indisponibilidade", async () => {
    mockLogin(() => problem(503, "service_unavailable"));
    await renderLogin();
    fill("pessoa.teste@example.com", "senha-de-teste");
    submit();
    expect(await screen.findByText(/serviço está indisponível/)).toBeTruthy();
  });

  it("validation_failed põe o erro junto do campo", async () => {
    mockLogin(() =>
      problem(422, "validation_failed", { errors: [{ field: "email", code: "invalid_format" }] }),
    );
    await renderLogin();
    fill("nao-e-email", "senha-de-teste");
    submit();
    await waitFor(() =>
      expect(screen.getByLabelText("E-mail").getAttribute("aria-invalid")).toBe("true"),
    );
    expect(screen.getByText("Formato inválido.")).toBeTruthy();
  });

  it("não chama a API com campos vazios", async () => {
    const bodies = mockLogin(() => HttpResponse.json(authContext()));
    await renderLogin();
    submit();
    expect(await screen.findByText("Informe o e-mail.")).toBeTruthy();
    expect(screen.getByText("Informe a senha.")).toBeTruthy();
    expect(bodies).toEqual([]);
  });
});

// FND-03 AC10.
it("oferece o link Esqueci minha senha para /recuperar-acesso", async () => {
  await renderLogin();
  const link = screen.getByRole("link", { name: "Esqueci minha senha" });
  expect(link.getAttribute("href")).toBe("/recuperar-acesso");
});

// FND-03 AC6: aviso de sessão encerrada.
it("mostra o aviso de sessão encerrada quando vem do 401", async () => {
  await renderLogin("/entrar?next=%2Ffinanceiro%2Fcontas&sessao=encerrada");
  expect(screen.getByText(/sessão foi encerrada/)).toBeTruthy();
});

it("sem o aviso, não fala em sessão encerrada", async () => {
  await renderLogin();
  expect(screen.queryByText(/sessão foi encerrada/)).toBeNull();
});

// Caso de borda: pessoa já autenticada em /entrar.
describe("pessoa já autenticada em /entrar", () => {
  it("vai para a página inicial", async () => {
    setLocation("/entrar?next=%2Ffinanceiro%2Fcontas");
    mockMe(authContext());
    renderWithSession(<LoginForm />);
    await waitFor(() => expect(navigations()).toEqual(["/inicio"]));
  });

  it("com troca obrigatória, vai para /conta/senha", async () => {
    setLocation("/entrar");
    mockMe(authContext({ must_change_password: true }));
    renderWithSession(<LoginForm />);
    await waitFor(() => expect(navigations()).toEqual(["/conta/senha"]));
  });
});
