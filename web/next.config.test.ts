import { afterEach, describe, expect, it, vi } from "vitest";

import nextConfig, { apiRewrites } from "./next.config";

// FND-01 AC1 e design 01-fundacao (next.config.ts): /api/v1/* na origem do front
// é encaminhado para o mesmo caminho em API_URL; sem API_URL, a configuração falha
// com mensagem clara, nunca com destino vazio.
describe("rewrite same-origin de /api/v1", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("encaminha /api/v1/:path* para o mesmo caminho em API_URL", () => {
    expect(apiRewrites("http://localhost:8080")).toEqual([
      {
        source: "/api/v1/:path*",
        destination: "http://localhost:8080/api/v1/:path*",
      },
    ]);
  });

  it("não duplica a barra quando API_URL termina com /", () => {
    expect(apiRewrites("http://api.interna:8080/")).toEqual([
      {
        source: "/api/v1/:path*",
        destination: "http://api.interna:8080/api/v1/:path*",
      },
    ]);
  });

  it.each([undefined, "", "   "])(
    "falha com mensagem que cita API_URL quando ela é %j",
    (valor) => {
      expect(() => apiRewrites(valor)).toThrow(/API_URL/);
    },
  );

  it("a configuração do Next lê API_URL do ambiente", async () => {
    vi.stubEnv("API_URL", "http://127.0.0.1:9090");
    const rewrites = await nextConfig.rewrites?.();
    expect(rewrites).toEqual([
      {
        source: "/api/v1/:path*",
        destination: "http://127.0.0.1:9090/api/v1/:path*",
      },
    ]);
  });

  it("a configuração do Next falha sem API_URL no ambiente", async () => {
    vi.stubEnv("API_URL", "");
    await expect(async () => nextConfig.rewrites?.()).rejects.toThrow(
      /API_URL/,
    );
  });
});
