import { ListTreeIcon, ReceiptIcon, WalletIcon } from "lucide-react";

import type { NavSection } from "@/components/app/nav";

// Itens de navegação do módulo financeiro, cada um com a permissão de leitura
// da sua área, o ícone do shell e a frase do atalho do início (DSS-01 AC3).
// A INT junta este arquivo à navegação global (nav-items.ts).

export const financeiroNav: NavSection = {
  label: "Financeiro",
  items: [
    {
      label: "Lançamentos",
      href: "/financeiro/lancamentos",
      permission: "financeiro:lancamento:read",
      icon: ReceiptIcon,
      description: "Receitas e despesas da torcida, com os comprovantes.",
    },
    {
      label: "Contas",
      href: "/financeiro/contas",
      permission: "financeiro:conta:read",
      icon: ListTreeIcon,
      description: "Plano de contas de receita e de despesa.",
    },
    {
      label: "Saldo",
      href: "/financeiro/saldo",
      permission: "financeiro:saldo:read",
      icon: WalletIcon,
      description: "Saldo atual: recebidos menos pagos.",
    },
  ],
};
