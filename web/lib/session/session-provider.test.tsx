import { useQuery } from "@tanstack/react-query";
import { act, fireEvent, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { useState } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, unwrap } from "@/lib/api/client";
import { ApiError } from "@/lib/api/problem";
import { authContext, conta, problem } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";
import { createTestQueryClient } from "@/test/render";

import { navigations, resetNavigation, setLocation } from "./navigation-mock";
import { useCan, usePermission, useSession } from "./session-provider";
import { apiUrl, mockMe, renderWithSession } from "./test-utils";

vi.mock("next/navigation", async () => (await import("./navigation-mock")).nextNavigationMock);

beforeEach(() => {
  resetNavigation();
  window.localStorage.clear();
  window.sessionStorage.clear();
});

function Probe() {
  const session = useSession();
  const canReadContas = usePermission("financeiro:conta:read");
  const canManage = useCan("financeiro:conta:read", "financeiro:conta:create");
  const [loginError, setLoginError] = useState<string>("");
  return (
    <div>
      <span data-testid="status">{session.status}</span>
      <span data-testid="name">{session.context?.user.name ?? ""}</span>
      <span data-testid="can-read">{String(canReadContas)}</span>
      <span data-testid="can-manage">{String(canManage)}</span>
      <span data-testid="login-error">{loginError}</span>
      <button
        type="button"
        onClick={() =>
          session.login("pessoa.teste@example.com", "senha-de-teste").catch((e: unknown) => {
            setLoginError(e instanceof ApiError ? e.code : "erro");
          })
        }
      >
        entrar
      </button>
      <button type="button" onClick={() => void session.logout()}>
        sair
      </button>
    </div>
  );
}

function status() {
  return screen.getByTestId("status").textContent;
}

describe("SessionProvider: estado da sessão", () => {
  it("fica carregando e passa a autenticado com o contexto de /auth/me", async () => {
    mockMe(authContext({ user: { id: "u1", email: "a@example.com", name: "Ana Teste" } }));
    renderWithSession(<Probe />);
    expect(status()).toBe("loading");
    await waitFor(() => expect(status()).toBe("authenticated"));
    expect(screen.getByTestId("name").textContent).toBe("Ana Teste");
  });

  it("fica anônimo sem sessão e não redireciona sozinho (sem laço em /entrar)", async () => {
    setLocation("/entrar");
    mockMe(null);
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("anonymous"));
    expect(navigations()).toEqual([]);
  });
});

// FND-04 AC3: a interface decide pelas permissões efetivas, nunca pelo papel.
describe("usePermission e useCan", () => {
  it("negam com papel administrativo mas sem a permissão", async () => {
    mockMe(authContext({ roles: ["ADMIN_SISTEMA", "PRESIDENTE"], permissions: [] }));
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("authenticated"));
    expect(screen.getByTestId("can-read").textContent).toBe("false");
    expect(screen.getByTestId("can-manage").textContent).toBe("false");
  });

  it("permitem com a permissão efetiva mesmo sem papel", async () => {
    mockMe(authContext({ roles: [], permissions: ["financeiro:conta:read"] }));
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("authenticated"));
    expect(screen.getByTestId("can-read").textContent).toBe("true");
    // useCan exige todas as permissões pedidas.
    expect(screen.getByTestId("can-manage").textContent).toBe("false");
  });

  it("useCan permite quando todas estão presentes", async () => {
    mockMe(authContext({ permissions: ["financeiro:conta:read", "financeiro:conta:create"] }));
    renderWithSession(<Probe />);
    await waitFor(() => expect(screen.getByTestId("can-manage").textContent).toBe("true"));
  });

  it("negam tudo sem sessão", async () => {
    mockMe(null);
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("anonymous"));
    expect(screen.getByTestId("can-read").textContent).toBe("false");
  });
});

describe("login", () => {
  it("guarda o contexto devolvido e usa o token CSRF dele nas escritas", async () => {
    mockMe(null);
    const logged = authContext({
      user: { id: "u2", email: "b@example.com", name: "Bia Teste" },
      csrf_token: "token-do-login",
    });
    let loginBody: unknown;
    server.use(
      http.post(apiUrl("/api/v1/auth/login"), async ({ request }) => {
        loginBody = await request.json();
        return HttpResponse.json(logged);
      }),
    );
    const csrf: (string | null)[] = [];
    server.use(
      http.post(apiUrl("/api/v1/financeiro/contas"), ({ request }) => {
        csrf.push(request.headers.get("X-CSRF-Token"));
        return HttpResponse.json(conta(), { status: 201 });
      }),
    );

    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("anonymous"));
    fireEvent.click(screen.getByRole("button", { name: "entrar" }));
    await waitFor(() => expect(status()).toBe("authenticated"));
    expect(screen.getByTestId("name").textContent).toBe("Bia Teste");
    expect(loginBody).toEqual({ email: "pessoa.teste@example.com", password: "senha-de-teste" });

    await api.financeiro.POST("/api/v1/financeiro/contas", { body: { tipo: "RECEITA", nome: "X" } });
    expect(csrf).toEqual(["token-do-login"]);
  });

  it("lança ApiError com o code quando a API recusa", async () => {
    mockMe(null);
    server.use(http.post(apiUrl("/api/v1/auth/login"), () => problem(401, "invalid_credentials")));
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("anonymous"));
    fireEvent.click(screen.getByRole("button", { name: "entrar" }));
    await waitFor(() =>
      expect(screen.getByTestId("login-error").textContent).toBe("invalid_credentials"),
    );
    expect(status()).toBe("anonymous");
    // 401 do login não é sessão encerrada: nada de redirecionamento.
    expect(navigations()).toEqual([]);
  });
});

