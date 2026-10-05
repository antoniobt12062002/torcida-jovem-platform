import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext, problem, user } from "@/test/msw/fixtures";

import { cellsOf, mockListUsers } from "./test-helpers";
import { UsersPage } from "./users-page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => resetNavigation());

const READER = authContext({ permissions: ["identity:user:read"] });

const ana = user({
  id: "00000000-0000-4000-8000-0000000000a1",
  name: "Ana Tesoureira",
  email: "ana@example.com",
  active: true,
  must_change_password: true,
  roles: ["ASSOCIADO", "TESOURARIA"],
  admin_membership: { reason: "Assumiu a tesouraria", granted_at: "2026-10-01T12:00:00Z" },
});
const bruno = user({
  id: "00000000-0000-4000-8000-0000000000b2",
  name: "Bruno Associado",
  email: "bruno@example.com",
  active: false,
  must_change_password: false,
  roles: ["ASSOCIADO"],
  admin_membership: null,
});
const carla = user({
  id: "00000000-0000-4000-8000-0000000000c3",
  name: "Carla Estoque",
  email: "carla@example.com",
  roles: ["ASSOCIADO", "ESTOQUE_LOJA"],
});

async function renderPage() {
  mockMe(READER);
  renderWithSession(<UsersPage />);
  await screen.findByRole("heading", { name: "Usuários" });
}

// USR-01 AC1.
describe("listagem", () => {
  it("chama GET /users e mostra nome, e-mail, situação, papéis, vínculo e troca pendente", async () => {
    const requests = mockListUsers(() => ({ items: [ana, bruno], next_cursor: null }));
    await renderPage();
    await screen.findByText("ana@example.com");

    expect(cellsOf("ana@example.com")).toEqual([
      "Ana Tesoureira",
      "ana@example.com",
      "Ativo",
      "Associado, Tesouraria",
      "Sim",
      "Pendente",
    ]);
    expect(cellsOf("bruno@example.com")).toEqual([
      "Bruno Associado",
      "bruno@example.com",
      "Inativo",
      "Associado",
      "Não",
      "Não",
    ]);
    // Primeira página, sem filtros nem cursor.
    expect(requests).toEqual([{ limit: "50" }]);
  });

  it("sem usuários, mostra o estado vazio", async () => {
    mockListUsers(() => ({ items: [], next_cursor: null }));
    await renderPage();
    expect(await screen.findByText("Nenhum usuário encontrado.")).toBeTruthy();
  });

  it("falha na consulta mostra o erro e permite tentar de novo", async () => {
    let fail = true;
    mockListUsers(() => (fail ? problem(503, "service_unavailable") : { items: [ana], next_cursor: null }));
    await renderPage();
    expect(
      await screen.findByText("O serviço está indisponível no momento. Tente de novo em instantes."),
    ).toBeTruthy();
    fail = false;
    fireEvent.click(screen.getByRole("button", { name: "Tentar de novo" }));
    expect(await screen.findByText("ana@example.com")).toBeTruthy();
  });
});

// USR-01 AC2.
describe("filtros", () => {
  it("filtro de papel refaz a consulta com role e mostra só o resultado da API", async () => {
    const requests = mockListUsers((query) =>
      query.get("role") === "TESOURARIA"
        ? { items: [ana], next_cursor: null }
        : { items: [ana, bruno], next_cursor: null },
    );
    await renderPage();
    await screen.findByText("bruno@example.com");

    fireEvent.change(screen.getByLabelText("Papel"), { target: { value: "TESOURARIA" } });

    await waitFor(() => expect(requests).toHaveLength(2));
    expect(await screen.findByText("ana@example.com")).toBeTruthy();
    expect(screen.queryByText("bruno@example.com")).toBeNull();
    expect(requests).toEqual([{ limit: "50" }, { limit: "50", role: "TESOURARIA" }]);
  });

  it("filtro de situação envia active=false e active=true", async () => {
    const requests = mockListUsers((query) =>
      query.get("active") === "false"
        ? { items: [bruno], next_cursor: null }
        : { items: [ana, bruno], next_cursor: null },
    );
    await renderPage();
    await screen.findByText("ana@example.com");

    fireEvent.change(screen.getByLabelText("Situação"), { target: { value: "inativos" } });
    await waitFor(() => expect(requests).toHaveLength(2));
    expect(await screen.findByText("bruno@example.com")).toBeTruthy();
    expect(screen.queryByText("ana@example.com")).toBeNull();

    fireEvent.change(screen.getByLabelText("Situação"), { target: { value: "ativos" } });
    await waitFor(() => expect(requests).toHaveLength(3));
    expect(requests).toEqual([
      { limit: "50" },
      { limit: "50", active: "false" },
      { limit: "50", active: "true" },
    ]);
  });

  it("filtro aplicado depois de carregar mais recomeça do início, sem cursor", async () => {
    const requests = mockListUsers((query) => {
      if (query.get("role") === "ESTOQUE_LOJA") return { items: [carla], next_cursor: null };
      if (query.get("cursor") === "pagina-2") return { items: [bruno], next_cursor: null };
      return { items: [ana], next_cursor: "pagina-2" };
    });
    await renderPage();
    fireEvent.click(await screen.findByRole("button", { name: "Carregar mais" }));
    await screen.findByText("bruno@example.com");

    fireEvent.change(screen.getByLabelText("Papel"), { target: { value: "ESTOQUE_LOJA" } });

    await screen.findByText("carla@example.com");
    expect(screen.queryByText("ana@example.com")).toBeNull();
    expect(screen.queryByText("bruno@example.com")).toBeNull();
    expect(requests).toEqual([
      { limit: "50" },
      { limit: "50", cursor: "pagina-2" },
      { limit: "50", role: "ESTOQUE_LOJA" },
    ]);
  });
});

// USR-01 AC3.
describe("carregar mais", () => {
  it("com next_cursor, acrescenta a próxima página; na última, some", async () => {
    const requests = mockListUsers((query) =>
      query.get("cursor") === "pagina-2"
        ? { items: [bruno], next_cursor: null }
        : { items: [ana], next_cursor: "pagina-2" },
    );
    await renderPage();
    await screen.findByText("ana@example.com");

    fireEvent.click(screen.getByRole("button", { name: "Carregar mais" }));

    await screen.findByText("bruno@example.com");
    expect(screen.getByText("ana@example.com")).toBeTruthy();
    await waitFor(() => expect(screen.queryByRole("button", { name: "Carregar mais" })).toBeNull());
    expect(requests).toEqual([{ limit: "50" }, { limit: "50", cursor: "pagina-2" }]);
  });

  it("sem next_cursor, não oferece Carregar mais", async () => {
    mockListUsers(() => ({ items: [ana], next_cursor: null }));
    await renderPage();
    await screen.findByText("ana@example.com");
    expect(screen.queryByRole("button", { name: "Carregar mais" })).toBeNull();
  });
});
