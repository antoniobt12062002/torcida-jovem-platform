import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";

import { authContext, problem } from "./fixtures";
import { server } from "./server";

// Teste de fumaça do harness (tasks 01/T1): o MSW está ligado no setup do
// Vitest, recusa requisição sem handler e descarta os handlers de cada teste.
describe("harness MSW", () => {
  it("responde pela origem do jsdom com o handler registrado no teste", async () => {
    server.use(
      http.get("http://localhost:3000/api/v1/auth/me", () =>
        HttpResponse.json(authContext({ permissions: ["financeiro:conta:read"] })),
      ),
    );

    const response = await fetch(`${window.location.origin}/api/v1/auth/me`);

    expect(window.location.origin).toBe("http://localhost:3000");
    expect(response.status).toBe(200);
    const body = await response.json();
    expect(body.permissions).toEqual(["financeiro:conta:read"]);
  });

  it("descarta os handlers do teste anterior e recusa requisição sem handler", async () => {
    await expect(
      fetch("http://localhost:3000/api/v1/auth/me"),
    ).rejects.toThrow();
  });

  it("serve problem+json com o content-type da API", async () => {
    server.use(
      http.get("http://localhost:3000/api/v1/financeiro/saldo", () =>
        problem(503, "service_unavailable"),
      ),
    );

    const response = await fetch("http://localhost:3000/api/v1/financeiro/saldo");

    expect(response.status).toBe(503);
    expect(response.headers.get("Content-Type")).toBe("application/problem+json");
    const body = await response.json();
    expect(body).toMatchObject({ status: 503, code: "service_unavailable" });
  });
});
