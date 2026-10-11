import type { NavSection } from "@/components/app/nav";

// Itens de navegação do módulo financeiro, cada um com a permissão de leitura
// da sua área. A INT junta este arquivo à navegação global (nav-items.ts).

export const financeiroNav: NavSection = {
  label: "Financeiro",
  items: [
    { label: "Lançamentos", href: "/financeiro/lancamentos", permission: "financeiro:lancamento:read" },
    { label: "Contas", href: "/financeiro/contas", permission: "financeiro:conta:read" },
    { label: "Saldo", href: "/financeiro/saldo", permission: "financeiro:saldo:read" },
  ],
};
