import { http, HttpResponse } from "msw";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { authContext, conta, ids, problem, problemBody } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";

import { api, configureApiClient, unwrap } from "./client";
import { ApiError } from "./problem";

const ORIGIN = "http://localhost:3000";
const TOKEN = "csrf-token-da-sessao";

type Captured = { method: string; url: string; csrf: string | null; credentials: string };

function capture(method: "get" | "post" | "put" | "patch" | "delete", path: string, status = 200) {
  const seen: Captured[] = [];
  server.use(
    http[method](`${ORIGIN}${path}`, ({ request }) => {
      seen.push({
        method: request.method,
        url: request.url,
        csrf: request.headers.get("X-CSRF-Token"),
        credentials: request.credentials,
      });
      return status === 204 ? new HttpResponse(null, { status }) : HttpResponse.json({}, { status });
    }),
  );
  return seen;
}

const callbacks = {
  onUnauthenticated: vi.fn(),
  onPasswordChangeRequired: vi.fn(),
  onCsrfInvalid: vi.fn(),
};
let unregister: () => void;

beforeEach(() => {
  unregister = configureApiClient({ getCsrfToken: () => TOKEN, ...callbacks });
});

afterEach(() => {
  unregister();
  vi.clearAllMocks();
});

// FND-01 AC2: clientes openapi-fetch na origem do front, credentials same-origin.
describe("clientes por contrato", () => {
  it("chamam a API pela origem atual com credentials same-origin", async () => {
    const financeiro = capture("get", "/api/v1/financeiro/contas");
    const identity = capture("get", "/api/v1/auth/me");
    const estoque = capture("get", "/api/v1/estoque/produtos");

    await api.financeiro.GET("/api/v1/financeiro/contas");
    await api.identity.GET("/api/v1/auth/me");
    await api.estoque.GET("/api/v1/estoque/produtos");

    for (const [seen, path] of [
      [financeiro, "/api/v1/financeiro/contas"],
      [identity, "/api/v1/auth/me"],
      [estoque, "/api/v1/estoque/produtos"],
    ] as const) {
      expect(seen).toHaveLength(1);
      expect(seen[0].url).toBe(`${ORIGIN}${path}`);
      expect(seen[0].credentials).toBe("same-origin");
    }
  });
});

// FND-01 AC3.
describe("X-CSRF-Token", () => {
  it("vai em POST com o token da sessão atual", async () => {
    const seen = capture("post", "/api/v1/financeiro/contas", 201);
    await api.financeiro.POST("/api/v1/financeiro/contas", {
      body: { tipo: "RECEITA", nome: "Mensalidades" },
    });
    expect(seen[0].csrf).toBe(TOKEN);
  });

  it("vai em PUT com o token da sessão atual", async () => {
    const seen = capture("put", `/api/v1/users/${ids.otherUser}/roles`);
    await api.identity.PUT("/api/v1/users/{id}/roles", {
      params: { path: { id: ids.otherUser } },
      body: { roles: ["TESOURARIA"] },
    });
    expect(seen[0].csrf).toBe(TOKEN);
  });

  it("vai em PATCH com o token da sessão atual", async () => {
    const seen = capture("patch", `/api/v1/financeiro/contas/${ids.conta}`);
    await api.financeiro.PATCH("/api/v1/financeiro/contas/{id}", {
      params: { path: { id: ids.conta } },
      body: { nome: "Novo nome" },
    });
    expect(seen[0].csrf).toBe(TOKEN);
  });

  it("vai em DELETE com o token da sessão atual", async () => {
    // O contrato atual não tem DELETE; o método é exercitado pela chamada genérica.
    const seen = capture("delete", "/api/v1/financeiro/contas", 204);
    const request = api.financeiro.request as unknown as (
      method: string,
      url: string,
    ) => Promise<unknown>;
    await request("delete", "/api/v1/financeiro/contas");
    expect(seen[0].method).toBe("DELETE");
    expect(seen[0].csrf).toBe(TOKEN);
  });

  it("não vai em GET", async () => {
    const seen = capture("get", "/api/v1/financeiro/saldo");
    await api.financeiro.GET("/api/v1/financeiro/saldo");
    expect(seen[0].csrf).toBeNull();
  });

  it("segue sem o cabeçalho quando ainda não há token", async () => {
    unregister();
    unregister = configureApiClient({ getCsrfToken: () => null, ...callbacks });
    const seen = capture("post", "/api/v1/auth/logout", 204);
    await api.identity.POST("/api/v1/auth/logout");
    expect(seen[0].csrf).toBeNull();
  });
});

