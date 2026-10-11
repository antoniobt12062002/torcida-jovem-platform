import type { NextConfig } from "next";

const API_PREFIX = "/api/v1";

// ADR-010 / WEB-D-001: o navegador fala só com a origem do front; o servidor
// Next encaminha /api/v1/* para a API Go em API_URL. Sem proxy.ts (WEB-D-009).
export function apiRewrites(apiUrl: string | undefined) {
  const base = apiUrl?.trim().replace(/\/+$/, "");
  if (!base) {
    throw new Error(
      "API_URL não está definida: o rewrite de /api/v1/* precisa do endereço da API " +
        "(por exemplo API_URL=http://localhost:8080).",
    );
  }
  return [
    {
      source: `${API_PREFIX}/:path*`,
      destination: `${base}${API_PREFIX}/:path*`,
    },
  ];
}

const nextConfig: NextConfig = {
  async rewrites() {
    return apiRewrites(process.env.API_URL);
  },
};

export default nextConfig;
