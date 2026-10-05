import { afterEach, expect, it, vi } from "vitest";

const redirect = vi.hoisted(() =>
  vi.fn((url: string) => {
    throw new Error(`NEXT_REDIRECT ${url}`);
  }),
);

vi.mock("next/navigation", () => ({ redirect }));

import Home from "./page";

afterEach(() => {
  redirect.mockClear();
  vi.restoreAllMocks();
});

// INT-01 AC3 e AC5: a raiz leva a /inicio e não consulta /healthz (nem nada da API).
it("redireciona para /inicio sem fazer nenhuma requisição", async () => {
  const fetchSpy = vi.spyOn(globalThis, "fetch");
  await expect(async () => {
    await Home();
  }).rejects.toThrow("NEXT_REDIRECT /inicio");
  expect(redirect).toHaveBeenCalledTimes(1);
  expect(redirect).toHaveBeenCalledWith("/inicio");
  expect(fetchSpy).not.toHaveBeenCalled();
});