// FND-03 AC9.
describe("token CSRF e contexto só em memória", () => {
  it("não grava nada em localStorage, sessionStorage nem cookie", async () => {
    const before = document.cookie;
    mockMe(null);
    server.use(
      http.post(apiUrl("/api/v1/auth/login"), () =>
        HttpResponse.json(authContext({ csrf_token: "token-secreto" })),
      ),
    );
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("anonymous"));
    fireEvent.click(screen.getByRole("button", { name: "entrar" }));
    await waitFor(() => expect(status()).toBe("authenticated"));

    expect(window.localStorage.length).toBe(0);
    expect(window.sessionStorage.length).toBe(0);
    expect(document.cookie).toBe(before);
    expect(document.cookie).not.toContain("token-secreto");
  });

  it("usa o token de /auth/me nas escritas", async () => {
    mockMe(authContext({ csrf_token: "token-do-me" }));
    const csrf: (string | null)[] = [];
    server.use(
      http.post(apiUrl("/api/v1/financeiro/contas"), ({ request }) => {
        csrf.push(request.headers.get("X-CSRF-Token"));
        return HttpResponse.json(conta(), { status: 201 });
      }),
    );
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("authenticated"));
    await api.financeiro.POST("/api/v1/financeiro/contas", { body: { tipo: "RECEITA", nome: "X" } });
    expect(csrf).toEqual(["token-do-me"]);
  });
});

// FND-03 AC5.
describe("logout", () => {
  it("chama POST /auth/logout, limpa todo o cache e leva a /entrar", async () => {
    setLocation("/financeiro/contas");
    mockMe(authContext({ csrf_token: "token-logout" }));
    const logoutCsrf: (string | null)[] = [];
    server.use(
      http.post(apiUrl("/api/v1/auth/logout"), ({ request }) => {
        logoutCsrf.push(request.headers.get("X-CSRF-Token"));
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const queryClient = createTestQueryClient();
    queryClient.setQueryData(["financeiro", "contas"], [conta()]);
    queryClient.setQueryData(["estoque", "produtos"], []);

    renderWithSession(<Probe />, { queryClient });
    await waitFor(() => expect(status()).toBe("authenticated"));
    fireEvent.click(screen.getByRole("button", { name: "sair" }));

    await waitFor(() => expect(navigations()).toEqual(["/entrar"]));
    expect(logoutCsrf).toEqual(["token-logout"]);
    expect(status()).toBe("anonymous");
    expect(screen.getByTestId("name").textContent).toBe("");
    expect(queryClient.getQueryData(["financeiro", "contas"])).toBeUndefined();
    expect(queryClient.getQueryData(["estoque", "produtos"])).toBeUndefined();
    const withData = queryClient
      .getQueryCache()
      .findAll()
      .filter((q) => q.state.data != null);
    expect(withData).toEqual([]);
  });

  it("com a sessão já expirada no servidor (401), também limpa e leva a /entrar", async () => {
    mockMe(authContext());
    server.use(http.post(apiUrl("/api/v1/auth/logout"), () => problem(401, "session_expired")));
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("authenticated"));
    fireEvent.click(screen.getByRole("button", { name: "sair" }));
    await waitFor(() => expect(navigations()).toEqual(["/entrar"]));
    expect(status()).toBe("anonymous");
  });

  it("se a API não encerra a sessão (503), avisa e mantém a pessoa autenticada", async () => {
    mockMe(authContext());
    server.use(http.post(apiUrl("/api/v1/auth/logout"), () => problem(503, "service_unavailable")));
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("authenticated"));
    fireEvent.click(screen.getByRole("button", { name: "sair" }));
    expect(await screen.findByText("Não foi possível sair. Tente de novo.")).toBeTruthy();
    expect(status()).toBe("authenticated");
    expect(navigations()).toEqual([]);
  });
});

function Contas() {
  const query = useQuery({
    queryKey: ["financeiro", "contas"],
    queryFn: () => unwrap(api.financeiro.GET("/api/v1/financeiro/contas")),
  });
  return <span data-testid="contas">{query.status}</span>;
}

// FND-03 AC6 e caso de borda dos 401 simultâneos.
describe("401 numa chamada autenticada", () => {
  it("limpa cache e contexto e leva a /entrar com aviso e o caminho atual em next", async () => {
    setLocation("/financeiro/contas?tipo=RECEITA");
    mockMe(authContext());
    const queryClient = createTestQueryClient();
    queryClient.setQueryData(["estoque", "produtos"], []);
    renderWithSession(<Probe />, { queryClient });
    await waitFor(() => expect(status()).toBe("authenticated"));

    server.use(http.get(apiUrl("/api/v1/financeiro/contas"), () => problem(401, "session_expired")));
    await act(async () => {
      await api.financeiro.GET("/api/v1/financeiro/contas");
    });

    expect(navigations()).toHaveLength(1);
    const url = new URL(navigations()[0], window.location.origin);
    expect(url.pathname).toBe("/entrar");
    expect(url.searchParams.get("next")).toBe("/financeiro/contas?tipo=RECEITA");
    expect(url.searchParams.get("sessao")).toBe("encerrada");
    await waitFor(() => expect(status()).toBe("anonymous"));
    expect(queryClient.getQueryData(["estoque", "produtos"])).toBeUndefined();
  });

  it("dois 401 quase ao mesmo tempo produzem um único redirecionamento", async () => {
    setLocation("/financeiro/contas");
    mockMe(authContext());
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("authenticated"));

    server.use(
      http.get(apiUrl("/api/v1/financeiro/contas"), () => problem(401, "session_expired")),
      http.get(apiUrl("/api/v1/estoque/produtos"), () => problem(401, "unauthenticated")),
    );
    await act(async () => {
      await Promise.all([
        api.financeiro.GET("/api/v1/financeiro/contas"),
        api.estoque.GET("/api/v1/estoque/produtos"),
      ]);
    });

    expect(navigations()).toHaveLength(1);
    expect(new URL(navigations()[0], window.location.origin).pathname).toBe("/entrar");
  });

  it("consultas da tela não ficam em laço depois do 401", async () => {
    setLocation("/financeiro/contas");
    mockMe(authContext());
    let contasCalls = 0;
    server.use(
      http.get(apiUrl("/api/v1/financeiro/contas"), () => {
        contasCalls += 1;
        return problem(401, "session_expired");
      }),
    );
    function Screen() {
      const { status: s } = useSession();
      return s === "authenticated" ? <Contas /> : null;
    }
    renderWithSession(<Screen />);
    await waitFor(() => expect(navigations()).toHaveLength(1));
    await new Promise((r) => setTimeout(r, 50));
    expect(contasCalls).toBe(1);
    expect(navigations()).toHaveLength(1);
  });
});

