import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, networkError } from "@/lib/api/problem";
import { authContext } from "@/test/msw/fixtures";

import { AppProviders, shouldRetry } from "./app-providers";
import { resetNavigation } from "./navigation-mock";
import { useSession } from "./session-provider";
import { mockMe } from "./test-utils";

vi.mock("next/navigation", async () => (await import("./navigation-mock")).nextNavigationMock);

beforeEach(() => resetNavigation());

function Name() {
  const { status, context } = useSession();
  return <span data-testid="who">{status === "authenticated" ? context?.user.name : status}</span>;
}

describe("AppProviders", () => {
  it("fornece consultas e sessão à árvore", async () => {
    mockMe(authContext({ user: { id: "u1", email: "a@example.com", name: "Ana Teste" } }));
    render(
      <AppProviders>
        <Name />
      </AppProviders>,
    );
    await waitFor(() => expect(screen.getByTestId("who").textContent).toBe("Ana Teste"));
  });
});

describe("shouldRetry", () => {
  const http = (status: number) => new ApiError({ status, code: `http_${status}`, title: "x" });

  it("não repete respostas 4xx", () => {
    expect(shouldRetry(0, http(401))).toBe(false);
    expect(shouldRetry(0, http(403))).toBe(false);
    expect(shouldRetry(0, http(404))).toBe(false);
  });

  it("repete falha de rede e 5xx, no máximo duas vezes", () => {
    expect(shouldRetry(0, networkError())).toBe(true);
    expect(shouldRetry(1, http(503))).toBe(true);
    expect(shouldRetry(2, http(503))).toBe(false);
  });
});
