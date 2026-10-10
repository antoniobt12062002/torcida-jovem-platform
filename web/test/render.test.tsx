import { useQuery } from "@tanstack/react-query";
import { screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { renderWithProviders } from "./render";

function Probe({ fail }: { fail?: boolean }) {
  const query = useQuery({
    queryKey: ["probe"],
    queryFn: async () => {
      if (fail) throw new Error("falhou");
      return "ok";
    },
  });
  return (
    <div>
      <span data-testid="status">{query.status}</span>
      <span data-testid="failures">{query.failureCount}</span>
    </div>
  );
}

// tasks 01/T1: renderWithProviders monta um QueryClient novo por teste, sem
// repetição de consultas.
describe("renderWithProviders", () => {
  it("fornece um QueryClient para o componente", async () => {
    renderWithProviders(<Probe />);
    await waitFor(() => expect(screen.getByTestId("status").textContent).toBe("success"));
  });

  it("não repete consulta que falhou", async () => {
    renderWithProviders(<Probe fail />);
    await waitFor(() => expect(screen.getByTestId("status").textContent).toBe("error"));
    expect(screen.getByTestId("failures").textContent).toBe("1");
  });

  it("não compartilha cache entre renderizações", () => {
    const first = renderWithProviders(<Probe />);
    const second = renderWithProviders(<Probe />);
    expect(first.queryClient).not.toBe(second.queryClient);
  });
});
