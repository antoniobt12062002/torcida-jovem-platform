import { fireEvent, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { navigations, resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { apiUrl, mockMe, renderWithSession } from "@/lib/session/test-utils";
import { type AuthContext, authContext, problem } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { ChangePasswordForm } from "./change-password-form";
import { describedByText, isInvalid, type } from "./test-helpers";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const CURRENT = "Senha atual";
const NEW = "Nova senha";
const CONFIRM = "Confirmar nova senha";
const SUBMIT = "Trocar senha";

function mockChange(respond: () => Response) {
  const requests: { body: unknown; csrf: string | null }[] = [];
  server.use(
    http.post(apiUrl("/api/v1/auth/password"), async ({ request }) => {
      requests.push({ body: await request.json(), csrf: request.headers.get("X-CSRF-Token") });
      return respond();
    }),
  );
  return requests;
}

/** /auth/me devolve `before` até a troca e `after` depois dela. */
function mockMeSequence(before: AuthContext, after: AuthContext) {
  const state = { changed: false, calls: 0 };
  server.use(
    http.get(apiUrl("/api/v1/auth/me"), () => {
      state.calls += 1;
      return HttpResponse.json(state.changed ? after : before);
    }),
    http.post(apiUrl("/api/v1/auth/password"), () => {
      state.changed = true;
      return new HttpResponse(null, { status: 204 });
    }),
  );
  return state;
}

async function renderForm(path = "/conta/senha") {
  setLocation(path);
  renderWithSession(<ChangePasswordForm />);
  await screen.findByRole("button", { name: SUBMIT });
}

function fill(current: string, next: string, confirm = next) {
  type(CURRENT, current);
  type(NEW, next);
  type(CONFIRM, confirm);
}

function submit() {
  fireEvent.click(screen.getByRole("button", { name: SUBMIT }));
}

// ACS-01 AC1.
describe("troca voluntária com sucesso", () => {
  it("envia senha atual e nova senha e confirma o sucesso", async () => {
    mockMe(authContext());
    const requests = mockChange(() => new HttpResponse(null, { status: 204 }));
    await renderForm();
    fill("senha-antiga-1", "senha-nova-123");
    submit();
    expect(await screen.findByText("Senha alterada com sucesso.")).toBeTruthy();
    expect(requests).toEqual([
      {
        body: { current_password: "senha-antiga-1", new_password: "senha-nova-123" },
        csrf: "csrf-token-de-teste",
      },
    ]);
    // Troca voluntária: a pessoa continua na página.
    expect(navigations()).toEqual([]);
  });
});

// ACS-01 AC2.
describe("confirmação diferente da nova senha", () => {
  it("mostra o erro no campo de confirmação e não chama a API", async () => {
    mockMe(authContext());
    const requests = mockChange(() => new HttpResponse(null, { status: 204 }));
    await renderForm();
    fill("senha-antiga-1", "senha-nova-123", "senha-nova-124");
    submit();
    const message = "A confirmação não confere com a nova senha.";
    expect(await screen.findByText(message)).toBeTruthy();
    const confirm = screen.getByLabelText(CONFIRM);
    expect(isInvalid(confirm)).toBe(true);
    expect(describedByText(confirm)).toContain(message);
    expect(isInvalid(screen.getByLabelText(NEW))).toBe(false);
    expect(requests).toEqual([]);
  });
});

// ACS-01 AC3.
describe("troca obrigatória com sucesso", () => {
  it("rebusca /auth/me e leva à página inicial", async () => {
    const state = mockMeSequence(
      authContext({ must_change_password: true }),
      authContext({ must_change_password: false }),
    );
    await renderForm();
    fill("senha-temporaria", "senha-nova-123");
    submit();
    await waitFor(() => expect(navigations()).toEqual(["/inicio"]));
    // Uma consulta ao montar e outra depois da troca.
    expect(state.calls).toBe(2);
  });

  it("leva ao destino original pedido em next", async () => {
    mockMeSequence(authContext({ must_change_password: true }), authContext());
    await renderForm("/conta/senha?next=%2Ffinanceiro%2Fcontas");
    fill("senha-temporaria", "senha-nova-123");
    submit();
    await waitFor(() => expect(navigations()).toEqual(["/financeiro/contas"]));
  });

  it("ignora next fora da origem e leva à página inicial", async () => {
    mockMeSequence(authContext({ must_change_password: true }), authContext());
    await renderForm(`/conta/senha?next=${encodeURIComponent("//exemplo.com")}`);
    fill("senha-temporaria", "senha-nova-123");
    submit();
    await waitFor(() => expect(navigations()).toEqual(["/inicio"]));
  });
});

// ACS-01 AC4 e caso de borda da troca voluntária.
describe("modo obrigatório", () => {
  it("explica a obrigatoriedade, não oferece navegação e mantém Sair", async () => {
    mockMe(authContext({ must_change_password: true }));
    const logouts = { count: 0 };
    server.use(
      http.post(apiUrl("/api/v1/auth/logout"), () => {
        logouts.count += 1;
        return new HttpResponse(null, { status: 204 });
      }),
    );
    await renderForm();
    expect(screen.getByText(/precisa trocar a senha antes de continuar/)).toBeTruthy();
    expect(screen.queryAllByRole("link")).toEqual([]);
    fireEvent.click(screen.getByRole("button", { name: "Sair" }));
    await waitFor(() => expect(navigations()).toEqual(["/entrar"]));
    expect(logouts.count).toBe(1);
  });

  it("sem troca obrigatória, não mostra o aviso de obrigatoriedade", async () => {
    mockMe(authContext());
    await renderForm();
    expect(screen.queryByText(/precisa trocar a senha antes de continuar/)).toBeNull();
  });
});

// Dica de tamanho mínimo (premissa da spec, ADR-005).
describe("dica de tamanho mínimo", () => {
  it("10 caracteres para quem tem vínculo administrativo", async () => {
    mockMe(
      authContext({
        admin_membership: { reason: "Diretoria", granted_at: "2026-10-01T12:00:00Z" },
      }),
    );
    await renderForm();
    expect(describedByText(screen.getByLabelText(NEW))).toContain("Mínimo de 10 caracteres.");
  });

  it("8 caracteres para os demais", async () => {
    mockMe(authContext());
    await renderForm();
    expect(describedByText(screen.getByLabelText(NEW))).toContain("Mínimo de 8 caracteres.");
  });
});

// ACS-01 AC5.
describe("erros de senha da API no campo certo", () => {
  it.each([
    [403, "invalid_current_password", CURRENT, "A senha atual está incorreta."],
    [422, "password_unchanged", NEW, "A nova senha precisa ser diferente da atual."],
    [422, "password_too_short", NEW, "A senha é curta demais. Mínimo de 8 caracteres."],
    [422, "password_too_long", NEW, "A senha é longa demais."],
    [422, "password_compromised", NEW, "Esta senha aparece em vazamentos conhecidos. Escolha outra."],
  ])("%i %s", async (status, code, label, message) => {
    mockMe(authContext());
    mockChange(() => problem(status, code));
    await renderForm();
    fill("senha-antiga-1", "senha-nova-123");
    submit();
    expect(await screen.findByText(message)).toBeTruthy();
    const field = screen.getByLabelText(label);
    expect(isInvalid(field)).toBe(true);
    expect(describedByText(field)).toContain(message);
    for (const other of [CURRENT, NEW, CONFIRM].filter((l) => l !== label)) {
      expect(isInvalid(screen.getByLabelText(other))).toBe(false);
    }
    expect(screen.queryByText("Senha alterada com sucesso.")).toBeNull();
  });

  it("password_too_short com vínculo administrativo cita o mínimo de 10", async () => {
    mockMe(
      authContext({
        admin_membership: { reason: "Diretoria", granted_at: "2026-10-01T12:00:00Z" },
      }),
    );
    mockChange(() => problem(422, "password_too_short"));
    await renderForm();
    fill("senha-antiga-1", "curta-123");
    submit();
    const message = "A senha é curta demais. Mínimo de 10 caracteres.";
    expect(await screen.findByText(message)).toBeTruthy();
    expect(describedByText(screen.getByLabelText(NEW))).toContain(message);
  });
});

// ACS-01 AC6.
describe("troca bloqueada por excesso de tentativas", () => {
  it("informa quanto tempo esperar a partir de Retry-After", async () => {
    mockMe(authContext());
    mockChange(() => problem(429, "password_change_blocked", {}, { "Retry-After": "900" }));
    await renderForm();
    fill("errada-123", "senha-nova-123");
    submit();
    expect(
      await screen.findByText("Muitas tentativas. Tente de novo em 15 minutos."),
    ).toBeTruthy();
  });

  it("sem Retry-After, não inventa tempo", async () => {
    mockMe(authContext());
    mockChange(() => problem(429, "password_change_blocked"));
    await renderForm();
    fill("errada-123", "senha-nova-123");
    submit();
    expect(
      await screen.findByText("Muitas tentativas. Tente de novo mais tarde."),
    ).toBeTruthy();
  });
});
