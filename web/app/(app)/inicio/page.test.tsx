import { screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { resetNavigation, setLocation } from "@/lib/session/navigation-mock";
import { mockMe, renderWithSession } from "@/lib/session/test-utils";
import { authContext } from "@/test/msw/fixtures";

import AppLayout from "../layout";
import InicioPage from "./page";

vi.mock("next/navigation", async () =>
  (await import("@/lib/session/navigation-mock")).nextNavigationMock,
);

beforeEach(() => {
  resetNavigation();
  setLocation("/inicio");
});

// Permissões efetivas reais de cada papel, como a API as concede (matriz
// agregada: api/internal/identity/app/roles_matrix.go, Contribution() de
// api/internal/financeiro/module.go e de api/internal/estoque/module.go).
const PERFIS = {
  TESOURARIA: [
    "financeiro:conta:create",
    "financeiro:conta:update",
    "financeiro:conta:deactivate",
    "financeiro:conta:read",
    "financeiro:lancamento:create",
    "financeiro:lancamento:update",
    "financeiro:lancamento:read",
    "financeiro:lancamento:receive",
    "financeiro:lancamento:pay",
    "financeiro:lancamento:cancel",
    "financeiro:saldo:read",
    "financeiro:comprovante:create",
    "financeiro:comprovante:read",
  ],
  ESTOQUE_LOJA: [
    "estoque:produto:create",
    "estoque:produto:read",
    "estoque:movimentacao:create",
    "estoque:movimentacao:adjust",
    "estoque:movimentacao:read",
    "estoque:saldo:read",
  ],
  CONSELHO_FISCAL: [
    "audit:log:read",
    "financeiro:prestacao_contas:read",
    "financeiro:prestacao_contas:approve",
    "financeiro:parecer:opine",
    "financeiro:conta:read",
    "financeiro:lancamento:read",
    "financeiro:saldo:read",
    "financeiro:comprovante:read",
    "estoque:produto:read",
    "estoque:movimentacao:read",
    "estoque:saldo:read",
  ],
  ADMIN_SISTEMA: [
    "identity:user:read",
    "identity:user:create",
    "identity:user:update",
    "identity:role:assign",
    "identity:admin:grant",
    "identity:admin:revoke",
    "identity:user:reset_password",
    "audit:log:read",
  ],
  ASSOCIADO: [],
} as const satisfies Record<string, readonly string[]>;

type Perfil = keyof typeof PERFIS;

type Link = { label: string; href: string | null };

function links(container: HTMLElement): Link[] {
  return within(container)
    .queryAllByRole("link")
    .map((a) => ({ label: a.textContent ?? "", href: a.getAttribute("href") }));
}

async function renderInicio(perfil: Perfil) {
  mockMe(authContext({ roles: [perfil], permissions: [...PERFIS[perfil]] }));
  renderWithSession(
    <AppLayout>
      <InicioPage />
    </AppLayout>,
  );
  await screen.findByRole("heading", { name: "Início" });
}

const navegacao = () => screen.getByRole("navigation", { name: "Navegação principal" });
const atalhos = () => screen.queryByRole("navigation", { name: "Atalhos das áreas" });

const INICIO: Link = { label: "Início", href: "/inicio" };
const LANCAMENTOS: Link = { label: "Lançamentos", href: "/financeiro/lancamentos" };
const CONTAS: Link = { label: "Contas", href: "/financeiro/contas" };
const SALDO: Link = { label: "Saldo", href: "/financeiro/saldo" };
const PRODUTOS: Link = { label: "Produtos", href: "/estoque/produtos" };
const USUARIOS: Link = { label: "Usuários", href: "/admin/usuarios" };

// INT-01 AC2 (navegação do shell) e AC4 (atalhos de /inicio), com as
// permissões reais de cada papel.
const esperado: Record<Perfil, { nav: Link[]; secoes: string[]; atalhos: Link[] }> = {
  TESOURARIA: {
    nav: [INICIO, LANCAMENTOS, CONTAS, SALDO],
    secoes: ["Financeiro"],
    atalhos: [LANCAMENTOS, CONTAS, SALDO],
  },
  ESTOQUE_LOJA: { nav: [INICIO, PRODUTOS], secoes: ["Estoque"], atalhos: [PRODUTOS] },
  CONSELHO_FISCAL: {
    nav: [INICIO, LANCAMENTOS, CONTAS, SALDO, PRODUTOS],
    secoes: ["Financeiro", "Estoque"],
    atalhos: [LANCAMENTOS, CONTAS, SALDO, PRODUTOS],
  },
  ADMIN_SISTEMA: { nav: [INICIO, USUARIOS], secoes: ["Administração"], atalhos: [USUARIOS] },
  ASSOCIADO: { nav: [INICIO], secoes: [], atalhos: [] },
};

describe.each(Object.keys(esperado) as Perfil[])("perfil %s", (perfil) => {
  it("a navegação do shell mostra exatamente os itens permitidos, na ordem", async () => {
    await renderInicio(perfil);
    expect(links(navegacao())).toEqual(esperado[perfil].nav);
  });

  it("/inicio mostra um atalho para cada área visível, agrupado pelo módulo", async () => {
    await renderInicio(perfil);
    const { atalhos: esperados, secoes } = esperado[perfil];
    if (esperados.length === 0) {
      expect(atalhos()).toBeNull();
      return;
    }
    const regiao = atalhos();
    expect(regiao).not.toBeNull();
    expect(links(regiao!)).toEqual(esperados);
    expect(
      within(regiao!)
        .getAllByRole("heading")
        .map((h) => h.textContent),
    ).toEqual(secoes);
  });
});

// INT-01 AC4: conta sem nenhuma área.
describe("conta sem acesso a módulos", () => {
  it("ASSOCIADO vê a mensagem explicando que a conta ainda não tem acesso a módulos", async () => {
    await renderInicio("ASSOCIADO");
    expect(screen.getByText("Sua conta ainda não tem acesso a nenhum módulo.")).toBeTruthy();
  });

  it("um perfil com alguma área não vê essa mensagem", async () => {
    await renderInicio("TESOURARIA");
    expect(screen.queryByText("Sua conta ainda não tem acesso a nenhum módulo.")).toBeNull();
  });
});