// FND-03 AC6-8 (lado do cliente) e FND-04 AC5.
describe("callbacks de 401 e 403", () => {
  function respond(method: "get" | "post", path: string, status: number, code: string) {
    server.use(http[method](`${ORIGIN}${path}`, () => problem(status, code)));
  }

  it("401 numa chamada autenticada chama onUnauthenticated", async () => {
    respond("get", "/api/v1/financeiro/contas", 401, "session_expired");
    await api.financeiro.GET("/api/v1/financeiro/contas");
    expect(callbacks.onUnauthenticated).toHaveBeenCalledTimes(1);
    expect(callbacks.onPasswordChangeRequired).not.toHaveBeenCalled();
    expect(callbacks.onCsrfInvalid).not.toHaveBeenCalled();
  });

  it("401 em POST /auth/login não chama onUnauthenticated", async () => {
    respond("post", "/api/v1/auth/login", 401, "invalid_credentials");
    const result = await api.identity.POST("/api/v1/auth/login", {
      body: { email: "pessoa.teste@example.com", password: "senha-errada" },
    });
    expect(result.response.status).toBe(401);
    expect(callbacks.onUnauthenticated).not.toHaveBeenCalled();
  });

  it("403 password_change_required chama onPasswordChangeRequired", async () => {
    respond("get", "/api/v1/financeiro/saldo", 403, "password_change_required");
    await api.financeiro.GET("/api/v1/financeiro/saldo");
    expect(callbacks.onPasswordChangeRequired).toHaveBeenCalledTimes(1);
    expect(callbacks.onUnauthenticated).not.toHaveBeenCalled();
    expect(callbacks.onCsrfInvalid).not.toHaveBeenCalled();
  });

  it("403 csrf_invalid chama onCsrfInvalid", async () => {
    respond("post", "/api/v1/financeiro/contas", 403, "csrf_invalid");
    await api.financeiro.POST("/api/v1/financeiro/contas", {
      body: { tipo: "DESPESA", nome: "Transporte" },
    });
    expect(callbacks.onCsrfInvalid).toHaveBeenCalledTimes(1);
    expect(callbacks.onUnauthenticated).not.toHaveBeenCalled();
    expect(callbacks.onPasswordChangeRequired).not.toHaveBeenCalled();
  });

  it("403 forbidden não chama nenhum callback", async () => {
    respond("post", "/api/v1/financeiro/contas", 403, "forbidden");
    await api.financeiro.POST("/api/v1/financeiro/contas", {
      body: { tipo: "DESPESA", nome: "Transporte" },
    });
    expect(callbacks.onUnauthenticated).not.toHaveBeenCalled();
    expect(callbacks.onPasswordChangeRequired).not.toHaveBeenCalled();
    expect(callbacks.onCsrfInvalid).not.toHaveBeenCalled();
  });

  it("callbacks desregistrados deixam de ser chamados", async () => {
    unregister();
    respond("get", "/api/v1/financeiro/contas", 401, "unauthenticated");
    await api.financeiro.GET("/api/v1/financeiro/contas");
    expect(callbacks.onUnauthenticated).not.toHaveBeenCalled();
  });
});

// FND-01 AC4: a tela recebe dado tipado ou ApiError, nunca exceção genérica.
describe("unwrap", () => {
  it("devolve o corpo de sucesso", async () => {
    server.use(
      http.get(`${ORIGIN}/api/v1/financeiro/contas`, () =>
        HttpResponse.json({ items: [conta()] }),
      ),
    );
    const data = await unwrap(api.financeiro.GET("/api/v1/financeiro/contas"));
    expect(data.items).toEqual([conta()]);
  });

  it("devolve o AuthContext do login", async () => {
    server.use(
      http.post(`${ORIGIN}/api/v1/auth/login`, () => HttpResponse.json(authContext())),
    );
    const data = await unwrap(
      api.identity.POST("/api/v1/auth/login", {
        body: { email: "pessoa.teste@example.com", password: "senha-de-teste" },
      }),
    );
    expect(data.csrf_token).toBe("csrf-token-de-teste");
  });

  it("lança ApiError com status, code e errors[] para problem+json", async () => {
    server.use(
      http.post(`${ORIGIN}/api/v1/financeiro/contas`, () =>
        HttpResponse.json(
          problemBody(422, "validation_failed", { errors: [{ field: "nome", code: "required" }] }),
          { status: 422, headers: { "Content-Type": "application/problem+json" } },
        ),
      ),
    );

    const error = await unwrap(
      api.financeiro.POST("/api/v1/financeiro/contas", { body: { tipo: "RECEITA", nome: "" } }),
    ).catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({
      status: 422,
      code: "validation_failed",
      errors: [{ field: "nome", code: "required" }],
    });
  });

  it("lança ApiError de indisponibilidade em falha de rede", async () => {
    server.use(http.get(`${ORIGIN}/api/v1/financeiro/saldo`, () => HttpResponse.error()));

    const error = await unwrap(api.financeiro.GET("/api/v1/financeiro/saldo")).catch(
      (e: unknown) => e,
    );

    expect(error).toBeInstanceOf(ApiError);
    expect((error as ApiError).code).toBe("network_error");
  });
});