describe("redirectToLogin", () => {
  it("leva a /entrar com o caminho atual em next, uma única vez", async () => {
    setLocation("/estoque/produtos?ativo=true");
    mockMe(null);
    function Gate() {
      const session = useSession();
      return (
        <button
          type="button"
          onClick={() => {
            session.redirectToLogin();
            session.redirectToLogin();
          }}
        >
          ir
        </button>
      );
    }
    renderWithSession(<Gate />);
    fireEvent.click(screen.getByRole("button", { name: "ir" }));
    fireEvent.click(screen.getByRole("button", { name: "ir" }));
    expect(navigations()).toHaveLength(1);
    const url = new URL(navigations()[0], window.location.origin);
    expect(url.pathname).toBe("/entrar");
    expect(url.searchParams.get("next")).toBe("/estoque/produtos?ativo=true");
    expect(url.searchParams.get("sessao")).toBeNull();
  });
});

// FND-03 AC7.
describe("403 password_change_required", () => {
  it("leva a /conta/senha", async () => {
    setLocation("/financeiro/contas");
    mockMe(authContext());
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("authenticated"));
    server.use(
      http.get(apiUrl("/api/v1/financeiro/contas"), () => problem(403, "password_change_required")),
    );
    await act(async () => {
      await api.financeiro.GET("/api/v1/financeiro/contas");
    });
    expect(navigations()).toEqual(["/conta/senha"]);
  });
});

// FND-03 AC8.
describe("403 csrf_invalid", () => {
  it("rebusca /auth/me, pede para repetir e não repete a escrita", async () => {
    const me = mockMe(authContext());
    let writes = 0;
    server.use(
      http.post(apiUrl("/api/v1/financeiro/contas"), () => {
        writes += 1;
        return problem(403, "csrf_invalid");
      }),
    );
    renderWithSession(<Probe />);
    await waitFor(() => expect(status()).toBe("authenticated"));
    expect(me.count).toBe(1);

    await act(async () => {
      await api.financeiro.POST("/api/v1/financeiro/contas", { body: { tipo: "RECEITA", nome: "X" } });
    });

    await waitFor(() => expect(me.count).toBe(2));
    expect(
      await screen.findByText("A sessão foi atualizada. Repita a ação, por favor."),
    ).toBeTruthy();
    await new Promise((r) => setTimeout(r, 50));
    expect(writes).toBe(1);
    expect(navigations()).toEqual([]);
  });
});
